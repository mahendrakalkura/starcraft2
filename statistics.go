package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func statistics() {
	statistics, err := mq.Statistics(context.Background())
	if err != nil {
		panic(fmt.Errorf("mq.Statistics(): %w", err))
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
