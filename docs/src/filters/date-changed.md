# date_changed

Matches files by their metadata change time (ctime).

## Syntax

```yaml
# Files with metadata changed more than 30 days ago
- date_changed:
    before:
      days_ago: 30

# Files with recent metadata changes
- date_changed:
    after:
      hours_ago: 24
```

## Operators

| operator | description |
|----------|-------------|
| `before` | Metadata was changed before this time |
| `after` | Metadata was changed after this time |

## Time specifications

See [date_modified](date-modified.md) for all available time specification options:

- Relative: `seconds_ago`, `minutes_ago`, `hours_ago`, `days_ago`, `weeks_ago`, `months_ago`, `years_ago`
- Absolute: `date` (YYYY-MM-DD), `unix` (timestamp)

## What is ctime?

The "change time" (ctime) is updated when file metadata changes, including:

- Permission changes (`chmod`)
- Ownership changes (`chown`)
- Link count changes (hard links created/removed)
- File content modifications (also updates mtime)

This is different from modification time (mtime), which only updates when file content changes.

## Examples

### Files with old metadata
```yaml
- date_changed:
    before:
      days_ago: 90
```

### Recently modified files (including metadata)
```yaml
- date_changed:
    after:
      hours_ago: 1
```

## Notes

- On Unix systems, ctime cannot be set manually
- Copying a file typically resets its ctime to the current time
- On Windows, this typically maps to the file's change time
