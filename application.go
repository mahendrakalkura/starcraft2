package main

import (
	"context"
	"fmt"

	"main/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Application struct {
	DB       *pgxpool.Pool
	Queries  *models.Queries
	Settings *Settings
}

func NewApplication() (*Application, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("godotenv.Load(): %w", err)
	}

	settings := NewSettings()

	ctx := context.Background()
	db, err := pgxpool.New(ctx, settings.Database)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New(): %w", err)
	}

	queries := models.New(db)

	return &Application{
		DB:       db,
		Queries:  queries,
		Settings: settings,
	}, nil
}

func (a *Application) Close() {
	if a.DB != nil {
		a.DB.Close()
	}
}
