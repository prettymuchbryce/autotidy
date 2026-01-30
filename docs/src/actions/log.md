# log

Logs a message when a file matches. Useful for debugging rules and testing configurations.

## Syntax

```yaml
# Simple form - just the message
- log: "Found file: ${name}.${ext}"

# Explicit form with level
- log:
    msg: "Processing ${name}"
    level: info
```

## Options

| option | type | required | default | description |
|--------|------|----------|---------|-------------|
| `msg` | string | Yes | - | Message to log (supports templates) |
| `level` | string | No | `info` | Log level |

## Log levels

| level | description |
|-------|-------------|
| `debug` | Detailed debugging information |
| `info` | General information (default) |
| `warn` | Warning messages |
| `error` | Error messages |

## Examples

### Simple logging
```yaml
- log: "Matched: ${name}.${ext}"
```

### Debug logging
```yaml
- log:
    msg: "File details: ${name}, ext: ${ext}"
    level: debug
```

### Log before action
```yaml
actions:
  - log: "Moving ${name}.${ext} to archive"
  - move: ~/Archive
```

### Testing rules
```yaml
rules:
  - name: Test Image Filter
    locations: ~/Downloads
    filters:
      - extension: [jpg, png, gif]
    actions:
      - log: "Would process image: ${name}.${ext}"
      # Comment out actual action while testing
      # - move: ~/Pictures
```

### Conditional logging
```yaml
rules:
  - name: Large File Warning
    locations: ~/Downloads
    filters:
      - size:
          gt: 1GB
    actions:
      - log:
          msg: "Large file detected: ${name}.${ext}"
          level: warn
```

## Template variables

The `msg` field supports all template variables:

```yaml
- log: "File: ${name}.${ext}, Year: ${year}, Month: ${month}"
```

See [Templates](../templates.md) for all available variables.

## Use cases

### Audit trail

Log actions for review:

```yaml
rules:
  - name: Archive with Logging
    locations: ~/Downloads
    filters:
      - extension: pdf
    actions:
      - log:
          msg: "Archiving PDF: ${name}"
          level: info
      - move: ~/Archive/PDFs
```

## Notes

- Does not modify files in any way
- Output goes to autotidy's log output
- Useful during rule development and debugging
- Can be combined with other actions in the same rule
