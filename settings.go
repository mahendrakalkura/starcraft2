package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type Settings struct {
	Database string
	Players  []string
	Replays  []string
	Workers  int
}

func NewSettings() *Settings {
	// Required: DATABASE
	s := os.Getenv("DATABASE")
	if s == "" {
		panic(fmt.Errorf("DATABASE environment variable is required"))
	}
	database := s

	// Required: REPLAYS
	s = os.Getenv("REPLAYS")
	if s == "" {
		panic(fmt.Errorf("REPLAYS environment variable is required"))
	}
	replays := strings.Split(s, ",")
	for r := range replays {
		replays[r] = strings.TrimSpace(replays[r])
	}

	// Optional: WORKERS (default: number of CPUs)
	workers := runtime.NumCPU()
	s = os.Getenv("WORKERS")
	if s != "" {
		w, err := strconv.Atoi(s)
		if err != nil {
			panic(fmt.Errorf("WORKERS must be a number: %w", err))
		}
		workers = w
	}

	// Optional: PLAYERS (default: empty = all players)
	players := []string{}
	s = os.Getenv("PLAYERS")
	if s != "" {
		players = strings.Split(s, ",")
		for p := range players {
			players[p] = strings.TrimSpace(players[p])
		}
	}

	return &Settings{
		Database: database,
		Players:  players,
		Replays:  replays,
		Workers:  workers,
	}
}
