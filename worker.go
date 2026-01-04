package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/icza/s2prot/rep"
)

func worker(m *Model, wg *sync.WaitGroup, number int) {
	defer wg.Done()

	m.Channels.Output <- Channel{File: "", Worker: number}

	for file := range m.Channels.Input {
		m.Channels.Output <- Channel{File: file, Worker: number}

		err := mq.GamesDeleteOne(context.Background(), file)
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

		game, err := buildGame(file, r)
		if err != nil {
			err = fmt.Errorf("buildGame(): %w", err)
			m.Channels.Output <- Channel{File: file, Worker: number, Error: err.Error()}
			continue
		}

		err = upsert(game)
		if err != nil {
			err = fmt.Errorf("upsert(): %w", err)
			m.Channels.Output <- Channel{File: file, Worker: number, Error: err.Error()}
			continue
		}

		m.Channels.Output <- Channel{File: "", Worker: number}
	}
}
