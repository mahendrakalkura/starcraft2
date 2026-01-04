package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	application, err := NewApplication()
	if err != nil {
		panic(err)
	}
	defer application.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	_ = ctx

	action := flag.String("action", "", "Action: refresh or sample")
	flag.Parse()

	if *action == "refresh" {
		refresh(application)
	}

	if *action == "sample" {
		sample(application)
	}

	if *action == "statistics" {
		statistics(application)
	}
}
