package main

import (
	"context"
	"fmt"

	"main/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Application struct {
	Database *pgxpool.Pool
	Queries  *models.Queries
	Settings *Settings
}

func (a *Application) Close() {
	if a.Database != nil {
		a.Database.Close()
	}
}

func NewApplication(ctx context.Context) (*Application, error) {
	settings, err := NewSettings()
	if err != nil {
		return nil, fmt.Errorf("NewSettings(): %w", err)
	}

	database, err := pgxpool.New(ctx, settings.DatabaseURL())
	if err != nil {
		return nil, fmt.Errorf("pgxpool.New(): %w", err)
	}

	queries := models.New(database)

	application := &Application{
		Database: database,
		Queries:  queries,
		Settings: settings,
	}

	return application, nil
}
