package main

import (
	"flag"
)

func main() {
	action := flag.String("action", "", "Action: refresh or sample")

	flag.Parse()

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
