package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	action := flag.String("action", "ingest", "Action: chat, durations, economy, impact, ingest, maps, matchup, maxout, mmr-history, openings, partners, report-1, report-2, rivals, sample, serve, statistics, streaks, or versus")
	file := flag.String("file", "", "Replay file path (required for sample action)")
	first := flag.String("first", "PhotonCannon", "First unit for the openings action")
	force := flag.Bool("force", false, "Reprocess all replay files from scratch (ingest action)")
	me := flag.String("me", "", "My player name (defaults to the first PLAYERS entry)")
	name := flag.String("name", "", "Player name (required for impact and versus actions)")
	race := flag.String("race", "Protoss", "My race for the maxout and openings actions")
	second := flag.String("second", "Stalker", "Second unit for the openings action")
	unit := flag.String("unit", "Carrier", "Unit for the maxout action")
	window := flag.Int("window", 5, "Opening window in in-game minutes for the openings action")
	with := flag.String("with", "", "Second player name for the impact action pair tables")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := NewApplication(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer application.Close()

	switch *action {
	case "chat":
		err = chat(ctx, application, *me)
	case "durations":
		err = durations(ctx, application, *me)
	case "economy":
		err = economy(ctx, application, *me, *race)
	case "impact":
		err = impact(ctx, application, *name, *with)
	case "ingest":
		err = ingest(ctx, application, *force)
	case "maps":
		err = maps(ctx, application, *me)
	case "matchup":
		err = matchup(ctx, application, *me)
	case "maxout":
		err = maxout(ctx, application, *me, *unit, *race)
	case "mmr-history":
		err = mmrHistory(ctx, application, *me)
	case "openings":
		err = openings(ctx, application, *me, *first, *second, *window, *race)
	case "partners":
		err = partners(ctx, application)
	case "report-1":
		err = report1(ctx, application)
	case "report-2":
		err = report2(ctx, application)
	case "rivals":
		err = rivals(ctx, application, *me)
	case "sample":
		err = sample(ctx, *file)
	case "serve":
		err = serve(ctx, application)
	case "statistics":
		err = statistics(ctx, application)
	case "streaks":
		err = streaks(ctx, application, *me)
	case "versus":
		err = versus(ctx, application, *me, *name)
	default:
		err = fmt.Errorf("unknown action %q", *action)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
