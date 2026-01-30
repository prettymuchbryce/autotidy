# Filters

Filters determine which files or directories are processed by a rule. They are specified in the `filters` section of a rule.

## Available filters

| filter | description |
|--------|-------------|
| [name](name.md) | Match by filename (glob or regex) |
| [extension](extension.md) | Match by file extension |
| [file_size](file_size.md) | Match by file size |
| [file_type](file-type.md) | Match by type (file, directory, symlink) |
| [date_modified](date-modified.md) | Match by modification time |
| [date_accessed](date-accessed.md) | Match by access time |
| [date_created](date-created.md) | Match by creation time |
| [date_changed](date-changed.md) | Match by metadata change time |
| [mime_type](mime-type.md) | Match by MIME type |

## Filter logic

Filters are AND'd together by default. Use `any:` for OR logic and `not:` for negation.

```yaml
# All filters must match (AND)
filters:
  - extension: pdf
  - date_modified:
      before:
        days_ago: 30

# At least one must match (OR)
filters:
  - any:
      - file_size: "> 100mb"
      - date_modified:
          before:
            days_ago: 30

# None must match (NOT)
filters:
  - not:
      - name: "*.tmp"
```

| operator | behavior |
|----------|----------|
| (default) | All filters must match (AND) |
| `any:` | At least one child must match (OR) |
| `not:` | None of the children must match |

Both `any:` and `not:` can be nested for complex boolean expressions.

## Examples

### Match all files (not directories)
```yaml
filters:
  - file_type: file
```

### Match specific file types
```yaml
filters:
  - extension: [jpg, png, gif]
```

### Match files by name pattern
```yaml
filters:
  - name: "Screenshot*"
```

### Match large files
```yaml
filters:
  - file_size: "> 100mb"
```

### Match old files
```yaml
filters:
  - date_modified:
      before:
        days_ago: 30
```

### Match old or large files (OR)
```yaml
filters:
  - any:
      - file_size: "> 100mb"
      - date_modified:
          before:
            days_ago: 30
```

### Match old PDFs (AND)
```yaml
filters:
  - extension: pdf
  - date_modified:
      before:
        days_ago: 30
```

### Exclude temp files (NOT)
```yaml
filters:
  - not:
      - extension: [tmp, temp, bak]
```

### Complex: (old OR large) AND documents AND NOT backup
```yaml
filters:
  - any:
      - file_size: "> 100mb"
      - date_modified:
          before:
            days_ago: 90
  - extension: [pdf, doc, docx]
  - not:
      - name: "*_backup*"
```
