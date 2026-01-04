package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/icza/s2prot/rep"
)

func worker(ctx context.Context, application *Application, m *Model, number int) {
	m.Channels.Output <- Channel{File: "", Worker: number}

	for file := range m.Channels.Input {
		m.Channels.Output <- Channel{File: file, Worker: number}

		err := application.Queries.GamesDeleteOne(ctx, file)
		if err != nil {
			m.Channels.Output <- Channel{File: file, Worker: number, Error: err.Error()}
		}

		log.SetOutput(io.Discard)

		r, err := rep.NewFromFile(file)
		if err != nil {
			err = fmt.Errorf("rep.NewFromFile(): %w", err)
			m.Channels.Output <- Channel{File: file, Worker: number, Error: err.Error()}
			continue
		}
		_ = r.Close()

		log.SetOutput(os.Stderr)

		game, err := buildGame(application.Settings, file, r)
		if err != nil {
			err = fmt.Errorf("buildGame(): %w", err)
			m.Channels.Output <- Channel{File: file, Worker: number, Error: err.Error()}
			continue
		}

		err = upsert(ctx, application.DB, application.Queries, game)
		if err != nil {
			err = fmt.Errorf("upsert(): %w", err)
			m.Channels.Output <- Channel{File: file, Worker: number, Error: err.Error()}
			continue
		}

		m.Channels.Output <- Channel{File: "", Worker: number}
	}
}
