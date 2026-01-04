package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/icza/s2prot"
	"github.com/icza/s2prot/rep"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
	"github.com/spf13/cast"
)

func buildColor(color [4]byte) (string, error) { // nolint
	return fmt.Sprintf("#%02X%02X%02X%02X", color[0], color[1], color[2], color[3]), nil
}

func buildFiles(paths []string) ([]string, error) {
	items := []Item{}

	for _, path := range paths {
		err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("filepath.Walk(): %w", err)
			}

			if !strings.HasSuffix(path, ".SC2Replay") {
				return nil
			}

			items = append(items, Item{ModTime: info.ModTime(), Path: path})

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("filepath.Walk(): %w", err)
		}
	}

	sort.Slice(items, func(a int, z int) bool {
		return items[a].ModTime.After(items[z].ModTime)
	})

	files := make([]string, len(items))
	for key, value := range items {
		files[key] = value.Path
	}

	return files, nil
}

func buildGame(file string, r *rep.Rep) (Game, error) {
	messages, err := buildMessages(r)
	if err != nil {
		return Game{}, fmt.Errorf("buildMessages(): %w", err)
	}

	stats, err := buildStats(r)
	if err != nil {
		return Game{}, fmt.Errorf("buildStats(): %w", err)
	}

	units, err := buildUnits(r)
	if err != nil {
		return Game{}, fmt.Errorf("buildUnits(): %w", err)
	}

	players, err := buildPlayers(r, messages, stats, units)
	if err != nil {
		return Game{}, fmt.Errorf("buildPlayers(): %w", err)
	}

	teams, err := buildTeams(players)
	if err != nil {
		return Game{}, fmt.Errorf("buildTeams(): %w", err)
	}

	gameType, err := buildType(r)
	if err != nil {
		return Game{}, fmt.Errorf("buildType(): %w", err)
	}

	game := Game{
		File:      file,
		Duration:  int64(r.Metadata.DurationSec()),
		Map:       r.Metadata.Title(),
		Mode:      r.AttrEvts.GameMode().String(),
		Timestamp: pgtype.Timestamp{Time: r.Details.TimeUTC(), Valid: true},
		Type:      gameType,
		Teams:     teams,
	}

	return game, nil
}

func buildMessages(r *rep.Rep) ([]*Message, error) {
	messages := []*Message{}

	for _, message := range r.MessageEvts {
		if message.Name != "Chat" {
			continue
		}

		message := &Message{
			Player:      message.Value("userid", "userId").(int64) + 1,
			Time:        message.Value("loop").(int64),
			RecipientID: message.Value("recipient").(int64),
			String:      message.Value("string").(string),
		}
		messages = append(messages, message)
	}

	return messages, nil
}

func buildPlayers(r *rep.Rep, messages []*Message, stats []*Stat, units []*Unit) ([]*Player, error) {
	players := []*Player{}

	detailsPlayerList := r.Details.Players()
	initDataUserInitData := r.InitData.UserInitDatas
	metadataPlayers := r.Metadata.Players()

	detailID := 0
	playerID := 1

	for key, value := range r.InitData.LobbyState.Slots {
		control := value.Control()

		observe := value.Observe()

		if control.Name == "Human" {
			if observe.Name == "Participant" {
				id := value.UserID()

				iduid := initDataUserInitData[id]

				dpl := detailsPlayerList[detailID]

				player := &Player{}

				player.Number = int64(key) + 1

				color, err := buildColor(dpl.Color)
				if err != nil {
					return nil, fmt.Errorf("buildColor(): %w", err)
				}
				player.Color = color
				player.Control = control.Name
				player.MMR = iduid.MMR()
				player.Name = dpl.Name
				player.Observe = observe.Name

				for _, v := range metadataPlayers {
					if v.PlayerID() == player.Number {
						player.APM = int64(v.APM())
						player.RacesAssigned = v.AssignedRace()
						player.RacesSelected = v.SelectedRace()
						player.Result = v.Result()
					}
				}

				for _, message := range messages {
					if message.Player == player.Number {
						player.Messages = append(player.Messages, message)
					}
				}

				for _, stat := range stats {
					if stat.Player == player.Number {
						player.Stats = append(player.Stats, stat)
					}
				}

				for _, unit := range units {
					if unit.Player == player.Number {
						player.Units = append(player.Units, unit)
					}
				}

				players = append(players, player)

				detailID++
				playerID++
			} else {
				playerID++
			}
		}
		if control.Name == "Computer" {
			detailID++
			playerID++
		}

	}

	return players, nil
}

func buildStats(r *rep.Rep) ([]*Stat, error) {
	stats := []*Stat{}

	for _, event := range r.TrackerEvts.Evts {
		_ = event.Value("evtTypeName").(string)

		playerID := event.Value("playerId")
		if playerID == nil {
			continue
		}

		v := event.Value("stats")
		if v == nil {
			continue
		}

		numbers := v.(s2prot.Struct)

		stat := &Stat{
			Player:                           cast.ToInt64(playerID),
			Time:                             cast.ToInt64(event.Value("loop")),
			FoodMade:                         cast.ToInt64(numbers.Value("scoreValueFoodMade")),
			FoodUsed:                         cast.ToInt64(numbers.Value("scoreValueFoodUsed")),
			MineralsCollectionRate:           cast.ToInt64(numbers.Value("scoreValueMineralsCollectionRate")),
			MineralsCurrent:                  cast.ToInt64(numbers.Value("scoreValueMineralsCurrent")),
			MineralsFriendlyFireArmy:         cast.ToInt64(numbers.Value("scoreValueMineralsFriendlyFireArmy")),
			MineralsFriendlyFireEconomy:      cast.ToInt64(numbers.Value("scoreValueMineralsFriendlyFireEconomy")),
			MineralsFriendlyFireTechnology:   cast.ToInt64(numbers.Value("scoreValueMineralsFriendlyFireTechnology")),
			MineralsKilledArmy:               cast.ToInt64(numbers.Value("scoreValueMineralsKilledArmy")),
			MineralsKilledEconomy:            cast.ToInt64(numbers.Value("scoreValueMineralsKilledEconomy")),
			MineralsKilledTechnology:         cast.ToInt64(numbers.Value("scoreValueMineralsKilledTechnology")),
			MineralsLostArmy:                 cast.ToInt64(numbers.Value("scoreValueMineralsLostArmy")),
			MineralsLostEconomy:              cast.ToInt64(numbers.Value("scoreValueMineralsLostEconomy")),
			MineralsLostTechnology:           cast.ToInt64(numbers.Value("scoreValueMineralsLostTechnology")),
			MineralsUsedActiveForces:         cast.ToInt64(numbers.Value("scoreValueMineralsUsedActiveForces")),
			MineralsUsedCurrentArmy:          cast.ToInt64(numbers.Value("scoreValueMineralsUsedCurrentArmy")),
			MineralsUsedCurrentEconomy:       cast.ToInt64(numbers.Value("scoreValueMineralsUsedCurrentEconomy")),
			MineralsUsedCurrentTechnology:    cast.ToInt64(numbers.Value("scoreValueMineralsUsedCurrentTechnology")),
			MineralsUsedInProgressArmy:       cast.ToInt64(numbers.Value("scoreValueMineralsUsedInProgressArmy")),
			MineralsUsedInProgressEconomy:    cast.ToInt64(numbers.Value("scoreValueMineralsUsedInProgressEconomy")),
			MineralsUsedInProgressTechnology: cast.ToInt64(numbers.Value("scoreValueMineralsUsedInProgressTechnology")),
			VespeneCollectionRate:            cast.ToInt64(numbers.Value("scoreValueVespeneCollectionRate")),
			VespeneCurrent:                   cast.ToInt64(numbers.Value("scoreValueVespeneCurrent")),
			VespeneFriendlyFireArmy:          cast.ToInt64(numbers.Value("scoreValueVespeneFriendlyFireArmy")),
			VespeneFriendlyFireEconomy:       cast.ToInt64(numbers.Value("scoreValueVespeneFriendlyFireEconomy")),
			VespeneFriendlyFireTechnology:    cast.ToInt64(numbers.Value("scoreValueVespeneFriendlyFireTechnology")),
			VespeneKilledArmy:                cast.ToInt64(numbers.Value("scoreValueVespeneKilledArmy")),
			VespeneKilledEconomy:             cast.ToInt64(numbers.Value("scoreValueVespeneKilledEconomy")),
			VespeneKilledTechnology:          cast.ToInt64(numbers.Value("scoreValueVespeneKilledTechnology")),
			VespeneLostArmy:                  cast.ToInt64(numbers.Value("scoreValueVespeneLostArmy")),
			VespeneLostEconomy:               cast.ToInt64(numbers.Value("scoreValueVespeneLostEconomy")),
			VespeneLostTechnology:            cast.ToInt64(numbers.Value("scoreValueVespeneLostTechnology")),
			VespeneUsedActiveForces:          cast.ToInt64(numbers.Value("scoreValueVespeneUsedActiveForces")),
			VespeneUsedCurrentArmy:           cast.ToInt64(numbers.Value("scoreValueVespeneUsedCurrentArmy")),
			VespeneUsedCurrentEconomy:        cast.ToInt64(numbers.Value("scoreValueVespeneUsedCurrentEconomy")),
			VespeneUsedCurrentTechnology:     cast.ToInt64(numbers.Value("scoreValueVespeneUsedCurrentTechnology")),
			VespeneUsedInProgressArmy:        cast.ToInt64(numbers.Value("scoreValueVespeneUsedInProgressArmy")),
			VespeneUsedInProgressEconomy:     cast.ToInt64(numbers.Value("scoreValueVespeneUsedInProgressEconomy")),
			VespeneUsedInProgressTechnology:  cast.ToInt64(numbers.Value("scoreValueVespeneUsedInProgressTechnology")),
			WorkersActiveCount:               cast.ToInt64(numbers.Value("scoreValueWorkersActiveCount")),
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

func buildTeams(players []*Player) ([]*Team, error) {
	teams := []*Team{}

	numbers := []int64{}
	for _, player := range players {
		numbers = append(numbers, player.Team)
	}
	numbers = lo.Uniq(numbers)

	for _, number := range numbers {
		team := &Team{
			Number:  number,
			Result:  "",
			Players: []*Player{},
		}
		for _, player := range players {
			if team.Number == player.Team {
				team.Players = append(team.Players, player)
				team.Result = player.Result
			}
		}
		teams = append(teams, team)
	}

	if len(teams) == 2 {
		if teams[0].Result == "Undecided" && teams[1].Result == "Undecided" {
			for team := range teams {
				for _, player := range teams[team].Players {
					if player.Name == "MuNi" || player.Name == "Pineapple" || player.Name == "SINDIOS" {
						teams[team].Result = "Loss"
						teams[1-team].Result = "Win"
						break
					}
				}
			}
		}
	}

	for team := range teams {
		for player := range teams[team].Players {
			teams[team].Players[player].Result = teams[team].Result
		}
	}

	return teams, nil
}

func buildType(r *rep.Rep) (string, error) {
	v := r.AttrEvts.Value("scopes", "16", "2001")
	if v == nil {
		return "", nil
	}

	s := v.(s2prot.Struct)

	return cast.ToString(s.Value("value")), nil
}

func buildUnits(r *rep.Rep) ([]*Unit, error) {
	units := []*Unit{}

	names := map[string]string{}

	for _, event := range r.TrackerEvts.Evts {
		action := event.Value("evtTypeName").(string)
		if action != "UnitBorn" {
			continue
		}

		unitTagIndex := event.Value("unitTagIndex").(int64)
		unitTagRecycle := event.Value("unitTagRecycle").(int64)
		unitTypeName := event.Value("unitTypeName").(string)
		names[fmt.Sprintf("%d_%d", unitTagIndex, unitTagRecycle)] = unitTypeName
	}

	for _, event := range r.TrackerEvts.Evts {
		action := event.Value("evtTypeName").(string)

		if action == "UnitBorn" {
			controlPlayerID := event.Value("controlPlayerId")
			if controlPlayerID == nil {
				continue
			}

			unit := &Unit{
				Player: controlPlayerID.(int64),
				Time:   event.Value("loop").(int64),
				Action: action,
				Name:   event.Value("unitTypeName").(string),
				X:      event.Value("x").(int64),
				Y:      event.Value("y").(int64),
			}

			units = append(units, unit)
		}

		if action == "UnitDied" {
			killerPlayerID := event.Value("killerPlayerId")
			if killerPlayerID == nil {
				continue
			}

			name, ok := names[fmt.Sprintf("%d_%d", event.Value("unitTagIndex").(int64), event.Value("unitTagRecycle").(int64))]
			if !ok {
				continue
			}

			unit := &Unit{
				Player: killerPlayerID.(int64),
				Time:   event.Value("loop").(int64),
				Action: action,
				Name:   name,
				X:      event.Value("x").(int64),
				Y:      event.Value("y").(int64),
			}

			units = append(units, unit)
		}
	}

	return units, nil
}
