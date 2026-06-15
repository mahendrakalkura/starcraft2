package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"main/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

const (
	maxQueries      = 3
	queryTimeout    = 30 * time.Second
	requestDeadline = 120 * time.Second
	openRouterURL   = "https://openrouter.ai/api/v1/chat/completions"
	titleMaxLength  = 50
)

//go:embed index.css index.html index.js
var webFS embed.FS

//go:embed sqlc/schema.sql
var schemaSQL string

var markdown = goldmark.New(goldmark.WithExtensions(extension.GFM))

var runSQLTool = orTool{
	Type: "function",
	Function: orToolFunction{
		Name:        "run_sql",
		Description: "Run a single read-only SQL SELECT query against the PostgreSQL database and return the matching rows as JSON. Use this to answer every question; never answer from memory.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"sql": {
					"type": "string",
					"description": "A single read-only PostgreSQL SELECT statement (a WITH ... SELECT is fine). No semicolons, no DDL or DML."
				}
			},
			"required": ["sql"]
		}`),
	},
}

type orRequest struct {
	Model      string      `json:"model"`
	Messages   []orMessage `json:"messages"`
	Tools      []orTool    `json:"tools,omitempty"`
	ToolChoice string      `json:"tool_choice,omitempty"`
}

type orMessage struct {
	Role       string       `json:"role"`
	Content    string       `json:"content,omitempty"`
	ToolCalls  []orToolCall `json:"tool_calls,omitempty"`
	ToolCallID string       `json:"tool_call_id,omitempty"`
}

type orToolCall struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Function orFunctionCall `json:"function"`
}

type orFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type orTool struct {
	Type     string         `json:"type"`
	Function orToolFunction `json:"function"`
}

type orToolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type orResponse struct {
	Choices []struct {
		Message orMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func serve(ctx context.Context, application *Application) error {
	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.FS(webFS)))
	mux.HandleFunc("GET /api/conversations", handleConversations(application))
	mux.HandleFunc("POST /api/ask", handleAsk(application))
	mux.HandleFunc("GET /api/conversation", handleConversation(application))
	mux.HandleFunc("POST /api/conversation/delete", handleDelete(application))

	server := &http.Server{Addr: fmt.Sprintf(":%s", application.Settings.GoPort), Handler: mux}

	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := server.Shutdown(shutdown)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	fmt.Printf("listening on :%s\n", application.Settings.GoPort)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func handleConversations(application *Application) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		rows, err := application.Queries.ConversationList(request.Context())
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		type item struct {
			ID        int64  `json:"id"`
			Title     string `json:"title"`
			UpdatedAt string `json:"updatedAt"`
		}
		items := make([]item, 0, len(rows))
		for _, row := range rows {
			updated := ""
			if row.UpdatedAt.Valid {
				updated = row.UpdatedAt.Time.Format(time.RFC3339)
			}
			items = append(items, item{ID: row.ID, Title: row.Title, UpdatedAt: updated})
		}
		writeJSON(writer, http.StatusOK, items)
	}
}

func handleAsk(application *Application) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body := struct {
			ConversationID int64  `json:"conversationId"`
			Prompt         string `json:"prompt"`
		}{}
		err := json.NewDecoder(request.Body).Decode(&body)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		body.Prompt = strings.TrimSpace(body.Prompt)
		if body.Prompt == "" {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "prompt is required"})
			return
		}

		ctx, cancel := context.WithTimeout(request.Context(), requestDeadline)
		defer cancel()

		conversationID := body.ConversationID
		title := promptTitle(body.Prompt)
		if conversationID == 0 {
			conversation, e := application.Queries.ConversationInsertOne(ctx, title)
			if e != nil {
				writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": e.Error()})
				return
			}
			conversationID = conversation.ID
		} else {
			conversation, e := application.Queries.ConversationGet(ctx, conversationID)
			if e != nil {
				writeJSON(writer, http.StatusNotFound, map[string]string{"error": "conversation not found"})
				return
			}
			title = conversation.Title
		}

		err = persistTurn(ctx, application, conversationID, "user", body.Prompt, nil, "")
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		answer, err := runAgent(ctx, application, conversationID)
		if err != nil {
			answer = fmt.Sprintf("**Error:** %s", err.Error())
			e := persistTurn(ctx, application, conversationID, "assistant", answer, nil, "")
			if e != nil {
				fmt.Fprintln(os.Stderr, e)
			}
		}

		err = application.Queries.ConversationTouch(ctx, conversationID)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		writeJSON(writer, http.StatusOK, map[string]any{
			"conversationId": conversationID,
			"title":          title,
			"html":           renderMarkdown(answer),
		})
	}
}

func handleConversation(application *Application) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		id, err := strconv.ParseInt(request.URL.Query().Get("id"), 10, 64)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}

		conversation, err := application.Queries.ConversationGet(request.Context(), id)
		if err != nil {
			writeJSON(writer, http.StatusNotFound, map[string]string{"error": "conversation not found"})
			return
		}

		turns, err := application.Queries.TurnsByConversation(request.Context(), id)
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		type exchange struct {
			Prompt string `json:"prompt"`
			HTML   string `json:"html"`
		}
		exchanges := []exchange{}
		index := -1
		for _, turn := range turns {
			switch turn.Role {
			case "user":
				exchanges = append(exchanges, exchange{Prompt: turn.Content})
				index = len(exchanges) - 1
			case "assistant":
				if index >= 0 && turn.Content != "" {
					exchanges[index].HTML = renderMarkdown(turn.Content)
				}
			}
		}

		writeJSON(writer, http.StatusOK, map[string]any{
			"id":        conversation.ID,
			"title":     conversation.Title,
			"exchanges": exchanges,
		})
	}
}

func handleDelete(application *Application) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		body := struct {
			ID int64 `json:"id"`
		}{}
		err := json.NewDecoder(request.Body).Decode(&body)
		if err != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}

		err = application.Queries.ConversationDelete(request.Context(), body.ID)
		if err != nil {
			writeJSON(writer, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(writer, http.StatusOK, map[string]bool{"ok": true})
	}
}

func runAgent(ctx context.Context, application *Application, conversationID int64) (string, error) {
	turns, err := application.Queries.TurnsByConversation(ctx, conversationID)
	if err != nil {
		return "", fmt.Errorf("application.Queries.TurnsByConversation(): %w", err)
	}

	messages := []orMessage{{Role: "system", Content: systemPrompt(application.Settings)}}
	for _, turn := range turns {
		messages = append(messages, turnToMessage(turn))
	}

	queriesUsed := 0
	for iteration := 0; iteration < maxQueries+2; iteration++ {
		toolChoice := "auto"
		if queriesUsed == 0 {
			toolChoice = "required"
		}
		if queriesUsed >= maxQueries {
			toolChoice = "none"
		}

		response, err := callOpenRouter(ctx, application.Settings, messages, toolChoice)
		if err != nil {
			return "", err
		}
		if len(response.Choices) == 0 {
			return "", fmt.Errorf("openrouter returned no choices")
		}

		message := response.Choices[0].Message
		message.Role = "assistant"
		err = persistTurn(ctx, application, conversationID, "assistant", message.Content, message.ToolCalls, "")
		if err != nil {
			return "", err
		}
		messages = append(messages, message)

		if len(message.ToolCalls) == 0 {
			return message.Content, nil
		}

		for _, call := range message.ToolCalls {
			result := runSQL(ctx, application.Database, toolCallSQL(call))
			err = persistTurn(ctx, application, conversationID, "tool", result, nil, call.ID)
			if err != nil {
				return "", err
			}
			messages = append(messages, orMessage{Role: "tool", ToolCallID: call.ID, Content: result})
			queriesUsed++
		}
	}

	return "", fmt.Errorf("the model did not produce a final answer")
}

// runSQL returns the rows as a JSON string, or the error text so the model can self-correct.
func runSQL(ctx context.Context, pool interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}, query string,
) string {
	query = strings.TrimSpace(query)
	query = strings.TrimRight(query, ";")
	if query == "" {
		return "error: empty sql"
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL statement_timeout = %d", int(queryTimeout/time.Millisecond)))
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	// Postgres serializes each row to JSON so every column type renders correctly.
	rows, err := tx.Query(ctx, fmt.Sprintf("SELECT to_jsonb(t) FROM (%s) t", query))
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	defer rows.Close()

	out := []json.RawMessage{}
	for rows.Next() {
		row := []byte{}
		err = rows.Scan(&row)
		if err != nil {
			return fmt.Sprintf("error: %s", err.Error())
		}
		out = append(out, row)
	}
	err = rows.Err()
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}

	encoded, err := json.Marshal(out)
	if err != nil {
		return fmt.Sprintf("error: %s", err.Error())
	}
	return string(encoded)
}

func callOpenRouter(ctx context.Context, settings *Settings, messages []orMessage, toolChoice string) (orResponse, error) {
	payload, err := json.Marshal(orRequest{
		Model:      settings.OpenRouterModel,
		Messages:   messages,
		Tools:      []orTool{runSQLTool},
		ToolChoice: toolChoice,
	})
	if err != nil {
		return orResponse{}, fmt.Errorf("json.Marshal(): %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(payload))
	if err != nil {
		return orResponse{}, fmt.Errorf("http.NewRequestWithContext(): %w", err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", settings.OpenRouterAPIKey))
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return orResponse{}, fmt.Errorf("openrouter request: %w", err)
	}
	defer func() {
		err := response.Body.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	decoded := orResponse{}
	err = json.NewDecoder(response.Body).Decode(&decoded)
	if err != nil {
		return orResponse{}, fmt.Errorf("openrouter decode: %w", err)
	}
	if decoded.Error != nil {
		return orResponse{}, fmt.Errorf("openrouter: %s", decoded.Error.Message)
	}
	if response.StatusCode != http.StatusOK {
		return orResponse{}, fmt.Errorf("openrouter status %d", response.StatusCode)
	}
	return decoded, nil
}

func persistTurn(ctx context.Context, application *Application, conversationID int64, role string, content string, toolCalls []orToolCall, toolCallID string) error {
	encoded := []byte(nil)
	if len(toolCalls) > 0 {
		marshaled, err := json.Marshal(toolCalls)
		if err != nil {
			return fmt.Errorf("json.Marshal(): %w", err)
		}
		encoded = marshaled
	}

	id := pgtype.Text{}
	if toolCallID != "" {
		id = pgtype.Text{String: toolCallID, Valid: true}
	}

	err := application.Queries.TurnInsertOne(ctx, models.TurnInsertOneParams{
		ConversationID: conversationID,
		Role:           role,
		Content:        content,
		ToolCalls:      encoded,
		ToolCallID:     id,
	})
	if err != nil {
		return fmt.Errorf("application.Queries.TurnInsertOne(): %w", err)
	}
	return nil
}

func turnToMessage(turn models.Turn) orMessage {
	message := orMessage{Role: turn.Role, Content: turn.Content}
	if len(turn.ToolCalls) > 0 {
		err := json.Unmarshal(turn.ToolCalls, &message.ToolCalls)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
	if turn.ToolCallID.Valid {
		message.ToolCallID = turn.ToolCallID.String
	}
	return message
}

func toolCallSQL(call orToolCall) string {
	arguments := struct {
		SQL string `json:"sql"`
	}{}
	err := json.Unmarshal([]byte(call.Function.Arguments), &arguments)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return arguments.SQL
}

func promptTitle(prompt string) string {
	runes := []rune(strings.TrimSpace(prompt))
	if len(runes) > titleMaxLength {
		return string(runes[:titleMaxLength])
	}
	return string(runes)
}

func renderMarkdown(source string) string {
	buffer := bytes.Buffer{}
	err := markdown.Convert([]byte(source), &buffer)
	if err != nil {
		return fmt.Sprintf("<p>%s</p>", html.EscapeString(source))
	}
	return buffer.String()
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	err := json.NewEncoder(writer).Encode(value)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func systemPrompt(settings *Settings) string {
	players := "none configured"
	if len(settings.GoPlayers) > 0 {
		players = strings.Join(settings.GoPlayers, ", ")
	}

	return fmt.Sprintf(`You answer questions about a StarCraft II 4v4 ladder replay database (PostgreSQL).

You have one tool, run_sql, that runs a single read-only SELECT and returns the rows as JSON. Always run at least one query to answer a question; never answer from memory or guess numbers. You may run up to %d queries per question.

Database schema (the conversations and turns tables hold this chat app's own history, not game data; ignore them for game questions):

%s

Domain rules you must follow when writing SQL:
- A tick is the game loop: 16 ticks per in-game second, 960 per in-game minute. games.duration is in in-game seconds, so tick/16 lines up with duration. PlayerStats rows (the stats table) are sampled every 160 ticks.
- food_used >= 195 means the player reached maximum supply ("maxed").
- units.action: 'born' and 'init' are units that player created and owns; 'killed' is a kill credited to that player (the killer), not a death. Deaths without a killer are not stored.
- team is an integer column on players; there is no teams table. Players on the same team in a game share the same team number. result is one of 'Win', 'Loss', 'Tie', 'Undecided'.
- mode is derived from team sizes, e.g. '1v1', '2v2', '4v4'.
- race_assigned is the race actually played (Random resolves to a real race here); race_selected is what was picked. mmr = 0 means unrated.
- The matches view has one row per game with comma-separated winners and losers name lists.
- For relative dates use Postgres now() and current_date (e.g. played_at > now() - interval '30 days'); never hardcode today's date.
- The tracked players of interest are: %s.

Answer style:
- Lead with the answer, then brief supporting detail. Be concise.
- When a query returns multiple rows (per map, per year, top opponents, etc.), present them as a GitHub-flavored markdown table.
- Always state the sample size, e.g. "over N games".
- Never speculate beyond the rows the query returned. If a query returns no rows, say so plainly.`, maxQueries, schemaSQL, players)
}
