package main

import (
	"context"
	"fmt"

	"main/models"

	"github.com/jedib0t/go-pretty/v6/table"
)

func economy(ctx context.Context, application *Application, me string, race string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s as %s: economy at 5/10/15 in-game minutes, wins vs losses\n\n", me, race)

	eap := models.EconomyAtParams{
		Name:         me,
		RaceAssigned: race,
	}
	rows, err := application.Queries.EconomyAt(ctx, eap)
	if err != nil {
		return fmt.Errorf("application.Queries.EconomyAt(): %w", err)
	}

	w := newTable()
	w.AppendHeader(table.Row{"Time", "Result", "Games", "Avg income", "Avg workers"})
	for _, row := range rows {
		minutes := row.Tick / ticksPerMinute
		w.AppendRow(table.Row{fmt.Sprintf("%d min", minutes), row.Result, row.Games, row.Income, row.Workers})
	}
	fmt.Println(w.Render())

	return nil
}
