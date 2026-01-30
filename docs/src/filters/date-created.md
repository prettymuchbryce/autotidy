# date_created

Matches files by their creation time (birth time).

## Syntax

```yaml
# Files created more than 30 days ago
- date_created:
    before:
      days_ago: 30

# Files created recently
- date_created:
    after:
      hours_ago: 24
```

## Operators

| operator | description |
|----------|-------------|
| `before` | File was created before this time |
| `after` | File was created after this time |

## Time specifications

See [date_modified](date-modified.md) for all available time specification options:

- Relative: `seconds_ago`, `minutes_ago`, `hours_ago`, `days_ago`, `weeks_ago`, `months_ago`, `years_ago`
- Absolute: `date` (YYYY-MM-DD), `unix` (timestamp)

## Examples

### Files created in the last week
```yaml
- date_created:
    after:
      weeks_ago: 1
```

### Files created before 2024
```yaml
- date_created:
    before:
      date: "2024-01-01"
```

## Platform support

| platform | support |
|----------|---------|
| macOS | Full support |
| Windows | Full support |
| Linux | Depends on filesystem (ext4 4.11+, btrfs) |

On filesystems that don't track creation time, this filter may not work as expected.
