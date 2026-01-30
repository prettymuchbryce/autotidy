# date_accessed

Matches files by their last access time.

## Syntax

```yaml
# Files not accessed in 30 days
- date_accessed:
    before:
      days_ago: 30

# Files accessed recently
- date_accessed:
    after:
      hours_ago: 24
```

## Operators

| operator | description |
|----------|-------------|
| `before` | File was accessed before this time |
| `after` | File was accessed after this time |

## Time specifications

See [date_modified](date-modified.md) for all available time specification options:

- Relative: `seconds_ago`, `minutes_ago`, `hours_ago`, `days_ago`, `weeks_ago`, `months_ago`, `years_ago`
- Absolute: `date` (YYYY-MM-DD), `unix` (timestamp)

## Examples

### Files not accessed in 90 days
```yaml
- date_accessed:
    before:
      days_ago: 90
```

### Files accessed today
```yaml
- date_accessed:
    after:
      hours_ago: 24
```

## Notes

- Access time tracking may be disabled on some filesystems for performance
- On Linux, the `noatime` mount option disables access time updates
- macOS may not update access times in all cases
