# VanityCal

Generate iCalendar files with anniversary events for important dates.

## Installation

```bash
go install moul.io/vanitycal@latest
```

## Usage

```bash
vanitycal -config events.toml -output calendar.ics
cat events.toml | vanitycal > calendar.ics
```

## Configuration

```toml
# Optional
timezone = "Europe/Paris"
calendar_name = "VanityCal 💚"

[anniversaries]
years = [1, 2, 3, 4, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50]
months = [1, 2, 3, 6, 9]
days = [0, 7, 100, 1000, 10000]  # 0 = D-Day

# Events (use either date OR month_day, not both)

# Anniversary events (past dates)
[[events]]
date = "2020-03-15"
title = "Company Founded"
description = "Optional description"

# Future events (auto-detected, generates both countdown AND anniversaries)
[[events]]
date = "2025-06-20"
title = "Product Launch"  # Generates D-7, D-30, etc. AND 7d, 1y, etc.

# Event filtering options
[[events]]
date = "2025-12-31"
title = "Countdown Only"
no_past = true  # Only countdown events (D-7, D-30), no anniversaries

[[events]]
date = "2020-01-01"
title = "Anniversary Only"
no_future = true  # Only past anniversaries (7d, 1y), no future events

# Recurring annual events (no year)
[[events]]
month_day = "12-25"
title = "Christmas"

[[events]]
month_day = "07-04"
title = "Independence Day"
```

## Examples

See `example.toml` for a complete configuration example.

## License

MIT