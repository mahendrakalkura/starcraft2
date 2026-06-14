.PHONY: build down ingest reset sqlc up

build:
	docker compose build

down:
	docker compose down

ingest:
	docker compose run --rm cli -action ingest

reset:
	docker compose exec -T postgres psql -U postgres -d starcraft2 < sqlc/schema.sql

sqlc:
	docker run --rm -v "$(PWD)":/src -w /src sqlc/sqlc generate

up:
	docker compose up -d --build
