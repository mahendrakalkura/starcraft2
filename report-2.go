package main

import (
	"context"
	"fmt"
	"strings"
)

func report2(ctx context.Context, application *Application) error {
	if len(application.Settings.Players) == 0 {
		return fmt.Errorf("PLAYERS environment variable is required for report-2")
	}

	results, err := application.Queries.Report2(ctx, application.Settings.Players)
	if err != nil {
		return fmt.Errorf("application.Queries.Report2(): %w", err)
	}

	for _, result := range results {
		date := result.GameDate.Time.Format("2006-01-02 (Mon)")

		color := "\033[31m" // red
		switch {
		case result.Wins > result.Losses:
			color = "\033[32m" // green
		case result.Wins == result.Losses:
			color = "\033[33m" // yellow
		}
		reset := "\033[0m"

		wins := strings.Repeat("X", int(result.Wins))
		losses := strings.Repeat("X", int(result.Losses))

		fmt.Printf("%s %s%s%s|%s%s%s\n", date, color, wins, reset, color, losses, reset)
	}

	return nil
}
