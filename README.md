# StarCraft II Replay Analyzer

Go-based parser and analyzer for StarCraft II replay files (.SC2Replay). Extracts comprehensive game data and stores it in PostgreSQL for analysis.

## Features

- **Parallel Processing**: Configurable concurrent workers for fast replay parsing
- **Environment-based Configuration**: All settings via .env file
- **Graceful Shutdown**: Context-based cancellation with signal handling
- **Atomic Transactions**: Database operations wrapped in transactions for consistency
- **Comprehensive Data Extraction**:
  - Game metadata (map, mode, duration, timestamp)
  - Player statistics (MMR, APM, race, color, region)
  - Per-tick resource tracking (minerals, vespene, supply, army value)
  - Chat messages with timestamps
  - Unit births and deaths
  - Team results with player filtering
- **Real-time TUI**: Bubble Tea progress display with ETA
- **Type-safe Database**: sqlc-generated PostgreSQL operations
- **Batch Operations**: Efficient multi-row inserts
- **Dependency Injection**: Clean architecture with explicit dependencies

## Requirements

- Go 1.25+
- PostgreSQL
- StarCraft II replay files

## Installation

```bash
# Clone repository
git clone <repository-url>
cd starcraft2

# Install dependencies
go mod download

# Build binary
go build
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
Comma-separated replay directory paths. Use quotes if paths contain spaces.
```
REPLAYS="/home/user/StarCraft II/Replays,/mnt/replays"
```

**WORKERS** (optional)
Number of concurrent worker goroutines. Defaults to number of CPUs.
```
WORKERS=12
```

**PLAYERS** (optional)
Comma-separated player names for result determination in undecided games. Empty = all players.
```
PLAYERS=PlayerOne,PlayerTwo
```

### Database Setup

```bash
# Create schema
psql -h localhost -U postgres -d starcraft2 < sqlc/schema.sql
```

## Usage

### Parse All Replays

```bash
./starcraft2 --action=refresh
```

Processes all .SC2Replay files from configured paths:
- Spawns configurable concurrent workers (WORKERS env var)
- Displays real-time progress with TUI
- Stores data in PostgreSQL with atomic transactions

### Parse Single Replay (Sample)

```bash
./starcraft2 --action=sample
```

Parses one replay file and outputs JSON to stdout.

### View Statistics

```bash
./starcraft2 --action=statistics
```

Displays row counts for all database tables.

## Database Schema

### Tables

- **games**: Game metadata (map, mode, duration, timestamp)
- **teams**: Team results and outcomes
- **players**: Player information (name, MMR, APM, race, color)
- **messages**: In-game chat messages
- **stats**: Per-tick player statistics (40+ metrics including resources, army composition)
- **units**: Unit birth/death events

### Key Metrics Tracked

Player stats include:
- Resources: minerals, vespene, minerals collection rate
- Supply: food used/made, army/worker counts
- Economy: workers active, mineral/vespene cost collections
- Military: army value, unit counts by type
- APM and other performance metrics

## Development

### Build

```bash
make build
```

### Lint

```bash
make lint
```

### Generate Database Code

After modifying `sqlc/schema.sql` or `sqlc/queries.sql`:

```bash
sqlc generate
```

## Dependencies

- [icza/s2prot](https://github.com/icza/s2prot) - StarCraft II protocol parser
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [pgx/v5](https://github.com/jackc/pgx) - PostgreSQL driver
- [sqlc](https://sqlc.dev/) - Type-safe SQL code generation
- [go-pretty](https://github.com/jedib0t/go-pretty) - Table formatting

## Architecture

### Application Structure

- **Dependency Injection**: Application struct holds all dependencies (DB pool, queries, settings)
- **Environment Configuration**: Settings loaded from .env on startup
- **Graceful Shutdown**: Signal handling with context cancellation
- **Transaction Safety**: All database inserts wrapped in transactions for atomicity

### Data Flow

1. **Initialization**: `NewApplication()` loads config and creates DB connection
2. **Discovery**: `buildFiles()` walks configured paths for .SC2Replay files
3. **Distribution**: Files distributed to worker goroutines via channel
4. **Parsing**: Each worker uses `icza/s2prot` to parse replay binary format
5. **Extraction**: `buildGame()` extracts all game entities and statistics
6. **Storage**: `upsert()` stores data via sqlc-generated methods in atomic transactions
7. **Progress**: Bubble Tea TUI shows real-time worker status
8. **Shutdown**: `Application.Close()` releases DB connection pool

### Concurrency Model

- Main thread: Application initialization + file discovery + TUI rendering
- Configurable workers: Parallel replay parsing (default: CPU count)
- Channel-based work distribution with WaitGroup synchronization
- Context-based cancellation for graceful shutdown
- Progress updates via Bubble Tea messages

## License

See LICENSE file for details.
