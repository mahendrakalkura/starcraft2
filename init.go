package main

import (
	"context"
	"os"
	"starcraft2/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var pp *pgxpool.Pool
var mq *models.Queries

func init() {
	err := godotenv.Load()
	checkErr(err)

	ctx := context.Background()
	pp, err = pgxpool.New(ctx, os.Getenv("POSTGRES"))
	checkErr(err)

	mq = models.New(pp)
}
