package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
)

func maps(ctx context.Context, application *Application, me string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s: per-map record (10+ games; recent = last 180 days; * = in current pool)\n\n", me)

	rows, err := application.Queries.MapsSummary(ctx, me)
	if err != nil {
		return fmt.Errorf("application.Queries.MapsSummary(): %w", err)
	}

	w := newTable()
	w.AppendHeader(table.Row{"Map", "Games", "Wins", "Losses", "Win %", "Recent games", "Recent win %", "Last played"})
	for _, row := range rows {
		name := row.Map
		if row.RecentGames > 0 {
			name = fmt.Sprintf("%s *", name)
		}
		w.AppendRow(table.Row{name, row.Games, row.Wins, row.Losses, percentage(row.Wins, row.Losses), row.RecentGames, percentage(row.RecentWins, row.RecentLosses), row.LastPlayed.Time.Format("2006-01-02")})
	}
	fmt.Println(w.Render())

	return nil
}
