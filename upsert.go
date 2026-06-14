package main

import (
	"context"
	"errors"
	"fmt"

	"main/models"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func insertReplay(ctx context.Context, application *Application, path string, replay *Replay, fingerprint string) error {
	tx, err := application.Database.Begin(ctx)
	if err != nil {
		return fmt.Errorf("application.Database.Begin(): %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	q := application.Queries.WithTx(tx)

	fiop := models.FilesInsertOneParams{
		Path:          path,
		ParserVersion: replay.ParserVersion,
		Status:        statusImported,
	}
	fileID, err := q.FilesInsertOne(ctx, fiop)
	if err != nil {
		return fmt.Errorf("q.FilesInsertOne(): %w", err)
	}

	giop := models.GamesInsertOneParams{
		FileID:      fileID,
		Amm:         replay.Game.Amm,
		Competitive: replay.Game.Competitive,
		Duration:    replay.Game.Duration,
		Fingerprint: fingerprint,
		Map:         replay.Game.Map,
		Mode:        replay.Game.Mode,
		PlayedAt:    pgtype.Timestamptz{Time: replay.Game.PlayedAt, Valid: true},
		Version:     replay.Game.Version,
	}
	gameID, err := q.GamesInsertOne(ctx, giop)
	if err != nil {
		return fmt.Errorf("q.GamesInsertOne(): %w", err)
	}

	for _, player := range replay.Players {
		piop := models.PlayersInsertOneParams{
			GameID:       gameID,
			Apm:          player.Apm,
			Clan:         player.Clan,
			Color:        player.Color,
			Mmr:          player.Mmr,
			Name:         player.Name,
			Number:       player.Number,
			RaceAssigned: player.RaceAssigned,
			RaceSelected: player.RaceSelected,
			Result:       player.Result,
			Team:         player.Team,
		}
		playerID, e := q.PlayersInsertOne(ctx, piop)
		if e != nil {
			return fmt.Errorf("q.PlayersInsertOne(): %w", e)
		}

		mimps := make([]models.MessagesInsertManyParams, len(player.Messages))
		for key, message := range player.Messages {
			mimps[key] = models.MessagesInsertManyParams{
				PlayerID:  playerID,
				Recipient: message.Recipient,
				Text:      message.Text,
				Tick:      message.Tick,
			}
		}
		_, err = q.MessagesInsertMany(ctx, mimps)
		if err != nil {
			return fmt.Errorf("q.MessagesInsertMany(): %w", err)
		}

		simps := make([]models.StatsInsertManyParams, len(player.Stats))
		for key, stat := range player.Stats {
			simps[key] = models.StatsInsertManyParams{
				PlayerID:                         playerID,
				Tick:                             stat.Tick,
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
		}
		_, err = q.StatsInsertMany(ctx, simps)
		if err != nil {
			return fmt.Errorf("q.StatsInsertMany(): %w", err)
		}

		uimps := make([]models.UnitsInsertManyParams, len(player.Units))
		for key, unit := range player.Units {
			uimps[key] = models.UnitsInsertManyParams{
				PlayerID: playerID,
				Action:   unit.Action,
				Name:     unit.Name,
				Tick:     unit.Tick,
				X:        unit.X,
				Y:        unit.Y,
			}
		}
		_, err = q.UnitsInsertMany(ctx, uimps)
		if err != nil {
			return fmt.Errorf("q.UnitsInsertMany(): %w", err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("tx.Commit(): %w", err)
	}

	return nil
}

// isFingerprintConflict reports whether the error is a unique violation on the
// games fingerprint, i.e. another worker imported a copy of this match first.
func isFingerprintConflict(err error) bool {
	pgError := &pgconn.PgError{}
	if errors.As(err, &pgError) {
		return pgError.Code == "23505" && pgError.ConstraintName == "games_fingerprint_key"
	}
	return false
}
