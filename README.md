# LifeCalendar

A minimal, simple Go CLI tool for generating life events calendars with color-coded days and plan tracking.

<img width="2537" height="1179" alt="screenshot" src="https://github.com/user-attachments/assets/4bc9cad7-3950-47c1-a5e1-9dac8e980c9c" />

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

### Data Organization

Calendar data is organized by year in the `data/` directory:

```
data/
├── 2025/
│   ├── public_holidays.csv
│   ├── vacations.csv
│   └── plans.csv
└── 2026/
    ├── public_holidays.csv
    ├── vacations.csv
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
- docker for linting
