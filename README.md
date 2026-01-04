# StarCraft II Replay Analyzer

![SPDX](https://img.shields.io/badge/SPDX-MIT-blue.svg)
![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)

Go-based parser and analyzer for StarCraft II replay files (.SC2Replay). Extracts comprehensive game data and stores it in PostgreSQL for analysis.

## Features

- **Parallel Processing**: Configurable concurrent workers for fast replay parsing
- **Environment-based Configuration**: All settings via .env file
- **Graceful Shutdown**: Context-based cancellation with signal handling (SIGINT/SIGTERM)
- **Atomic Transactions**: Database operations wrapped in transactions for consistency
- **Comprehensive Data Extraction**:
  - Game metadata (map, mode, duration, timestamp, type)
  - Player statistics (MMR, APM, assigned/selected race, color, control type)
  - Per-tick resource tracking (minerals, vespene, supply, workers)
  - Per-tick economy metrics (collection rates, friendly fire, units killed/lost)
  - Chat messages with timestamps and recipient targeting
  - Unit birth/death events with coordinates
  - Team results with configurable player filtering
- **Real-time TUI**: Bubble Tea progress display with worker status, completion percentage, and ETA
- **Type-safe Database**: sqlc-generated PostgreSQL operations
- **Batch Operations**: Efficient multi-row inserts for messages, stats, and units
- **Dependency Injection**: Clean architecture with explicit dependencies

## Requirements

- Go 1.25+
- PostgreSQL 12+
- StarCraft II replay files (.SC2Replay)

## Installation

```bash
# Clone repository
git clone <repository-url>
cd starcraft2

# Install dependencies and build
make build
```

## Configuration

### Environment Variables

Create `.env` file in project root:

```bash
cp .env.sample .env
```

Edit `.env` with your configuration:

**DATABASE** (required)
PostgreSQL connection string
```
DATABASE=postgres://postgres@localhost:5432/starcraft2
```

**REPLAYS** (required)
Comma-separated replay directory paths. Paths with spaces must be quoted.
```
REPLAYS="/home/user/StarCraft II/Replays,/mnt/replays"
```

**WORKERS** (optional)
Number of concurrent worker goroutines. Defaults to `runtime.NumCPU()`.
```
WORKERS=12
```

**PLAYERS** (optional)
Comma-separated player names for result determination in undecided games.
When both teams have "Undecided" result, if a player in this list is found, their team is marked as "Loss" and opponent as "Win".
Leave empty to disable filtering.
```
PLAYERS=PlayerOne,PlayerTwo,PlayerThree
```

### Database Setup

Initialize the database schema:

```bash
# Option 1: Direct psql
psql -h localhost -U postgres -d starcraft2 < sqlc/schema.sql

# Option 2: Using Makefile (WARNING: drops all data)
make reset
```

**Note**: Both methods execute `DROP SCHEMA public CASCADE` which deletes all existing data.

## Usage

### Parse All Replays

```bash
./starcraft2 --action=refresh

# Or using Makefile
make refresh
```

Processes all .SC2Replay files from configured paths:
- Spawns workers (count from WORKERS env var or CPU count)
- Deletes existing game record before inserting (by file path)
- Displays real-time TUI with worker status, progress, and ETA
- Stores all data in PostgreSQL using atomic transactions
- Press `q` or `Ctrl+C` to quit

### Parse Single Replay (Debug)

```bash
./starcraft2 --action=sample
```

Parses the hardcoded replay file in `sample.go` and outputs:
- `r.json`: Raw replay data structure
- `game.json`: Extracted game data structure

Useful for debugging replay parsing logic.

### View Statistics

```bash
./starcraft2 --action=statistics
```

Displays row counts for all database tables in a formatted table.

## Database Schema

### Tables

**games**
Game-level metadata
- `id`: Primary key
- `file`: Unique file path
- `duration`: Game duration in seconds
- `map`: Map name
- `mode`: Game mode (e.g., "1v1", "2v2")
- `timestamp`: Game start time (UTC)
- `type`: Game type from replay attributes

**teams**
Team results (multiple teams per game)
- `id`: Primary key
- `game_id`: Foreign key to games
- `number`: Team number (1-based)
- `result`: Result ("Win", "Loss", "Undecided", "Tie")

**players**
Player information (multiple players per team)
- `id`: Primary key
- `team_id`: Foreign key to teams
- `number`: Player slot number
- `apm`: Actions per minute
- `color`: Player color (hex format)
- `control`: Control type ("Human", "Computer")
- `mmr`: Matchmaking rating
- `name`: Player name
- `observe`: Observer status ("Participant", "Observer")
- `races_assigned`: Assigned race
- `races_selected`: Selected race

**messages**
Chat messages (multiple per player)
- `id`: Primary key
- `player_id`: Foreign key to players
- `time`: Game time (loop count)
- `recipient_id`: Message target (0=All, 2=Allies, 4=Observers)
- `string`: Message text

**stats**
Per-tick player statistics (multiple per player, sampled periodically)
- `id`: Primary key
- `player_id`: Foreign key to players
- `time`: Game time (loop count)
- `food_made`, `food_used`: Supply statistics
- `minerals_*`, `vespene_*`: 36+ resource metrics including:
  - Current resources and collection rates
  - Resources used for army/economy/technology (current and in-progress)
  - Resources from kills (army/economy/technology)
  - Resources lost (army/economy/technology)
  - Friendly fire resources (army/economy/technology)
- `workers_active_count`: Active worker count

**units**
Unit birth/death events (multiple per player)
- `id`: Primary key
- `player_id`: Foreign key to players
- `time`: Game time (loop count)
- `action`: Event type ("UnitBorn", "UnitDied")
- `name`: Unit type name
- `x`, `y`: Map coordinates

### Relationships

```
games (1) → (N) teams (1) → (N) players (1) → (N) messages/stats/units
```

## Development

### Build

Formats code, tidies dependencies, and builds binary:

```bash
make build
```

### Lint

Runs golangci-lint:

```bash
make lint
```

### Run Application

```bash
make refresh    # Run --action=refresh
```

### Reset Database

**WARNING**: Drops all data and recreates schema:

```bash
make reset
```

### Generate Database Code

After modifying `sqlc/schema.sql` or `sqlc/queries.sql`:

```bash
make sqlc
# or
sqlc generate
```

## Dependencies

- [icza/s2prot](https://github.com/icza/s2prot) - StarCraft II replay protocol parser
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [jackc/pgx/v5](https://github.com/jackc/pgx) - PostgreSQL driver and toolkit
- [sqlc](https://sqlc.dev/) - Type-safe SQL code generation
- [jedib0t/go-pretty/v6](https://github.com/jedib0t/go-pretty) - Table rendering
- [joho/godotenv](https://github.com/joho/godotenv) - .env file loader
- [samber/lo](https://github.com/samber/lo) - Generic utility functions
- [spf13/cast](https://github.com/spf13/cast) - Type conversion utilities
- [dustin/go-humanize](https://github.com/dustin/go-humanize) - Number formatting

## Architecture

### Application Structure

**Dependency Injection**
- `Application` struct holds all dependencies (DB pool, queries, settings)
- Created via `NewApplication()` which loads .env and initializes DB connection
- Cleaned up via `Application.Close()` which releases DB pool

**Environment Configuration**
- `Settings` struct loaded from .env on startup via `NewSettings()`
- Panics if required variables (DATABASE, REPLAYS) are missing
- Optional variables have sensible defaults (WORKERS=CPU count, PLAYERS=empty)

**Graceful Shutdown**
- Signal handling for SIGINT and SIGTERM via `signal.NotifyContext()`
- Context propagation throughout application (currently unused but prepared)
- Deferred `Application.Close()` ensures cleanup on exit

**Transaction Safety**
- All database inserts wrapped in `pgx` transactions
- Games, teams, players, messages, stats, and units inserted atomically
- Rollback on any error prevents partial/orphaned records

### Data Flow

1. **Initialization**
   `NewApplication()` → load .env → create DB pool → initialize queries

2. **Discovery**
   `buildFiles()` → walk REPLAYS paths → find .SC2Replay files → sort by modification time

3. **Distribution**
   Files pushed to channel → workers consume via `range` loop

4. **Parsing**
   `icza/s2prot` → parse binary replay → extract protocol buffer data

5. **Extraction**
   `buildGame()` → transform replay data → create Game struct with teams/players/messages/stats/units

6. **Storage**
   `upsert()` → begin transaction → delete existing game → insert game/teams/players/messages/stats/units → commit

7. **Progress**
   Workers send `Channel` messages → Bubble Tea updates TUI → display completion/ETA

8. **Shutdown**
   All workers finish → channels closed → TUI exits → `Application.Close()` releases pool

### Concurrency Model

**Main Goroutine**
- Application initialization
- File discovery
- Bubble Tea TUI rendering

**Worker Goroutines** (configurable count)
- Consume files from input channel
- Parse replay with `icza/s2prot`
- Extract data via `buildGame()`
- Insert via `upsert()`
- Send progress via output channel

**Synchronization**
- `sync.WaitGroup.Go()` for worker spawning (Go 1.25+)
- Channel-based work distribution
- Sequential channel closure after `WaitGroup.Wait()`
- Bubble Tea message passing for progress updates

**Player Filtering**
When both teams have "Undecided" result:
- `buildTeams()` checks each player name against `Settings.Players`
- If match found, that team → "Loss", opponent → "Win"
- Useful for marking losses when leaving early (before official result)

## License

See LICENSE file for details.
