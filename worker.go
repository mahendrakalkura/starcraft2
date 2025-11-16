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
		checkErr(err)

		log.SetOutput(io.Discard)

		r, err := rep.NewFromFile(file)
		checkErr(err)
		r.Close()

		log.SetOutput(os.Stderr)

		upsert(buildGame(file, r))

		m.Channels.Output <- Channel{File: "", Worker: number}
	}
}
