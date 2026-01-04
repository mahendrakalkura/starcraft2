package main

import (
	"context"
	"fmt"

	"main/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	settings *Settings
	pp       *pgxpool.Pool
	mq       *models.Queries
)

func init() {
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Errorf("godotenv.Load(): %w", err))
	}

	settings = NewSettings()

	ctx := context.Background()
	pp, err = pgxpool.New(ctx, settings.Database)
	if err != nil {
		panic(fmt.Errorf("pgxpool.New(): %w", err))
	}

	mq = models.New(pp)
}

func cleanup() {
	if pp != nil {
		pp.Close()
	}
}
