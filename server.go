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
	if application.Settings.OpenRouterKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY environment variable is required for the serve action")
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", http.FileServer(http.FS(webFS)))
	mux.HandleFunc("GET /api/conversations", handleConversations(application))
	mux.HandleFunc("POST /api/ask", handleAsk(application))
	mux.HandleFunc("GET /api/conversation", handleConversation(application))
	mux.HandleFunc("POST /api/conversation/delete", handleDelete(application))

	server := &http.Server{Addr: ":" + application.Settings.Port, Handler: mux}

	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()

	fmt.Printf("listening on :%s\n", application.Settings.Port)
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func handleConversations(application *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := application.Queries.ConversationsList(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
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
		writeJSON(w, http.StatusOK, items)
	}
}

func handleAsk(application *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ConversationID int64  `json:"conversationId"`
			Prompt         string `json:"prompt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		body.Prompt = strings.TrimSpace(body.Prompt)
		if body.Prompt == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "prompt is required"})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), requestDeadline)
		defer cancel()

		conversationID := body.ConversationID
		title := promptTitle(body.Prompt)
		if conversationID == 0 {
			conversation, err := application.Queries.ConversationInsertOne(ctx, title)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
			conversationID = conversation.ID
		} else {
			conversation, err := application.Queries.ConversationGet(ctx, conversationID)
			if err != nil {
				writeJSON(w, http.StatusNotFound, map[string]string{"error": "conversation not found"})
				return
			}
			title = conversation.Title
		}

		if err := persistTurn(ctx, application, conversationID, "user", body.Prompt, nil, ""); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}

		answer, err := runAgent(ctx, application, conversationID)
		if err != nil {
			answer = "**Error:** " + err.Error()
			_ = persistTurn(ctx, application, conversationID, "assistant", answer, nil, "")
		}
		_ = application.Queries.ConversationTouch(ctx, conversationID)

		writeJSON(w, http.StatusOK, map[string]any{
			"conversationId": conversationID,
			"title":          title,
			"html":           renderMarkdown(answer),
		})
	}
}

func handleConversation(application *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return
		}

		conversation, err := application.Queries.ConversationGet(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "conversation not found"})
			return
		}

		turns, err := application.Queries.TurnsByConversation(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
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

		writeJSON(w, http.StatusOK, map[string]any{
			"id":        conversation.ID,
			"title":     conversation.Title,
			"exchanges": exchanges,
		})
	}
}

func handleDelete(application *Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ID int64 `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
			return
		}
		if err := application.Queries.ConversationDelete(r.Context(), body.ID); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}

// runAgent runs the tool-calling loop for the latest question in a conversation,
// persisting each assistant and tool message, and returns the final markdown answer.
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
		if err := persistTurn(ctx, application, conversationID, "assistant", message.Content, message.ToolCalls, ""); err != nil {
			return "", err
		}
		messages = append(messages, message)

		if len(message.ToolCalls) == 0 {
			return message.Content, nil
		}

		for _, call := range message.ToolCalls {
			result := runSQL(ctx, application.Database, toolCallSQL(call))
			if err := persistTurn(ctx, application, conversationID, "tool", result, nil, call.ID); err != nil {
				return "", err
			}
			messages = append(messages, orMessage{Role: "tool", ToolCallID: call.ID, Content: result})
			queriesUsed++
		}
	}

	return "", fmt.Errorf("the model did not produce a final answer")
}

// runSQL executes one model-supplied query in a read-only transaction with a
// statement timeout and returns the rows as a JSON string, or the error text so
// the model can correct itself.
func runSQL(ctx context.Context, pool interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
}, query string) string {
	query = strings.TrimSpace(query)
	query = strings.TrimRight(query, ";")
	if query == "" {
		return "error: empty sql"
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{AccessMode: pgx.ReadOnly})
	if err != nil {
		return "error: " + err.Error()
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = "+strconv.Itoa(int(queryTimeout/time.Millisecond))); err != nil {
		return "error: " + err.Error()
	}

	// Let Postgres serialize each row to JSON so all column types render correctly.
	rows, err := tx.Query(ctx, "SELECT to_jsonb(t) FROM ("+query+") t")
	if err != nil {
		return "error: " + err.Error()
	}
	defer rows.Close()

	out := []json.RawMessage{}
	for rows.Next() {
		var row []byte
		if err := rows.Scan(&row); err != nil {
			return "error: " + err.Error()
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return "error: " + err.Error()
	}

	encoded, err := json.Marshal(out)
	if err != nil {
		return "error: " + err.Error()
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
	request.Header.Set("Authorization", "Bearer "+settings.OpenRouterKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return orResponse{}, fmt.Errorf("openrouter request: %w", err)
	}
	defer response.Body.Close()

	var decoded orResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
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
	var encoded []byte
	if len(toolCalls) > 0 {
		var err error
		encoded, err = json.Marshal(toolCalls)
		if err != nil {
			return fmt.Errorf("json.Marshal(): %w", err)
		}
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
		_ = json.Unmarshal(turn.ToolCalls, &message.ToolCalls)
	}
	if turn.ToolCallID.Valid {
		message.ToolCallID = turn.ToolCallID.String
	}
	return message
}

func toolCallSQL(call orToolCall) string {
	var arguments struct {
		SQL string `json:"sql"`
	}
	_ = json.Unmarshal([]byte(call.Function.Arguments), &arguments)
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
	var buffer bytes.Buffer
	if err := markdown.Convert([]byte(source), &buffer); err != nil {
		return "<p>" + html.EscapeString(source) + "</p>"
	}
	return buffer.String()
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func systemPrompt(settings *Settings) string {
	schema := schemaSQL
	if index := strings.Index(schema, "CREATE TABLE files"); index >= 0 {
		schema = schema[index:]
	}

	players := "none configured"
	if len(settings.Players) > 0 {
		players = strings.Join(settings.Players, ", ")
	}

	return fmt.Sprintf(`You answer questions about a StarCraft II 4v4 ladder replay database (PostgreSQL).

You have one tool, run_sql, that runs a single read-only SELECT and returns the rows as JSON. Always run at least one query to answer a question; never answer from memory or guess numbers. You may run up to %d queries per question.

Database schema (only these tables exist for you; query nothing else):

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
- Never speculate beyond the rows the query returned. If a query returns no rows, say so plainly.`, maxQueries, schema, players)
}
