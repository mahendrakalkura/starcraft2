package main

import (
	"context"
	"os"

	"main/models"
)

func upsert(game Game) {
	ctx := context.Background()

	giop := models.GamesInsertOneParams{
		Duration:  game.Duration,
		File:      game.File,
		Map:       game.Map,
		Mode:      game.Mode,
		Timestamp: game.Timestamp,
		Type:      game.Type,
	}
	gameID, err := mq.GamesInsertOne(ctx, giop)
	check(err)

	for _, team := range game.Teams {
		tiop := models.TeamsInsertOneParams{
			GameID: gameID,
			Number: team.Number,
			Result: team.Result,
		}
		teamID, err := mq.TeamsInsertOne(ctx, tiop)
		check(err)

		for _, player := range team.Players {
			siop := models.PlayersInsertOneParams{
				TeamID:        teamID,
				Number:        player.Number,
				APM:           player.APM,
				Color:         player.Color,
				Control:       player.Control,
				MMR:           player.MMR,
				Name:          player.Name,
				RacesAssigned: player.RacesAssigned,
				RacesSelected: player.RacesSelected,
			}
			playerID, err := mq.PlayersInsertOne(ctx, siop)
			if err != nil {
				err = os.WriteFile("dump.json", []byte(dump(game)), 0o644)
				check(err)
			}
			check(err)

			mimps := []models.MessagesInsertManyParams{}
			for _, message := range player.Messages {
				mimp := models.MessagesInsertManyParams{
					PlayerID:    playerID,
					Time:        message.Time,
					RecipientID: message.RecipientID,
					String:      message.String,
				}
				mimps = append(mimps, mimp)
			}
			mim := mq.MessagesInsertMany(ctx, mimps)
			mim.Exec(func(key int, err error) {
				check(err)
			})
			_ = mim.Close()

			simps := []models.StatsInsertManyParams{}
			for _, stat := range player.Stats {
				simp := models.StatsInsertManyParams{
					PlayerID:                         playerID,
					Time:                             stat.Time,
					FoodMade:                         stat.FoodMade,
					FoodUsed:                         stat.FoodUsed,
					MineralsCollectionRate:           stat.MineralsCollectionRate,
					MineralsCurrent:                  stat.MineralsCurrent,
					MineralsFriendlyFireArmy:         stat.MineralsFriendlyFireArmy,
					MineralsFriendlyFireEconomy:      stat.MineralsFriendlyFireEconomy,
					MineralsFriendlyFireTechnology:   stat.MineralsFriendlyFireTechnology,
					MineralsKilledArmy:               stat.MineralsKilledArmy,
					MineralsKilledEconomy:            stat.MineralsKilledEconomy,
					MineralsKilledTechnology:         stat.MineralsKilledTechnology,
					MineralsLostArmy:                 stat.MineralsLostArmy,
					MineralsLostEconomy:              stat.MineralsLostEconomy,
					MineralsLostTechnology:           stat.MineralsLostTechnology,
					MineralsUsedActiveForces:         stat.MineralsUsedActiveForces,
					MineralsUsedCurrentArmy:          stat.MineralsUsedCurrentArmy,
					MineralsUsedCurrentEconomy:       stat.MineralsUsedCurrentEconomy,
					MineralsUsedCurrentTechnology:    stat.MineralsUsedCurrentTechnology,
					MineralsUsedInProgressArmy:       stat.MineralsUsedInProgressArmy,
					MineralsUsedInProgressEconomy:    stat.MineralsUsedInProgressEconomy,
					MineralsUsedInProgressTechnology: stat.MineralsUsedInProgressTechnology,
					VespeneCollectionRate:            stat.VespeneCollectionRate,
					VespeneCurrent:                   stat.VespeneCurrent,
					VespeneFriendlyFireArmy:          stat.VespeneFriendlyFireArmy,
					VespeneFriendlyFireEconomy:       stat.VespeneFriendlyFireEconomy,
					VespeneFriendlyFireTechnology:    stat.VespeneFriendlyFireTechnology,
					VespeneKilledArmy:                stat.VespeneKilledArmy,
					VespeneKilledEconomy:             stat.VespeneKilledEconomy,
					VespeneKilledTechnology:          stat.VespeneKilledTechnology,
					VespeneLostArmy:                  stat.VespeneLostArmy,
					VespeneLostEconomy:               stat.VespeneLostEconomy,
					VespeneLostTechnology:            stat.VespeneLostTechnology,
					VespeneUsedActiveForces:          stat.VespeneUsedActiveForces,
					VespeneUsedCurrentArmy:           stat.VespeneUsedCurrentArmy,
					VespeneUsedCurrentEconomy:        stat.VespeneUsedCurrentEconomy,
					VespeneUsedCurrentTechnology:     stat.VespeneUsedCurrentTechnology,
					VespeneUsedInProgressArmy:        stat.VespeneUsedInProgressArmy,
					VespeneUsedInProgressEconomy:     stat.VespeneUsedInProgressEconomy,
					VespeneUsedInProgressTechnology:  stat.VespeneUsedInProgressTechnology,
					WorkersActiveCount:               stat.WorkersActiveCount,
				}
				simps = append(simps, simp)
			}
			smi := mq.StatsInsertMany(ctx, simps)
			smi.Exec(func(key int, err error) {
				check(err)
			})
			_ = smi.Close()

			uimps := []models.UnitsInsertManyParams{}
			for _, unit := range player.Units {
				uimp := models.UnitsInsertManyParams{
					PlayerID: playerID,
					Time:     unit.Time,
					Action:   unit.Action,
					Name:     unit.Name,
					X:        unit.X,
					Y:        unit.Y,
				}
				uimps = append(uimps, uimp)
			}
			umi := mq.UnitsInsertMany(ctx, uimps)
			umi.Exec(func(key int, err error) {
				check(err)
			})
			_ = umi.Close()
		}
	}
}
