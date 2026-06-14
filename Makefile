.PHONY: build reset sqlc up down ingest

build:
	cd sidecar && cargo build --release
	cp sidecar/target/release/sc2json sc2json
	gofmt -l -w .
	go mod tidy
	go build -o main .

reset:
	psql "$$(grep ^DATABASE .env | cut -d= -f2-)" < sqlc/schema.sql
	psql "$$(grep ^DATABASE .env | cut -d= -f2-)" < sqlc/chat_schema.sql

sqlc:
	sqlc generate

# Docker Compose helpers.
up:
	docker compose up -d --build

down:
	docker compose down

ingest:
	docker compose run --rm cli -action ingest
