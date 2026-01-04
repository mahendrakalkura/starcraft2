package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func statistics(ctx context.Context, application *Application) {
	statistics, err := application.Queries.Statistics(ctx)
	if err != nil {
		panic(fmt.Errorf("application.Queries.Statistics(): %w", err))
	}

	t := table.NewWriter()

	t.AppendHeader(table.Row{"Table", "Count"})
	t.SetStyle(table.StyleLight)

	t.Style().Format.Header = text.FormatDefault

	for _, stat := range statistics {
		t.AppendRow(table.Row{stat.Key, stat.Value})
	}

	fmt.Println(t.Render())
}
