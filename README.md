# StarCraft II Replay Analyzer

Go-based parser and analyzer for StarCraft II replay files (.SC2Replay). Extracts comprehensive game data and stores it in PostgreSQL for analysis.

## Features

- **Parallel Processing**: 12 concurrent workers for fast replay parsing
- **Comprehensive Data Extraction**:
  - Game metadata (map, mode, duration, timestamp)
  - Player statistics (MMR, APM, race, color, region)
  - Per-tick resource tracking (minerals, vespene, supply, army value)
  - Chat messages with timestamps
  - Unit births and deaths
  - Team results
- **Real-time TUI**: Bubble Tea progress display with ETA
- **Type-safe Database**: sqlc-generated PostgreSQL operations
- **Batch Operations**: Efficient multi-row inserts

## Requirements

- Go 1.23+
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

### Database Setup

```bash
# Copy sample environment file
cp .env.sample .env

# Edit .env with your PostgreSQL connection
# Default: postgres://postgres:postgres@0.0.0.0:5432/starcraft2?sslmode=disable
```

### Initialize Database

```bash
# Create schema
psql -h localhost -U postgres -d starcraft2 < sqlc/schema.sql

# Or use the reset script
./reset.sh
```

### Replay File Paths

Edit `variables.go` to configure paths where replay files are located:

```go
var Paths = []string{
    "/path/to/your/replays",
}
```

## Usage

### Parse All Replays

```bash
./starcraft2 --action=refresh
```

Processes all .SC2Replay files from configured paths:
- Spawns 12 concurrent workers
- Displays real-time progress with TUI
- Stores data in PostgreSQL

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

### Generate Database Code

```bash
# After modifying sqlc/schema.sql or sqlc/queries.sql
sqlc generate

# Or use the script
./sqlc.sh
```

### Live Reload

```bash
# Watches .go files and auto-rebuilds
air
```

### Code Quality

```bash
# Run linter
golangci-lint run
```

## Dependencies

- [icza/s2prot](https://github.com/icza/s2prot) - StarCraft II protocol parser
- [bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [pgx/v5](https://github.com/jackc/pgx) - PostgreSQL driver
- [sqlc](https://sqlc.dev/) - Type-safe SQL code generation
- [go-pretty](https://github.com/jedib0t/go-pretty) - Table formatting

## Architecture

### Data Flow

1. **Discovery**: `buildFiles()` walks configured paths for .SC2Replay files
2. **Distribution**: Files distributed to 12 worker goroutines via channel
3. **Parsing**: Each worker uses `icza/s2prot` to parse replay binary format
4. **Extraction**: `buildGame()` extracts all game entities and statistics
5. **Storage**: `upsert()` stores data via sqlc-generated methods
6. **Progress**: Bubble Tea TUI shows real-time worker status

### Concurrency Model

- Main thread: File discovery + TUI rendering
- 12 workers: Parallel replay parsing
- Channel-based work distribution
- Progress updates via Bubble Tea messages

## License

See LICENSE file for details.
