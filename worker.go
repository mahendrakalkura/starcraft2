package main

import (
	"context"
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
		check(err)

		log.SetOutput(io.Discard)

		r, err := rep.NewFromFile(file)
		if err != nil {
			continue
		}
		_ = r.Close()

		log.SetOutput(os.Stderr)

		game, err := buildGame(file, r)
		check(err)

		err = upsert(game)
		check(err)

		m.Channels.Output <- Channel{File: "", Worker: number}
	}
}
