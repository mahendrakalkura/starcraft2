package main

import (
	"context"
	"fmt"

	"main/models"

	"github.com/jedib0t/go-pretty/v6/table"
)

func impact(ctx context.Context, application *Application, name string, with string) error {
	if name == "" {
		return fmt.Errorf("--name is required for the impact action")
	}
	if len(application.Settings.Players) == 0 {
		return fmt.Errorf("PLAYERS environment variable is required for the impact action")
	}

	fmt.Printf("impact of %s relative to %v\n\n", name, application.Settings.Players)

	isp := models.ImpactSingleParams{
		Column1: application.Settings.Players,
		Name:    name,
	}
	sides, err := application.Queries.ImpactSingle(ctx, isp)
	if err != nil {
		return fmt.Errorf("application.Queries.ImpactSingle(): %w", err)
	}

	w := newTable()
	w.AppendHeader(table.Row{"Side", "Games", "Wins", "Losses", "Win %"})
	for _, side := range sides {
		w.AppendRow(table.Row{side.Side, side.Games, side.Wins, side.Losses, percentage(side.Wins, side.Losses)})
	}
	fmt.Println(w.Render())

	if with == "" {
		return nil
	}

	fmt.Printf("\npaired with %s\n\n", with)

	ipp := models.ImpactPairParams{
		Column1: application.Settings.Players,
		Name:    name,
		Name_2:  with,
	}
	pairs, err := application.Queries.ImpactPair(ctx, ipp)
	if err != nil {
		return fmt.Errorf("application.Queries.ImpactPair(): %w", err)
	}

	w = newTable()
	w.AppendHeader(table.Row{"Side", "Games", "Wins", "Losses", "Win %"})
	for _, pair := range pairs {
		w.AppendRow(table.Row{pair.Side, pair.Games, pair.Wins, pair.Losses, percentage(pair.Wins, pair.Losses)})
	}
	fmt.Println(w.Render())

	return nil
}
