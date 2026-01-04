package main

import (
	"context"
	"os"

	"main/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	pp *pgxpool.Pool
	mq *models.Queries
)

func init() {
	err := godotenv.Load()
	check(err)

	ctx := context.Background()
	pp, err = pgxpool.New(ctx, os.Getenv("POSTGRES"))
	check(err)

	mq = models.New(pp)
}
