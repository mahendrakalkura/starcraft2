package main

import (
	"context"
	"fmt"
	"sort"

	"main/models"

	"github.com/jedib0t/go-pretty/v6/table"
)

// ticksPerMinute is one in-game minute: 16 game loops per in-game second.
const ticksPerMinute = 960

type openingCounts struct {
	games  int64
	wins   int64
	losses int64
}

func (c *openingCounts) add(result string) {
	c.games++
	if result == "Win" {
		c.wins++
	}
	if result == "Loss" {
		c.losses++
	}
}

func openings(ctx context.Context, application *Application, me string, first string, second string, window int, race string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	opp := models.OpeningsPerGameParams{
		Name:         me,
		RaceAssigned: race,
		Name_2:       first,
		Name_3:       second,
	}
	games, err := application.Queries.OpeningsPerGame(ctx, opp)
	if err != nil {
		return fmt.Errorf("application.Queries.OpeningsPerGame(): %w", err)
	}

	windowTicks := int32(window * ticksPerMinute)

	fmt.Printf("%s as %s: %s vs %s opening (first started within %d min)\n\n", me, race, first, second, window)

	overall := map[string]*openingCounts{}
	byYear := map[int32]*openingCounts{}
	byTiming := map[string]*openingCounts{}
	byMap := map[string]*openingCounts{}

	for _, game := range games {
		opening := "neither"
		if game.FirstA <= windowTicks && game.FirstA < game.FirstB {
			opening = first
		} else if game.FirstB <= windowTicks && game.FirstB < game.FirstA {
			opening = second
		}

		if overall[opening] == nil {
			overall[opening] = &openingCounts{}
		}
		overall[opening].add(game.Result)

		if opening != first {
			continue
		}

		if byYear[game.Year] == nil {
			byYear[game.Year] = &openingCounts{}
		}
		byYear[game.Year].add(game.Result)

		timing := "4:00+"
		switch {
		case game.FirstA < 2*ticksPerMinute:
			timing = "< 2:00"
		case game.FirstA < 3*ticksPerMinute:
			timing = "2:00 - 3:00"
		case game.FirstA < 4*ticksPerMinute:
			timing = "3:00 - 4:00"
		}
		if byTiming[timing] == nil {
			byTiming[timing] = &openingCounts{}
		}
		byTiming[timing].add(game.Result)

		if byMap[game.Map] == nil {
			byMap[game.Map] = &openingCounts{}
		}
		byMap[game.Map].add(game.Result)
	}

	w := newTable()
	w.AppendHeader(table.Row{"Opening", "Games", "Wins", "Losses", "Win %"})
	for _, opening := range []string{first, second, "neither"} {
		counts := overall[opening]
		if counts == nil {
			continue
		}
		w.AppendRow(table.Row{opening, counts.games, counts.wins, counts.losses, percentage(counts.wins, counts.losses)})
	}
	fmt.Println(w.Render())

	years := []int32{}
	for year := range byYear {
		years = append(years, year)
	}
	sort.Slice(years, func(a int, z int) bool { return years[a] < years[z] })

	w = newTable()
	w.AppendHeader(table.Row{fmt.Sprintf("%s by year", first), "Games", "Wins", "Losses", "Win %"})
	for _, year := range years {
		counts := byYear[year]
		w.AppendRow(table.Row{year, counts.games, counts.wins, counts.losses, percentage(counts.wins, counts.losses)})
	}
	fmt.Println(w.Render())

	w = newTable()
	w.AppendHeader(table.Row{fmt.Sprintf("first %s at", first), "Games", "Wins", "Losses", "Win %"})
	for _, timing := range []string{"< 2:00", "2:00 - 3:00", "3:00 - 4:00", "4:00+"} {
		counts := byTiming[timing]
		if counts == nil {
			continue
		}
		w.AppendRow(table.Row{timing, counts.games, counts.wins, counts.losses, percentage(counts.wins, counts.losses)})
	}
	fmt.Println(w.Render())

	maps := []string{}
	for name, counts := range byMap {
		if counts.games >= 15 {
			maps = append(maps, name)
		}
	}
	sort.Slice(maps, func(a int, z int) bool {
		ca := byMap[maps[a]]
		cz := byMap[maps[z]]
		return float64(ca.wins)*float64(cz.wins+cz.losses) > float64(cz.wins)*float64(ca.wins+ca.losses)
	})

	w = newTable()
	w.AppendHeader(table.Row{fmt.Sprintf("%s by map (15+ games)", first), "Games", "Wins", "Losses", "Win %"})
	for _, name := range maps {
		counts := byMap[name]
		w.AppendRow(table.Row{name, counts.games, counts.wins, counts.losses, percentage(counts.wins, counts.losses)})
	}
	fmt.Println(w.Render())

	return nil
}
