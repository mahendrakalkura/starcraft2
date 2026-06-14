package main

import "fmt"

// percentage renders wins/(wins+losses) as "NN.N", or "" with no decided games.
func percentage(wins int64, losses int64) string {
	decided := wins + losses
	if decided == 0 {
		return ""
	}
	return fmt.Sprintf("%.1f", 100*float64(wins)/float64(decided))
}

// resolveMe returns the explicit -me flag value, falling back to the first
// PLAYERS entry.
func resolveMe(application *Application, me string) (string, error) {
	if me != "" {
		return me, nil
	}
	if len(application.Settings.Players) > 0 {
		return application.Settings.Players[0], nil
	}
	return "", fmt.Errorf("set --me or the PLAYERS environment variable")
}
