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
	action := flag.String("action", "ingest", "Action: ingest, sample, or serve")
	file := flag.String("file", "", "Replay file path (required for sample action)")
	force := flag.Bool("force", false, "Reprocess all replay files from scratch (ingest action)")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := NewApplication(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer application.Close()

	switch *action {
	case "ingest":
		err = ingest(ctx, application, *force)
	case "sample":
		err = sample(ctx, *file)
	case "serve":
		err = serve(ctx, application)
	default:
		err = fmt.Errorf("unknown action %q", *action)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
