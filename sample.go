package main

import (
	"io"
	"log"
	"os"

	"github.com/icza/s2prot/rep"
)

func sample() {
	// file := "/home/mahendra/Windows/Documents and Settings/Name/Documents/StarCraft II/Accounts/395302547/1-S2-1-11323619/Replays/Multiplayer/Dystopian Complex.SC2Replay"
	file := "/home/mahendra/Windows/Documents and Settings/Name/Documents/StarCraft II/Accounts/395302547/2-S2-1-9209518/Replays/Multiplayer/Ashen Cradle (26).SC2Replay"

	log.SetOutput(io.Discard)

	r, err := rep.NewFromFile(file)
	check(err)
	_ = r.Close()

	log.SetOutput(os.Stderr)

	_ = os.WriteFile("r.json", []byte(dump(r)), 0o644)

	game := buildGame(file, r)

	_ = os.WriteFile("game.json", []byte(dump(game)), 0o644)
}
