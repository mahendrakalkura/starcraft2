package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func report1(ctx context.Context, application *Application) error {
	matches, err := application.Queries.Report1(ctx)
	if err != nil {
		return fmt.Errorf("application.Queries.Report1(): %w", err)
	}

	w := table.NewWriter()
	w.AppendHeader(table.Row{"Played At", "Map", "Mode", "Duration", "Winners", "Losers"})
	w.Style().Format.Header = text.FormatDefault

	for _, match := range matches {
		playedAt := match.PlayedAt.Time.Format("2006-01-02 15:04:05")
		duration := fmt.Sprintf("%02d:%02d", match.Duration/60, match.Duration%60)
		w.AppendRow(table.Row{playedAt, match.Map, match.Mode, duration, string(match.Winners), string(match.Losers)})
	}

	fmt.Println(w.Render())

	return nil
}
