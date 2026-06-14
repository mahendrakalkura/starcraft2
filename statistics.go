package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func statistics(ctx context.Context, application *Application) error {
	statuses, err := application.Queries.FilesCountByStatus(ctx)
	if err != nil {
		return fmt.Errorf("application.Queries.FilesCountByStatus(): %w", err)
	}

	games, err := application.Queries.GamesCount(ctx)
	if err != nil {
		return fmt.Errorf("application.Queries.GamesCount(): %w", err)
	}

	players, err := application.Queries.PlayersCount(ctx)
	if err != nil {
		return fmt.Errorf("application.Queries.PlayersCount(): %w", err)
	}

	messages, err := application.Queries.MessagesCount(ctx)
	if err != nil {
		return fmt.Errorf("application.Queries.MessagesCount(): %w", err)
	}

	stats, err := application.Queries.StatsCount(ctx)
	if err != nil {
		return fmt.Errorf("application.Queries.StatsCount(): %w", err)
	}

	units, err := application.Queries.UnitsCount(ctx)
	if err != nil {
		return fmt.Errorf("application.Queries.UnitsCount(): %w", err)
	}

	w := table.NewWriter()
	w.AppendHeader(table.Row{"Table", "Rows"})
	w.SetColumnConfigs([]table.ColumnConfig{{Number: 2, Align: text.AlignRight}})
	w.Style().Format.Header = text.FormatDefault

	for _, status := range statuses {
		w.AppendRow(table.Row{fmt.Sprintf("files (%s)", status.Status), status.Count})
	}
	w.AppendRow(table.Row{"games", games})
	w.AppendRow(table.Row{"players", players})
	w.AppendRow(table.Row{"messages", messages})
	w.AppendRow(table.Row{"stats", stats})
	w.AppendRow(table.Row{"units", units})

	fmt.Println(w.Render())

	return nil
}
