package main

import (
	"context"
	"fmt"

	"github.com/jedib0t/go-pretty/v6/table"
)

func chat(ctx context.Context, application *Application, me string) error {
	me, err := resolveMe(application, me)
	if err != nil {
		return err
	}

	fmt.Printf("%s: chat habits\n\n", me)

	phrases, err := application.Queries.ChatPhrases(ctx, me)
	if err != nil {
		return fmt.Errorf("application.Queries.ChatPhrases(): %w", err)
	}

	w := newTable()
	w.AppendHeader(table.Row{"Phrase", "Times"})
	for _, phrase := range phrases {
		w.AppendRow(table.Row{phrase.Phrase, phrase.Count})
	}
	fmt.Println(w.Render())

	ggs, err := application.Queries.ChatGg(ctx, me)
	if err != nil {
		return fmt.Errorf("application.Queries.ChatGg(): %w", err)
	}

	w = newTable()
	w.AppendHeader(table.Row{"Result", "Games", "Said gg", "gg %"})
	for _, gg := range ggs {
		w.AppendRow(table.Row{gg.Result, gg.Games, gg.SaidGg, percentage(gg.SaidGg, gg.Games-gg.SaidGg)})
	}
	fmt.Println(w.Render())

	return nil
}
