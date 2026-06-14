package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
)

type mmrMonth struct {
	games  int64
	losses int64
	rated  int64
	sum    int64
	wins   int64
}

func mmrHistory(ctx context.Context, application *Application, me string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s: MMR history (rated games only; MMR 0 = unrated)\n\n", me)

	games, err := application.Queries.GamesSequence(ctx, me)
	if err != nil {
		return fmt.Errorf("application.Queries.GamesSequence(): %w", err)
	}

	months := map[string]*mmrMonth{}
	races := map[string]*mmrMonth{}
	for _, game := range games {
		month := game.PlayedAt.Time.Format("2006-01")
		if months[month] == nil {
			months[month] = &mmrMonth{}
		}
		if races[game.RaceAssigned] == nil {
			races[game.RaceAssigned] = &mmrMonth{}
		}
		for _, bucket := range []*mmrMonth{months[month], races[game.RaceAssigned]} {
			bucket.games++
			if game.Result == "Win" {
				bucket.wins++
			}
			if game.Result == "Loss" {
				bucket.losses++
			}
			if game.Mmr > 0 {
				bucket.rated++
				bucket.sum += int64(game.Mmr)
			}
		}
	}

	keys := []string{}
	for month := range months {
		keys = append(keys, month)
	}
	sort.Strings(keys)

	w := newTable()
	w.AppendHeader(table.Row{"Month", "Games", "Avg MMR", "Wins", "Losses", "Win %"})
	for _, month := range keys {
		bucket := months[month]
		average := int64(0)
		if bucket.rated > 0 {
			average = bucket.sum / bucket.rated
		}
		w.AppendRow(table.Row{month, bucket.games, average, bucket.wins, bucket.losses, percentage(bucket.wins, bucket.losses)})
	}
	fmt.Println(w.Render())

	names := []string{}
	for race := range races {
		names = append(names, race)
	}
	sort.Strings(names)

	w = newTable()
	w.AppendHeader(table.Row{"Race", "Games", "Avg MMR", "Wins", "Losses", "Win %"})
	for _, race := range names {
		bucket := races[race]
		average := int64(0)
		if bucket.rated > 0 {
			average = bucket.sum / bucket.rated
		}
		w.AppendRow(table.Row{race, bucket.games, average, bucket.wins, bucket.losses, percentage(bucket.wins, bucket.losses)})
	}
	fmt.Println(w.Render())

	return nil
}
