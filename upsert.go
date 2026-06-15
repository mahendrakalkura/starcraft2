package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"main/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func insertReplay(ctx context.Context, application *Application, path string, replay *Replay, fingerprint string) error {
	tx, err := application.Database.Begin(ctx)
	if err != nil {
		return fmt.Errorf("application.Database.Begin(): %w", err)
	}
	defer func() {
		err := tx.Rollback(ctx)
		if err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			fmt.Fprintln(os.Stderr, err)
		}
	}()

	queries := application.Queries.WithTx(tx)

	fileParams := models.FilesInsertOneParams{
		Path:          path,
		ParserVersion: replay.ParserVersion,
		Status:        statusImported,
	}
	fileID, err := queries.FilesInsertOne(ctx, fileParams)
	if err != nil {
		return fmt.Errorf("queries.FilesInsertOne(): %w", err)
	}

	gameParams := models.GamesInsertOneParams{
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
	gameID, err := queries.GamesInsertOne(ctx, gameParams)
	if err != nil {
		return fmt.Errorf("queries.GamesInsertOne(): %w", err)
	}

	for _, player := range replay.Players {
		playerParams := models.PlayersInsertOneParams{
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
		playerID, e := queries.PlayersInsertOne(ctx, playerParams)
		if e != nil {
			return fmt.Errorf("queries.PlayersInsertOne(): %w", e)
		}

		messageParams := make([]models.MessagesInsertManyParams, len(player.Messages))
		for key, message := range player.Messages {
			messageParams[key] = models.MessagesInsertManyParams{
				PlayerID:  playerID,
				Recipient: message.Recipient,
				Text:      message.Text,
				Tick:      message.Tick,
			}
		}
		_, err = queries.MessagesInsertMany(ctx, messageParams)
		if err != nil {
			return fmt.Errorf("queries.MessagesInsertMany(): %w", err)
		}

		statParams := make([]models.StatsInsertManyParams, len(player.Stats))
		for key, stat := range player.Stats {
			statParams[key] = models.StatsInsertManyParams{
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
		_, err = queries.StatsInsertMany(ctx, statParams)
		if err != nil {
			return fmt.Errorf("queries.StatsInsertMany(): %w", err)
		}

		unitParams := make([]models.UnitsInsertManyParams, len(player.Units))
		for key, unit := range player.Units {
			unitParams[key] = models.UnitsInsertManyParams{
				PlayerID: playerID,
				Action:   unit.Action,
				Name:     unit.Name,
				Tick:     unit.Tick,
				X:        unit.X,
				Y:        unit.Y,
			}
		}
		_, err = queries.UnitsInsertMany(ctx, unitParams)
		if err != nil {
			return fmt.Errorf("queries.UnitsInsertMany(): %w", err)
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
