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

# Required
[[events]]
date = "2020-03-15"
title = "Company Founded"
description = "Optional description"
```

## Examples

See `example.toml` for a complete configuration example.

## License

MIT