.PHONY: build clean down ingest lint reset sqlc up

build:
	docker compose build

clean:
	docker compose down
	docker run --rm -v "$(PWD)":/sources postgres:18.4 rm -rf /sources/postgres

down:
	docker compose down

ingest:
	docker compose run --rm cli -action ingest
	docker compose down

lint:
	docker compose run --rm --no-deps --entrypoint sh cli -c 'sqlc generate && golangci-lint run ./... && biome check index.css index.html index.js'

reset:
	docker compose exec -T postgres psql -U postgres -d starcraft2 < sqlc/schema.sql

sqlc:
	docker compose run --rm --no-deps --entrypoint sqlc cli generate

up:
	docker compose up --build
