# name

Matches files by filename using glob patterns or regular expressions.

## Syntax

```yaml
# Glob pattern (shorthand)
- name: "*.txt"

# Glob pattern (explicit)
- name:
    glob: "report_*.pdf"

# Regular expression
- name:
    regex: "^file_\d{4}\.txt$"
```

## Options

| option | type | description |
|--------|------|-------------|
| `glob` | string | Glob pattern to match |
| `regex` | string | Regular expression pattern |

Only one of `glob` or `regex` can be specified.

## Glob patterns

The glob pattern is matched against the **base filename only** (not the full path).

| pattern | matches |
|---------|---------|
| `*` | Any sequence of characters |
| `?` | Any single character |
| `[abc]` | Any character in the set |
| `[a-z]` | Any character in the range |
| `**` | Any path (in glob context) |

## Examples

### Match all text files
```yaml
- name: "*.txt"
```

### Match files starting with "report"
```yaml
- name: "report*"
```

### Match files with numbers in name
```yaml
- name:
    regex: ".*\d+.*"
```

### Match screenshot files
```yaml
- name: "Screenshot*"
```

### Exclude hidden files (macOS)
```yaml
filters:
  - not:
      - any:
          - name: ".*"
          - name: .DS_Store
```
