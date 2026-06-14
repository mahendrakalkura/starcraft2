# StarCraft II Replay Analyzer

Parses StarCraft II replay files (.SC2Replay) into PostgreSQL, runs analysis reports, and serves a natural-language Q&A web page over the data. Two binaries: `sc2json` (Rust sidecar that turns one replay into compact JSON) and `main` (Go CLI that ingests replays, runs reports, and hosts the web UI).

For how the system works internally (architecture, the agent loop, the data model), see CONTEXT.md.

## Requirements

- Go 1.26+
- PostgreSQL 18+
- Rust (cargo)
- sqlc (only when changing the schema or queries)

To run everything in containers instead, you only need Docker and Docker Compose (see Docker below).

## Build

```bash
make build
```

This builds the Rust sidecar in release mode, copies `sc2json` next to the Go binary, and builds `main`.

## Configuration

Copy `.env.sample` to `.env` and edit. A `.env` file is optional if the variables are set in the environment.

```
DATABASE=postgres://user:password@host:5432/starcraft2?sslmode=disable
OPENROUTER_API_KEY=sk-or-...
OPENROUTER_MODEL=deepseek/deepseek-chat
PORT=8080
PLAYERS=PlayerA,PlayerB,PlayerC
REPLAYS="/path/to/StarCraft II/Accounts/"
WORKERS=16
```

- `DATABASE` (required): PostgreSQL connection string.
- `OPENROUTER_API_KEY` (required for `serve`): OpenRouter key for the web UI.
- `OPENROUTER_MODEL` (optional): any tool-calling model, defaults to `deepseek/deepseek-chat`.
- `PLAYERS` (optional): tracked player names; the first is the default `-me` for reports. Also used to resolve games where every result is "Undecided" (the recording player left early).
- `PORT` (optional): web server port for `serve`, defaults to `8080`.
- `REPLAYS` (required for `ingest`): comma-separated directories, walked recursively for .SC2Replay files.
- `SC2JSON` (optional): path to the sidecar binary, defaults to `sc2json` next to the `main` executable.
- `WORKERS` (optional): worker count, defaults to CPU count.

Initialize the database (WARNING: drops all data):

```bash
make reset
```

## Usage

```bash
./main                      # ingest (default action)
./main -action ingest -force    # wipe imported data and reprocess everything
./main -action serve            # start the natural-language Q&A web UI
./main -action statistics       # row counts per table
./main -action sample -file path/to/replay.SC2Replay   # pretty-print sidecar JSON
```

Each ingested file ends in one state: `imported`, `skipped_ai` (had a Computer player), `duplicate`, or `failed`. Already-processed paths are skipped on re-run; delete a `files` row to retry it. See CONTEXT.md for the full pipeline.

Analysis reports (`-me` defaults to the first `PLAYERS` entry):

```bash
./main -action chat                              # most-typed phrases and gg rate by result
./main -action durations                         # game-length bell curve with win rates
./main -action economy                           # avg income/workers at 5/10/15 min, wins vs losses
./main -action impact -name PlayerB -with PlayerC    # tracked-team win rate with player(s) as ally vs opponent
./main -action maps                              # per-map record with recent form and current-pool marker
./main -action matchup                           # win rate by my race x mode and by opposing race mix
./main -action maxout                            # maxed-supply win rates by Carrier count + loss breakdown
./main -action mmr-history                       # monthly games/MMR/win rate plus per-race summary
./main -action openings                          # cannon vs stalker openings: overall, year, timing, map
./main -action partners                          # win rate per tracked-team lineup
./main -action report-1                          # match list: winners, losers, map, mode, duration
./main -action report-2                          # daily win/loss bars for tracked PLAYERS
./main -action rivals                            # most-faced opponents with MMR diff and anomaly notes
./main -action streaks                           # win rate after win/loss streaks, longest streaks
./main -action versus -name Opponent            # head-to-head: record, by year, race, map, last 10
```

## Web UI

`./main -action serve` starts a public, no-auth web page where anyone can ask questions about the database in plain English. An AI agent (DeepSeek via OpenRouter) generates a read-only SQL query, runs it, and writes a markdown answer. Set `OPENROUTER_API_KEY` (and optionally `OPENROUTER_MODEL`). Listens on `PORT` (default 8080). Chats are saved and listed in a sidebar shared by everyone.

This is a wide-open toy: no authentication, no rate limit, no usage cap. Anyone who can reach the URL can run paid AI calls and read the game data. Run it only where that is acceptable (for example a host reached by a private IP).

## Docker

Everything runs as Docker Compose containers: `db` (Postgres), `web` (the serve UI), and `cli` (the same binary for one-off ingest and reports, on a `cli` profile so it does not start by default).

```bash
cp .env.sample .env       # set OPENROUTER_API_KEY, POSTGRES_PASSWORD, REPLAYS_HOST
make up                   # build and start db + web (http://<host>:8080)
make ingest               # docker compose run --rm cli -action ingest
docker compose run --rm cli -action versus -name Opponent
```

On first boot the db container initializes the game schema (`sqlc/schema.sql`) and the chat schema (`sqlc/chat_schema.sql`). Both `conversations`/`turns` live in `public`, so re-importing replays clears saved chats; recreate them after a reset:

```bash
docker compose exec -T db psql -U postgres -d starcraft2 < sqlc/schema.sql
docker compose exec -T db psql -U postgres -d starcraft2 < sqlc/chat_schema.sql
docker compose run --rm cli -action ingest -force
```

## Development

```bash
make sqlc    # regenerate models after editing the schema or query files
```

## License

See LICENSE file for details.
