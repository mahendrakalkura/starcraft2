package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Settings maps one-to-one to the variables in .env. The field names match the
// environment variable names, so there is only ever one name to reason about.
// Every variable is required: NewSettings reports any that are missing and the
// program exits. There are no defaults.
type Settings struct {
	GoEnvironment    string
	GoPlayers        []string
	GoPort           string
	GoReplays        []string
	GoWorkers        int
	OpenRouterAPIKey string
	OpenRouterModel  string
	PostgresDB       string
	PostgresHost     string
	PostgresPassword string
	PostgresPort     string
	PostgresUser     string
}

func NewSettings() (*Settings, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("godotenv.Load(): %w", err)
	}

	missing := []string{}
	get := func(key string) string {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			missing = append(missing, key)
		}
		return value
	}

	goEnvironment := get("GO_ENVIRONMENT")
	goPlayers := get("GO_PLAYERS")
	goPort := get("GO_PORT")
	goReplays := get("GO_REPLAYS")
	goWorkers := get("GO_WORKERS")
	openRouterAPIKey := get("OPENROUTER_API_KEY")
	openRouterModel := get("OPENROUTER_MODEL")
	postgresDB := get("POSTGRES_DB")
	postgresHost := get("POSTGRES_HOST")
	postgresPassword := get("POSTGRES_PASSWORD")
	postgresPort := get("POSTGRES_PORT")
	postgresUser := get("POSTGRES_USER")

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	workers, err := strconv.Atoi(goWorkers)
	if err != nil {
		return nil, fmt.Errorf("GO_WORKERS must be a number: %w", err)
	}
	if workers < 1 {
		return nil, fmt.Errorf("GO_WORKERS must be >= 1, got %d", workers)
	}

	return &Settings{
		GoEnvironment:    goEnvironment,
		GoPlayers:        splitList(goPlayers),
		GoPort:           goPort,
		GoReplays:        splitList(goReplays),
		GoWorkers:        workers,
		OpenRouterAPIKey: openRouterAPIKey,
		OpenRouterModel:  openRouterModel,
		PostgresDB:       postgresDB,
		PostgresHost:     postgresHost,
		PostgresPassword: postgresPassword,
		PostgresPort:     postgresPort,
		PostgresUser:     postgresUser,
	}, nil
}

// DatabaseURL builds the pgx connection string from the Postgres settings.
func (s *Settings) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		s.PostgresUser, s.PostgresPassword, s.PostgresHost, s.PostgresPort, s.PostgresDB,
	)
}

func splitList(value string) []string {
	parts := strings.Split(value, ",")
	for i := range parts {
		parts[i] = strings.Trim(strings.TrimSpace(parts[i]), `"`)
	}
	return parts
}
