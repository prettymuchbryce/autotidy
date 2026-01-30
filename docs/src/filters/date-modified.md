# date_modified

Matches files by their last modification time.

## Syntax

```yaml
# Files modified before a relative time
- date_modified:
    before:
      days_ago: 30

# Files modified after a relative time
- date_modified:
    after:
      hours_ago: 24

# Files modified before an absolute date
- date_modified:
    before:
      date: "2024-01-01"
```

## Operators

| operator | description |
|----------|-------------|
| `before` | File was modified before this time |
| `after` | File was modified after this time |

## Time specifications

### Relative time

| option | description |
|--------|-------------|
| `seconds_ago` | Seconds in the past |
| `minutes_ago` | Minutes in the past |
| `hours_ago` | Hours in the past |
| `days_ago` | Days in the past |
| `weeks_ago` | Weeks in the past |
| `months_ago` | Months in the past (~30.44 days) |
| `years_ago` | Years in the past (~365.25 days) |

### Absolute time

| option | format | example |
|--------|--------|---------|
| `date` | `YYYY-MM-DD` | `2024-01-15` |
| `date` | `YYYY-MM-DDTHH:MM:SS` | `2024-01-15T14:30:00` |
| `unix` | Unix timestamp | `1704067200` |

## Examples

### Files older than 30 days
```yaml
- date_modified:
    before:
      days_ago: 30
```

### Files modified in the last hour
```yaml
- date_modified:
    after:
      hours_ago: 1
```

### Files modified before 2024
```yaml
- date_modified:
    before:
      date: "2024-01-01"
```

### Files modified in the last week
```yaml
- date_modified:
    after:
      weeks_ago: 1
```

## Example: clean old downloads

```yaml
rules:
  - name: Archive Old Downloads
    locations: ~/Downloads
    filters:
      - date_modified:
          before:
            days_ago: 90
    actions:
      - move: ~/Archive/Downloads
```
