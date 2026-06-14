package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
)

func rivals(ctx context.Context, application *Application, me string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s: most-faced opponents (30+ games)\n\n", me)

	rows, err := application.Queries.RivalsTop(ctx, me)
	if err != nil {
		return fmt.Errorf("application.Queries.RivalsTop(): %w", err)
	}

	w := newTable()
	w.AppendHeader(table.Row{"Opponent", "Games", "Wins", "Losses", "Win %", "MMR diff", "Note"})
	for _, row := range rows {
		difference := row.TheirMmr - row.MyMmr

		note := ""
		decided := row.Wins + row.Losses
		if decided > 0 {
			rate := 100 * float64(row.Wins) / float64(decided)
			if rate < 45 && difference <= 50 {
				note = "beats us at even MMR"
			}
			if rate >= 60 && difference >= 100 {
				note = "we punch up"
			}
		}

		w.AppendRow(table.Row{row.Name, row.Games, row.Wins, row.Losses, percentage(row.Wins, row.Losses), fmt.Sprintf("%+d", difference), note})
	}
	fmt.Println(w.Render())

	return nil
}
