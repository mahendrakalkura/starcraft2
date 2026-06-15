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
	err := run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	action := flag.String("action", "ingest", "Action: ingest, sample, or serve")
	file := flag.String("file", "", "Replay file path (required for sample action)")
	force := flag.Bool("force", false, "Reprocess all replay files from scratch (ingest action)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := NewApplication(ctx)
	if err != nil {
		return err
	}
	defer application.Close()

	switch *action {
	case "ingest":
		return ingest(ctx, application, *force)
	case "sample":
		return sample(ctx, *file)
	case "serve":
		return serve(ctx, application)
	default:
		return fmt.Errorf("unknown action %q", *action)
	}
}
