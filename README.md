# StarCraft II Replay Analyzer

Parses StarCraft II replay files (.SC2Replay) into PostgreSQL and serves a natural-language Q&A web page over the data. Ask a question in plain English; an AI agent writes a read-only SQL query, runs it, and answers from the rows it gets back.

Everything runs in Docker: the Go binary, the Rust sidecar (`sc2json`), and `sqlc` all run inside one image against the repo mounted at `/sources`. There is no host build or run path. See CONTEXT.md for how it works internally, and DEPLOY.md to install it on a server behind nginx.

## Requirements

- Docker Engine with the Docker Compose plugin.
- An OpenRouter API key.

## Order of operations

```bash
cp .env.sample .env     # 1. copy the sample env, then fill in every variable (see Configuration)
make build              # 2. build the image (Go binary + Rust sidecar + tooling)
make ingest             # 3. import replays - this starts postgres for you, then exits
make up                 # 4. start the web UI (runs in the foreground; Ctrl-C to stop)
```

Then open http://localhost:8080 and ask a question.

That is the whole flow. A few things worth knowing:

- **You do not have to start anything before `make ingest`.** The `cli` service depends on `postgres` with a health check, so step 3 starts postgres, waits until it is ready, imports every replay under `GO_REPLAYS`, and the one-shot container exits. postgres stays running. Re-running is cheap - already-imported paths are skipped.
- **`make up` stays in the foreground** and streams logs, so run it in its own terminal. With `GO_ENVIRONMENT=development` (the sample default) the server runs under [air](https://github.com/air-verse/air) and rebuilds automatically on any `.go`, `.sql`, or frontend change.
- **Order is flexible.** If you want the UI running while you import, start `make up` in one terminal and `make ingest` in another; the two do not depend on each other beyond postgres.
- The UI is published on `GO_PORT` on all interfaces and has **no authentication** - keep it behind a reverse proxy or firewall in any shared environment (see DEPLOY.md).

## Configuration

All configuration lives in `.env`, which the Go process loads directly. The variable names in `.env` are exactly the names used in the code, so there is only one name per setting. Every variable is required and there are no defaults: if any is missing, the program prints which ones and exits. Copy `.env.sample` to `.env` and fill in all of them.

```
+--------------------+----------------------------------------------------------+
| Variable           | Purpose                                                  |
+--------------------+----------------------------------------------------------+
| GO_ENVIRONMENT     | development (air hot reload) or production (build once). |
| GO_PLAYERS         | Tracked player names, comma-separated.                   |
| GO_PORT            | Port the UI listens on and is published at.              |
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

`GO_REPLAYS` is bind-mounted into the `cli` container at the same path it has on the host, so the path the app walks and the path you set are identical. The Postgres connection string is built from the `POSTGRES_*` values, so the password is never duplicated.

## Services

```
+----------+----------------------------+----------------------------+
| Service  | What it runs               | Lifecycle                  |
+----------+----------------------------+----------------------------+
| postgres | postgres:18.4              | always up, health checked  |
| ui       | main -action serve         | always up, depends on db   |
| cli      | main -action ingest|sample | on-demand (compose run)    |
+----------+----------------------------+----------------------------+
```

`make up` starts `postgres` and `ui`. `cli` is the same image on a `cli` profile, so it does not start with `make up`; you invoke it on demand (`make ingest`, or the sample below).

Inspect one replay's parsed JSON without importing it:

```bash
docker compose run --rm cli -action sample -file /path/to/replay.SC2Replay
```

## Make targets

```
make build    build the image (Go binary + Rust sidecar + tooling)
make clean    stop the stack and delete the ./postgres data directory
make down     stop and remove the containers
make ingest   import replays (one-shot; starts postgres if needed)
make lint     run golangci-lint + biome in a one-off container
make reset    re-apply sqlc/schema.sql to postgres (drops all data)
make sqlc     regenerate the models/ package from sqlc/
make up       build and start postgres + ui (foreground, streams logs)
```

## Database

`make reset` re-applies `sqlc/schema.sql` to the running postgres. The schema begins with `DROP SCHEMA public CASCADE`, so it wipes all games and saved chats; re-import afterward:

```bash
make reset
make ingest
```

To start completely fresh, `make clean` stops the stack and deletes the bind-mounted `./postgres` directory; then run the order of operations again from `make build`.

## License

See LICENSE file for details.
