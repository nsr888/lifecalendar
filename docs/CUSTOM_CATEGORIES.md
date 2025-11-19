# Custom Categories Feature

This document explains how to create and use custom categories in the life calendar application.

## Overview

The life calendar now supports unlimited custom categories beyond the predefined ones (vacations, personal days, etc.). Custom categories are automatically discovered from CSV files and can be styled individually. The calendar displays day numbers with colored backgrounds instead of icons.

## Creating Custom Categories

### 1. Create a CSV File

Create a new CSV file in the `data/{year}/` directory. The filename becomes the category name:

```
data/2025/team_events.csv
data/2025/training.csv
data/2025/conferences.csv
```

### 2. CSV Format

All categories use the same unified format:

```csv
date_start,date_end,label
2025-01-15,2025-01-15,Team Building Workshop
2025-06-10,2025-06-12,Company Offsite
2025-11-28,2025-11-28,Awards Ceremony
```

**Supported Formats:**

- **Single dates**: `date_start` and `date_end` are the same
- **Date ranges**: Different `date_start` and `date_end` values
- **Labels**: Optional descriptive text for each entry

**Examples:**

```csv
# Single day events
date_start,date_end,label
2025-03-15,2025-03-15,Workshop

# Multi-day events
date_start,date_end,label
2025-07-01,2025-07-05,Conference

# Without labels (optional)
date_start,date_end
2025-09-10,2025-09-12
```

## Styling Custom Categories

Add styling configuration to `config.toml`:

```toml
[categories]
[categories.team_events]
bg = "#FF6B6B"
fg = "#ffffff"
bold = true
priority = 3

[categories.training]
bg = "#4ECDC4"
fg = "#ffffff"
priority = 4

[categories.conferences]
bg = "#9B59B6"
fg = "#ffffff"
italic = true
priority = 6
```

### Configuration Options

- **bg**: Background color (hex code)
- **fg**: Foreground color (hex code)
- **bold**: Make text bold (true/false)
- **italic**: Make text italic (true/false)
- **priority**: Display priority (lower numbers = higher priority)

### Priority System

Categories with lower priority numbers override higher priority categories when dates overlap:

```toml
priority = 1  # Highest priority (shows on top)
priority = 2  # High priority
priority = 3  # Medium priority
priority = 99 # Low priority
```

### Auto-Generated Colors

If a category exists in the data folder but has no configuration in `config.toml`, it will automatically receive:

- A randomly generated background color based on the category name hash
- White foreground color (`#ffffff`) for good contrast
- Priority of 999 (low priority)
- No special styling (no bold/italic)

The color generation ensures that the same category name always gets the same color across different years.

## CSV Format

All categories use the same unified format:

```csv
date_start,date_end,label
2025-01-15,2025-01-15,Team Building Workshop
2025-06-10,2025-06-12,Company Offsite
```

**Label is optional** - you can omit the label column:

```csv
date_start,date_end
2025-01-15,2025-01-15
2025-06-10,2025-06-12
```

**Supported Headers:**

- `date_start,date_end,label` - Full format with optional labels
- `date_start,date_end` - Minimal format without labels
- `date` - Single date format (treated as date_start=date_end)

**Core Categories:**

- `public_holidays.csv` - Public holidays
- `vacations.csv` - Vacation periods
- `personal_days.csv` - Personal days off
- `plans.csv` - Planned trips/activities
- `weekends.csv` - Auto-generated weekends

All core categories use the same unified format.

## Examples

### Team Events Category

**File: `data/2025/team_events.csv`** (with labels)

```csv
date_start,date_end,label
2025-01-15,2025-01-15,Team Building Workshop
2025-06-10,2025-06-12,Company Offsite
2025-11-28,2025-11-28,Awards Ceremony
```

**File: `data/2025/public_holidays.csv`** (without labels)

```csv
date_start,date_end
2025-12-24,2025-12-24
2025-12-25,2025-12-25
2025-10-01,2025-10-01
```

**Config:**

```toml
[categories.team_events]
bg = "#FF6B6B"
fg = "#ffffff"
bold = true
priority = 3

[categories.public_holidays]
bg = "#7a2936"
fg = "#ffffff"
priority = 1
```

### Training Category

**File: `data/2025/training.csv`**

```csv
date_start,date_end,label
2025-02-20,2025-02-21,Go Programming Course
2025-04-15,2025-04-17,Project Management Training
2025-08-05,2025-08-07,Leadership Workshop
```

**Config:**

```toml
[categories.training]
bg = "#4ECDC4"
fg = "#ffffff"
priority = 4
```

### Auto-Generated Category (No Config)

**File: `data/2025/meetings.csv`**

```csv
date_start,date_end,label
2025-03-10,2025-03-10,Team Meeting
2025-07-15,2025-07-15,Client Call
2025-09-05,2025-09-05,All Hands
```

This category will automatically get a generated color since no configuration is provided in `config.toml`.

## Tips

1. **Category names** are derived from filenames (without `.csv` extension)
2. **Labels are optional** - use `date_start,date_end` or `date_start,date_end,label`
3. **Colors** use hex codes (`#RRGGBB`) format, or get auto-generated
4. **Day numbers** are always displayed with colored backgrounds
5. **Priorities** help resolve conflicts when dates overlap
6. **Labels** are for reference only and don't appear in the calendar grid
7. **CSV files** are automatically discovered - no code changes needed
8. **Configuration is optional** - unconfigured categories get auto-generated colors

## Troubleshooting

### Category Not Loading

- Ensure CSV file is in the correct `data/{year}/` directory
- Check CSV format has correct headers (`date_start,date_end,label`)
- Verify dates are in `YYYY-MM-DD` format

### Styling Not Applied

- Check configuration in `config.toml` under `[categories.category_name]`
- Ensure colors are valid hex codes (`#RRGGBB`)
- Verify category name matches filename (case-sensitive)
- If no configuration is provided, the category will use auto-generated colors

### Color Issues

- Auto-generated colors provide good contrast with white text
- Manually specified colors should ensure good readability
- Use colors with sufficient contrast for accessibility
