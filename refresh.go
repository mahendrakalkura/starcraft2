package main

import (
	"context"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
)

func refresh() {
	files, err := buildFiles(paths)
	if err != nil {
		panic(err)
	}

	m := NewModel(len(files))

	p := tea.NewProgram(m)

	ctx := context.Background()

	go func() {
		wg := sync.WaitGroup{}
		for w := range settings.Workers {
			wg.Go(func() {
				worker(ctx, &m, w+1)
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
