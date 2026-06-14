package main

import "slices"

// resolveResults handles games where every result is "Undecided" (e.g. the
// recording player left before the official result): if a tracked player is
// found, their team is marked as the loser and the other team as the winner.
// Only two-team games are resolved.
func resolveResults(replay *Replay, tracked []string) {
	teams := map[int32]bool{}
	for _, player := range replay.Players {
		teams[player.Team] = true
		if player.Result != "Undecided" {
			return
		}
	}
	if len(teams) != 2 {
		return
	}

	loser := int32(0)
	for _, player := range replay.Players {
		if slices.Contains(tracked, player.Name) {
			loser = player.Team
			break
		}
	}
	if loser == 0 {
		return
	}

	for key, player := range replay.Players {
		if player.Team == loser {
			replay.Players[key].Result = "Loss"
		} else {
			replay.Players[key].Result = "Win"
		}
	}
}
