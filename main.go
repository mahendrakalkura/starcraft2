package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	action := flag.String("action", "", "Action: refresh, sample, or statistics")
	file := flag.String("file", "", "Replay file path (required for sample action)")
	flag.Parse()

	if *action == "sample" && *file == "" {
		fmt.Fprintln(os.Stderr, "Error: --file flag required with --action=sample")
		fmt.Fprintln(os.Stderr, "Usage: ./starcraft2 --action=sample --file=/path/to/replay.SC2Replay")
		os.Exit(1)
	}

	application, err := NewApplication()
	if err != nil {
		panic(err)
	}
	defer application.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if *action == "refresh" {
		refresh(ctx, application)
	}

	if *action == "sample" {
		sample(application, *file)
	}

	if *action == "statistics" {
		statistics(ctx, application)
	}
}
