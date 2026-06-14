package main

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"time"
)

// fingerprint identifies a match independent of file path: same game copied to
// several files produces the same fingerprint.
func fingerprint(replay *Replay) string {
	roster := make([]string, len(replay.Players))
	for key, player := range replay.Players {
		roster[key] = fmt.Sprintf("%d:%s", player.Team, player.Name)
	}
	sort.Strings(roster)

	payload := fmt.Sprintf(
		"%s|%s|%s",
		replay.Game.PlayedAt.UTC().Format(time.RFC3339),
		replay.Game.Map,
		strings.Join(roster, ","),
	)

	return fmt.Sprintf("%x", sha256.Sum256([]byte(payload)))
}
