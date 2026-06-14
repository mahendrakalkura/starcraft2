package main

import (
	"context"
	"fmt"

	"main/models"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func versus(ctx context.Context, application *Application, me string, name string) error {
	if name == "" {
		return fmt.Errorf("--name is required for the versus action")
	}

	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s versus %s\n\n", me, name)

	vrp := models.VersusRelationParams{Name: me, Name_2: name}
	relations, err := application.Queries.VersusRelation(ctx, vrp)
	if err != nil {
		return fmt.Errorf("application.Queries.VersusRelation(): %w", err)
	}

	w := newTable()
	w.AppendHeader(table.Row{"Relation", "Games", "Wins", "Losses", "Win %"})
	for _, relation := range relations {
		w.AppendRow(table.Row{relation.Relation, relation.Games, relation.Wins, relation.Losses, percentage(relation.Wins, relation.Losses)})
	}
	fmt.Println(w.Render())

	vyp := models.VersusByYearParams{Name: me, Name_2: name}
	years, err := application.Queries.VersusByYear(ctx, vyp)
	if err != nil {
		return fmt.Errorf("application.Queries.VersusByYear(): %w", err)
	}

	w = newTable()
	w.AppendHeader(table.Row{"Year", "Games", "Wins", "Losses", "Win %"})
	for _, year := range years {
		w.AppendRow(table.Row{year.Year, year.Games, year.Wins, year.Losses, percentage(year.Wins, year.Losses)})
	}
	fmt.Println(w.Render())

	vbp := models.VersusByRaceParams{Name: me, Name_2: name}
	races, err := application.Queries.VersusByRace(ctx, vbp)
	if err != nil {
		return fmt.Errorf("application.Queries.VersusByRace(): %w", err)
	}

	w = newTable()
	w.AppendHeader(table.Row{"Race", "Mode", "Games", "Wins", "Losses", "Win %", "My MMR", "Their MMR", "My APM", "Their APM"})
	for _, race := range races {
		w.AppendRow(table.Row{race.Race, race.Mode, race.Games, race.Wins, race.Losses, percentage(race.Wins, race.Losses), race.MyMmr, race.TheirMmr, race.MyApm, race.TheirApm})
	}
	fmt.Println(w.Render())

	vmp := models.VersusByMapParams{Name: me, Name_2: name}
	maps, err := application.Queries.VersusByMap(ctx, vmp)
	if err != nil {
		return fmt.Errorf("application.Queries.VersusByMap(): %w", err)
	}

	w = newTable()
	w.AppendHeader(table.Row{"Map", "Games", "Wins", "Losses", "Win %"})
	for _, m := range maps {
		w.AppendRow(table.Row{m.Map, m.Games, m.Wins, m.Losses, percentage(m.Wins, m.Losses)})
	}
	fmt.Println(w.Render())

	vcp := models.VersusRecentParams{Name: me, Name_2: name}
	recents, err := application.Queries.VersusRecent(ctx, vcp)
	if err != nil {
		return fmt.Errorf("application.Queries.VersusRecent(): %w", err)
	}

	w = newTable()
	w.AppendHeader(table.Row{"Date", "Map", "Mode", "Result"})
	for _, recent := range recents {
		w.AppendRow(table.Row{recent.PlayedAt.Time.Format("2006-01-02"), recent.Map, recent.Mode, recent.Result})
	}
	fmt.Println(w.Render())

	return nil
}

func newTable() table.Writer {
	w := table.NewWriter()
	w.Style().Format.Header = text.FormatDefault
	return w
}
