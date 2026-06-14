package main

import (
	"context"
	"fmt"

	"main/models"

	"github.com/jedib0t/go-pretty/v6/table"
)

func maxout(ctx context.Context, application *Application, me string, unit string, race string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s as %s: maxed out (peak supply >= 195) with %s\n\n", me, race, unit)

	mbp := models.MaxoutBucketsParams{
		Name:         me,
		RaceAssigned: race,
		Name_2:       unit,
	}
	buckets, err := application.Queries.MaxoutBuckets(ctx, mbp)
	if err != nil {
		return fmt.Errorf("application.Queries.MaxoutBuckets(): %w", err)
	}

	w := newTable()
	w.AppendHeader(table.Row{"Scenario", "Games", "Wins", "Losses", "Win %"})
	for _, bucket := range buckets {
		w.AppendRow(table.Row{bucket.Scenario, bucket.Games, bucket.Wins, bucket.Losses, percentage(bucket.Wins, bucket.Losses)})
	}
	fmt.Println(w.Render())

	fmt.Printf("\nlosses while maxed with %s\n\n", unit)

	mop := models.MaxoutLossOpponentsParams{
		Name:         me,
		RaceAssigned: race,
		Name_2:       unit,
	}
	opponents, err := application.Queries.MaxoutLossOpponents(ctx, mop)
	if err != nil {
		return fmt.Errorf("application.Queries.MaxoutLossOpponents(): %w", err)
	}

	w = newTable()
	w.AppendHeader(table.Row{"Opponent", "Race", "Games", "Avg MMR"})
	for _, opponent := range opponents {
		w.AppendRow(table.Row{opponent.Name, opponent.Race, opponent.Games, opponent.AvgMmr})
	}
	fmt.Println(w.Render())

	mmp := models.MaxoutLossMapsParams{
		Name:         me,
		RaceAssigned: race,
		Name_2:       unit,
	}
	maps, err := application.Queries.MaxoutLossMaps(ctx, mmp)
	if err != nil {
		return fmt.Errorf("application.Queries.MaxoutLossMaps(): %w", err)
	}

	w = newTable()
	w.AppendHeader(table.Row{"Map", "Losses"})
	for _, m := range maps {
		w.AppendRow(table.Row{m.Map, m.Losses})
	}
	fmt.Println(w.Render())

	mcp := models.MaxoutLossCompositionParams{
		Name:         me,
		RaceAssigned: race,
		Name_2:       unit,
	}
	compositions, err := application.Queries.MaxoutLossComposition(ctx, mcp)
	if err != nil {
		return fmt.Errorf("application.Queries.MaxoutLossComposition(): %w", err)
	}

	w = newTable()
	w.AppendHeader(table.Row{"Opposing unit", "Built"})
	for _, composition := range compositions {
		w.AppendRow(table.Row{composition.Name, composition.Built})
	}
	fmt.Println(w.Render())

	return nil
}
