package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/icza/s2prot/rep"
)

func sample(application *Application, file string) {
	log.SetOutput(io.Discard)

	r, err := rep.NewFromFile(file)
	if err != nil {
		panic(fmt.Errorf("rep.NewFromFile(%q): %w", file, err))
	}
	_ = r.Close()

	log.SetOutput(os.Stderr)

	_ = os.WriteFile("r.json", []byte(dump(r)), 0o644)

	game, err := buildGame(application.Settings, file, r)
	if err != nil {
		panic(fmt.Errorf("buildGame(%q): %w", file, err))
	}

	_ = os.WriteFile("game.json", []byte(dump(game)), 0o644)
	fmt.Printf("Wrote r.json and game.json for %s\n", file)
}
