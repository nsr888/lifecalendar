# LifeCalendar

A minimal, simple Go CLI tool for generating life events calendars with color-coded days and plan tracking.

## Features

- **Calendar Generation**: Month-by-month calendars for specified years
- **Year Configuration**: Render one or more years via config.toml
- **Color Coding**: Holidays (red), Vacations (green), Personal Days (blue), Weekends (gray)
- **Plan Management**: Travel plans with automatic color generation
- **Compact Display**: 3×4 month grid with ANSI colors
- **TOML Configuration**: Flexible date definitions and categories

## Project Structure

```
.
├── cmd/
│   └── main.go          # Entry point (minimal CLI)
├── internal/
│   ├── config/          # TOML configuration loading
│   ├── calendar/        # Core calendar business logic
│   ├── render/          # Terminal rendering
│   ├── storage/         # CSV data storage
│   └── entity/          # Shared types
├── pkg/
│   └── colors/          # Color scheme management
├── config.toml          # Configuration file
├── Makefile             # Build targets
└── go.mod              # Go module definition
```

## Building and Running

```bash
# Build the project
make build

# Run the application
make run

# Format code
make fmt

# Lint code (requires golangci-lint)
make lint

# Clean build artifacts
make clean
```

## Configuration

Edit `config.toml` to customize years, categories, and colors:

### Year Configuration

```toml
# Years to render calendars for (optional - defaults to current year if not specified)
years = [2025, 2026]     # Multiple years (rendered separately)
# years = [2025]          # Single year
# years = []              # Empty array will use current year
```

### Full Configuration Example

```toml
# Years to render
years = [2025, 2026]


# Rendering settings
[rendering]
first_weekday = 0        # Monday = 0, Sunday = 6
weekend_days = [5, 6]    # Saturday = 5, Sunday = 6

# Category colors (can be overridden in CSV files)
[colors]
public_holidays = "darkred"
vacations = "darkgreen"
personal_days = "darkblue"
```

### Data Organization

Calendar data is organized by year in the `data/` directory:

```
data/
├── 2025/
│   ├── public_holidays.csv
│   ├── vacations.csv
│   ├── personal_days.csv
│   └── plans.csv
└── 2026/
    ├── public_holidays.csv
    ├── vacations.csv
    ├── personal_days.csv
    └── plans.csv
```

Each year you configure must have a corresponding directory with the required CSV files.

## Architecture

The codebase follows Go's simplicity principles:

- **Single Responsibility**: Each package has one clear purpose
- **Explicit Dependencies**: No hidden constructors or singletons
- **Pure Functions**: Business logic separated from I/O
- **Minimal Complexity**: Simple types and focused functions

### Package Overview

- `cmd/main.go`: Application entry point and configuration wiring
- `internal/config`: TOML parsing and configuration loading
- `internal/calendar`: Core calendar calculations and date logic
- `internal/render`: Terminal output formatting and ANSI colors
- `internal/storage`: CSV data loading and validation
- `internal/entity`: Shared types and data structures
- `pkg/colors`: Color generation and ANSI escape codes

## Requirements

- Go 1.24.2+
- `golangci-lint` for linting (optional)
