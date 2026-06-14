# StarCraft II Replay Analyzer

Parses StarCraft II replay files (.SC2Replay) into PostgreSQL and serves a natural-language Q&A web page over the data. Ask questions in plain English; an AI agent writes a read-only SQL query, runs it, and answers from the rows it gets back.

Everything runs in Docker containers. There is no host build or host run path: the Go binary, the Rust sidecar, and `sqlc` all run inside the image, against the repo mounted at `/sources`. For how the system works internally (architecture, the agent loop, the data model), see CONTEXT.md. To install on a server behind nginx, see DEPLOY.md.

## Requirements

- Docker Engine with the Docker Compose plugin.
- An OpenRouter API key (for the web UI).

## Run locally (fresh clone)

```bash
git clone <repo-url> starcraft2
cd starcraft2

cp .env.sample .env        # then edit: OPENROUTER_API_KEY, PLAYERS, REPLAYS_HOST, POSTGRES_PASSWORD
make up                    # build the image and start postgres + ui
```

`.env.sample` ships with `ENVIRONMENT=development`, so the `ui` container runs under [air](https://github.com/air-verse/air): edit any `.go`, `.sql`, or frontend file and the server rebuilds and restarts automatically. The repo is mounted into the container at `/sources`, and `models/` is generated in-container on startup (run `make sqlc` if you also want it on the host for your editor).

Open http://localhost:8080 and ask a question. The UI is published on `127.0.0.1` only.

To import replays (the `cli` service is on a profile, so it does not start with `make up`):

```bash
make up                    # postgres must be running first
make ingest                # docker compose run --rm cli -action ingest
```

`make ingest` runs a one-shot container that parses every replay under `REPLAYS_HOST`, stores the new ones, and exits. `postgres` and `ui` keep running. Re-running is cheap: already-processed paths are skipped.

## Configuration

All configuration is in `.env`, read by Docker Compose for `${VAR}` substitution. Copy `.env.sample` to `.env` and edit.

```
+--------------------+----------------------------------------------------------+
| Variable           | Purpose                                                  |
+--------------------+----------------------------------------------------------+
| ENVIRONMENT        | development (air hot reload) or production (build once). |
| OPENROUTER_API_KEY | OpenRouter key for the web UI (required).                |
| OPENROUTER_MODEL   | Any tool-calling model. Defaults to deepseek/deepseek-chat. |
| PLAYERS            | Tracked player names, comma-separated.                   |
| PORT               | Host port for the UI (published on 127.0.0.1).           |
| POSTGRES_PASSWORD  | Password for the postgres container.                     |
| REPLAYS_HOST       | Host replay directory, mounted read-only at /replays.    |
| WORKERS            | Ingest worker count. Empty = CPU count.                  |
+--------------------+----------------------------------------------------------+
```

`REPLAYS` inside the container is fixed to `/replays`; ingest walks it recursively, so point `REPLAYS_HOST` at a parent directory to cover several replay folders at once.

## Services

`docker compose up` starts `postgres` (the database) and `ui` (the web server). `cli` is the same image on a `cli` profile for one-off commands and does not start by default.

```
+----------+-----------------------------------+----------------------------+
| Service  | What it runs                      | Lifecycle                  |
+----------+-----------------------------------+----------------------------+
| cli      | main -action ingest|sample|...    | on-demand (compose run)    |
| postgres | postgres:18.4                     | always up, healthchecked   |
| ui       | main -action serve                | always up, depends on db   |
+----------+-----------------------------------+----------------------------+
```

The CLI actions:

```bash
make ingest                                          # import replays
docker compose run --rm cli -action statistics       # row counts per table
docker compose run --rm cli -action sample -file /replays/some.SC2Replay   # pretty-print parsed JSON
```

## Make targets

```
make build    # docker compose build
make up        # build and start postgres + ui (detached)
make down      # stop and remove containers
make ingest    # one-shot replay import
make sqlc      # regenerate models/ on the host (runs sqlc in a container)
make reset     # re-apply sqlc/schema.sql to the running postgres (WARNING: drops all data)
```

## Reset the database

`make reset` runs `sqlc/schema.sql` against the running `postgres` container. The schema starts with `DROP SCHEMA public CASCADE`, so this wipes all games and saved chats:

```bash
make reset
make ingest
```

To start completely fresh, stop the stack and delete the bind-mounted data directory (`./postgres`), then `make up`.

## License

See LICENSE file for details.
