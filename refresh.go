package main

import (
	"context"
	"io"
	"log"
	"os"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

func refresh(ctx context.Context, application *Application) {
	log.SetOutput(io.Discard)
	defer log.SetOutput(os.Stderr)

	files, err := buildFiles(application.Settings.Replays)
	if err != nil {
		panic(err)
	}

	m := NewModel(len(files))

	p := tea.NewProgram(m)

	go func() {
		wg := sync.WaitGroup{}
		for w := range application.Settings.Workers {
			wg.Go(func() {
				worker(ctx, application, &m, w+1)
			})
		}
		wg.Wait()
		close(m.Channels.Input)
		close(m.Channels.Output)
	}()

	go func() {
		for _, file := range files {
			m.Channels.Input <- file
		}
	}()

	go func() {
		for message := range m.Channels.Output {
			p.Send(message)
		}
	}()

	_, err = p.Run()
	if err != nil {
		panic(err)
	}
}
