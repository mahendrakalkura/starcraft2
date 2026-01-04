package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	defer cleanup()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	action := flag.String("action", "", "Action: refresh or sample")

	flag.Parse()

	_ = ctx

	if *action == "refresh" {
		refresh()
	}

	if *action == "sample" {
		sample()
	}

	if *action == "statistics" {
		statistics()
	}
}
