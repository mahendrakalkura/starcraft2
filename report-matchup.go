package main

import (
	"context"
	"fmt"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
)

func matchup(ctx context.Context, application *Application, me string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s: matchups\n\n", me)

	games, err := application.Queries.MatchupPerGame(ctx, me)
	if err != nil {
		return fmt.Errorf("application.Queries.MatchupPerGame(): %w", err)
	}

	byRaceMode := map[string]*openingCounts{}
	byOpposition := map[string]*openingCounts{}

	for _, game := range games {
		raceMode := fmt.Sprintf("%s|%s", game.MyRace, game.Mode)
		if byRaceMode[raceMode] == nil {
			byRaceMode[raceMode] = &openingCounts{}
		}
		byRaceMode[raceMode].add(game.Result)

		opposition := fmt.Sprintf("%s|%s", game.MyRace, game.Opposition)
		if byOpposition[opposition] == nil {
			byOpposition[opposition] = &openingCounts{}
		}
		byOpposition[opposition].add(game.Result)
	}

	raceModes := sortedKeysByGames(byRaceMode)
	w := newTable()
	w.AppendHeader(table.Row{"My race", "Mode", "Games", "Wins", "Losses", "Win %"})
	for _, key := range raceModes {
		counts := byRaceMode[key]
		race, mode := splitKey(key)
		w.AppendRow(table.Row{race, mode, counts.games, counts.wins, counts.losses, percentage(counts.wins, counts.losses)})
	}
	fmt.Println(w.Render())

	oppositions := []string{}
	for key, counts := range byOpposition {
		if counts.games >= 10 {
			oppositions = append(oppositions, key)
		}
	}
	sort.Slice(oppositions, func(a int, z int) bool {
		return byOpposition[oppositions[a]].games > byOpposition[oppositions[z]].games
	})

	w = newTable()
	w.AppendHeader(table.Row{"My race", "Opposition", "Games", "Wins", "Losses", "Win %"})
	for _, key := range oppositions {
		counts := byOpposition[key]
		race, opposition := splitKey(key)
		w.AppendRow(table.Row{race, opposition, counts.games, counts.wins, counts.losses, percentage(counts.wins, counts.losses)})
	}
	fmt.Println(w.Render())

	return nil
}

func sortedKeysByGames(counts map[string]*openingCounts) []string {
	keys := []string{}
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(a int, z int) bool {
		return counts[keys[a]].games > counts[keys[z]].games
	})
	return keys
}

func splitKey(key string) (string, string) {
	for position := range key {
		if key[position] == '|' {
			return key[:position], key[position+1:]
		}
	}
	return key, ""
}
