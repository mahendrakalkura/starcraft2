# StarCraft II Replay Analyzer

Parses StarCraft II replay files (.SC2Replay) into PostgreSQL and serves a natural-language Q&A web page over the data. Ask questions in plain English; an AI agent writes a read-only SQL query, runs it, and answers from the rows it gets back.

Everything runs in Docker containers. There is no host build or host run path: the Go binary, the Rust sidecar, and `sqlc` all run inside the image, against the repo mounted at `/sources`. The Go process reads its configuration directly from `.env` (via godotenv) - the compose layer does no variable mapping. For how the system works internally, see CONTEXT.md. To install on a server behind nginx, see DEPLOY.md.

## Requirements

- Docker Engine with the Docker Compose plugin.
- An OpenRouter API key (for the web UI).

## Run locally (fresh clone)

```bash
git clone <repo-url> starcraft2
cd starcraft2

cp .env.sample .env        # then fill in every variable (see Configuration)
make up                    # build and start postgres + ui (streams logs, stays in the foreground)
```

`.env.sample` ships with `GO_ENVIRONMENT=development`, so the `ui` container runs under [air](https://github.com/air-verse/air): edit any `.go`, `.sql`, or frontend file and the server rebuilds and restarts automatically. The repo is mounted into the container at `/sources`, and `models/` is generated in-container on startup (run `make sqlc` if you also want it on the host for your editor).

`make up` runs in the foreground so you can watch the logs; open http://localhost:8080 in a browser and ask a question. The UI is published on `${GO_PORT}` on all interfaces, so put it behind a reverse proxy or firewall in any shared environment (it has no authentication).

To import replays (from a second terminal; the `cli` service is on a profile, so it does not start with `make up`):

```bash
make ingest                # docker compose run --rm cli -action ingest
```

`make ingest` runs a one-shot container that parses every replay under `GO_REPLAYS`, stores the new ones, and exits. `postgres` and `ui` keep running. Re-running is cheap: already-processed paths are skipped.

## Configuration

All configuration lives in `.env`. The Go process loads it directly (godotenv) and the variable names in `.env` are exactly the names used in the code - there is only ever one name per setting. Every variable is required and there are no defaults: if any is missing, the program prints which ones and exits. Copy `.env.sample` to `.env` and fill in all of them.

```
+--------------------+----------------------------------------------------------+
| Variable           | Purpose                                                  |
+--------------------+----------------------------------------------------------+
| GO_ENVIRONMENT     | development (air hot reload) or production (build once). |
| GO_PLAYERS         | Tracked player names, comma-separated.                   |
| GO_PORT            | Port the UI listens on and is published at (127.0.0.1).  |
| GO_REPLAYS         | Replay directory, walked recursively for .SC2Replay.     |
| GO_WORKERS         | Ingest worker count (positive integer).                  |
| OPENROUTER_API_KEY | OpenRouter key for the web UI.                           |
| OPENROUTER_MODEL   | Tool-calling model, e.g. deepseek/deepseek-v4-flash.     |
| POSTGRES_DB        | Database name.                                           |
| POSTGRES_HOST      | Database host (the compose service name: postgres).      |
| POSTGRES_PASSWORD  | Database password.                                       |
| POSTGRES_PORT      | Database port (5432).                                    |
| POSTGRES_USER      | Database user.                                           |
+--------------------+----------------------------------------------------------+
```

The Go process builds its Postgres connection string from the `POSTGRES_*` values, so the password is never duplicated. `GO_REPLAYS` is bind-mounted into the `cli` container at the same path it has on the host, so the path the app walks and the path you set are identical.

## Services

`docker compose up` starts `postgres` (the database) and `ui` (the web server). `cli` is the same image on a `cli` profile for one-off commands and does not start by default.

```
+----------+-----------------------------------+----------------------------+
| Service  | What it runs                      | Lifecycle                  |
+----------+-----------------------------------+----------------------------+
| cli      | main -action ingest|sample        | on-demand (compose run)    |
| postgres | postgres:18.4                     | always up, healthchecked   |
| ui       | main -action serve                | always up, depends on db   |
+----------+-----------------------------------+----------------------------+
```

The CLI actions:

```bash
make ingest                                          # import replays
docker compose run --rm cli -action sample -file /path/to/some.SC2Replay   # pretty-print parsed JSON
```

## Make targets

```
make build    # docker compose build
make up        # build and start postgres + ui (foreground, streams logs)
make down      # stop and remove containers
make ingest    # one-shot replay import
make lint      # golangci-lint + biome, all inside a one-off container
make sqlc      # regenerate models/ (runs sqlc via docker compose run)
make reset     # re-apply sqlc/schema.sql to the running postgres (WARNING: drops all data)
make clean     # stop the stack and delete the ./postgres data directory
```

## Reset the database

`make reset` runs `sqlc/schema.sql` against the running `postgres` container. The schema starts with `DROP SCHEMA public CASCADE`, so this wipes all games and saved chats:

```bash
make reset
make ingest
```

To start completely fresh, run `make clean` (stops the stack and removes the bind-mounted `./postgres` directory), then `make up`.

## License

See LICENSE file for details.
