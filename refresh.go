package main

import (
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

	go func() {
		wg := sync.WaitGroup{}
		for w := range 12 {
			wg.Add(1)
			go worker(&m, &wg, w+1)
		}
		wg.Wait()
	}()

	go func() {
		for _, file := range files {
			m.Channels.Input <- file
		}
	}()

	go func() {
		for {
			message := <-m.Channels.Output
			p.Send(message)
		}
	}()

	_, err = p.Run()
	if err != nil {
		panic(err)
	}
}
