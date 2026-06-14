package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
)

func partners(ctx context.Context, application *Application) error {
	if len(application.Settings.Players) == 0 {
		return fmt.Errorf("PLAYERS environment variable is required for the partners action")
	}

	fmt.Printf("lineups of %v\n\n", application.Settings.Players)

	games, err := application.Queries.PartnersPerGame(ctx, application.Settings.Players)
	if err != nil {
		return fmt.Errorf("application.Queries.PartnersPerGame(): %w", err)
	}

	lineups := map[string]*openingCounts{}
	for _, game := range games {
		if lineups[game.Lineup] == nil {
			lineups[game.Lineup] = &openingCounts{}
		}
		lineups[game.Lineup].add(game.Result)
	}

	w := newTable()
	w.AppendHeader(table.Row{"Lineup", "Games", "Wins", "Losses", "Win %"})
	for _, lineup := range sortedKeysByGames(lineups) {
		counts := lineups[lineup]
		w.AppendRow(table.Row{lineup, counts.games, counts.wins, counts.losses, percentage(counts.wins, counts.losses)})
	}
	fmt.Println(w.Render())

	return nil
}
