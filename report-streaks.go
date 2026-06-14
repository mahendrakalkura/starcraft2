package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
)

func streaks(ctx context.Context, application *Application, me string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s: streaks and tilt\n\n", me)

	games, err := application.Queries.GamesSequence(ctx, me)
	if err != nil {
		return fmt.Errorf("application.Queries.GamesSequence(): %w", err)
	}

	baseline := &openingCounts{}
	afterTwoWins := &openingCounts{}
	afterTwoLosses := &openingCounts{}
	afterThreeLosses := &openingCounts{}

	longestWin := 0
	longestLoss := 0
	winEnd := ""
	lossEnd := ""
	run := 0
	last := ""

	for _, game := range games {
		if game.Result != "Win" && game.Result != "Loss" {
			continue
		}

		baseline.add(game.Result)
		if last == "Win" && run >= 2 {
			afterTwoWins.add(game.Result)
		}
		if last == "Loss" && run >= 2 {
			afterTwoLosses.add(game.Result)
		}
		if last == "Loss" && run >= 3 {
			afterThreeLosses.add(game.Result)
		}

		if game.Result == last {
			run++
		} else {
			run = 1
			last = game.Result
		}

		date := game.PlayedAt.Time.Format("2006-01-02")
		if last == "Win" && run > longestWin {
			longestWin = run
			winEnd = date
		}
		if last == "Loss" && run > longestLoss {
			longestLoss = run
			lossEnd = date
		}
	}

	w := newTable()
	w.AppendHeader(table.Row{"Situation", "Games", "Wins", "Losses", "Win %"})
	w.AppendRow(table.Row{"baseline", baseline.games, baseline.wins, baseline.losses, percentage(baseline.wins, baseline.losses)})
	w.AppendRow(table.Row{"after 2+ wins", afterTwoWins.games, afterTwoWins.wins, afterTwoWins.losses, percentage(afterTwoWins.wins, afterTwoWins.losses)})
	w.AppendRow(table.Row{"after 2+ losses", afterTwoLosses.games, afterTwoLosses.wins, afterTwoLosses.losses, percentage(afterTwoLosses.wins, afterTwoLosses.losses)})
	w.AppendRow(table.Row{"after 3+ losses", afterThreeLosses.games, afterThreeLosses.wins, afterThreeLosses.losses, percentage(afterThreeLosses.wins, afterThreeLosses.losses)})
	fmt.Println(w.Render())

	fmt.Printf("\nlongest win streak: %d (ended %s)\n", longestWin, winEnd)
	fmt.Printf("longest loss streak: %d (ended %s)\n", longestLoss, lossEnd)

	return nil
}
