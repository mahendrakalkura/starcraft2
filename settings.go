package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Settings struct {
	Database        string
	OpenRouterKey   string
	OpenRouterModel string
	Players         []string
	Port            string
	Replays         []string
	Workers         int
}

func NewSettings() (*Settings, error) {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("godotenv.Load(): %w", err)
	}

	database := os.Getenv("DATABASE")
	if database == "" {
		return nil, fmt.Errorf("DATABASE environment variable is required")
	}

	// REPLAYS is required only for ingest, which checks it; serve does not need it.
	replays := []string{}
	s := os.Getenv("REPLAYS")
	if s != "" {
		replays = strings.Split(s, ",")
		for r := range replays {
			replays[r] = strings.Trim(strings.TrimSpace(replays[r]), `"`)
		}
	}

	openRouterModel := os.Getenv("OPENROUTER_MODEL")
	if openRouterModel == "" {
		openRouterModel = "deepseek/deepseek-chat"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	workers := runtime.NumCPU()
	s = os.Getenv("WORKERS")
	if s != "" {
		workers, err = strconv.Atoi(s)
		if err != nil {
			return nil, fmt.Errorf("WORKERS must be a number: %w", err)
		}
		if workers < 1 {
			return nil, fmt.Errorf("WORKERS must be >= 1, got %d", workers)
		}
	}

	players := []string{}
	s = os.Getenv("PLAYERS")
	if s != "" {
		players = strings.Split(s, ",")
		for p := range players {
			players[p] = strings.TrimSpace(players[p])
		}
	}

	settings := &Settings{
		Database:        database,
		OpenRouterKey:   os.Getenv("OPENROUTER_API_KEY"),
		OpenRouterModel: openRouterModel,
		Players:         players,
		Port:            port,
		Replays:         replays,
		Workers:         workers,
	}

	return settings, nil
}
