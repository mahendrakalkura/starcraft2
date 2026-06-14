package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

func durationBinLabel(bin int32) string {
	if bin >= 8 {
		return "40+ min"
	}
	return fmt.Sprintf("%d - %d min", bin*5, bin*5+5)
}

func durations(ctx context.Context, application *Application, me string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s: game duration distribution\n\n", me)

	bins, err := application.Queries.DurationsBins(ctx, me)
	if err != nil {
		return fmt.Errorf("application.Queries.DurationsBins(): %w", err)
	}

	peak := int64(0)
	for _, bin := range bins {
		if bin.Games > peak {
			peak = bin.Games
		}
	}

	w := newTable()
	w.AppendHeader(table.Row{"Duration", "Games", "", "Wins", "Losses", "Win %"})
	for _, bin := range bins {
		bar := ""
		if peak > 0 {
			bar = strings.Repeat("#", int(bin.Games*40/peak))
		}
		w.AppendRow(table.Row{durationBinLabel(bin.Bin), bin.Games, bar, bin.Wins, bin.Losses, percentage(bin.Wins, bin.Losses)})
	}
	fmt.Println(w.Render())

	races, err := application.Queries.DurationsBinsByRace(ctx, me)
	if err != nil {
		return fmt.Errorf("application.Queries.DurationsBinsByRace(): %w", err)
	}

	race := ""
	w2 := table.Writer(nil)
	for _, bin := range races {
		if bin.Race != race {
			if w2 != nil {
				fmt.Println(w2.Render())
			}
			race = bin.Race
			fmt.Printf("\nas %s\n\n", race)
			w2 = newTable()
			w2.AppendHeader(table.Row{"Duration", "Games", "Wins", "Losses", "Win %"})
		}
		w2.AppendRow(table.Row{durationBinLabel(bin.Bin), bin.Games, bin.Wins, bin.Losses, percentage(bin.Wins, bin.Losses)})
	}
	if w2 != nil {
		fmt.Println(w2.Render())
	}

	return nil
}
