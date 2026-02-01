# Actions

Actions define what happens to files that match your filters. Each rule can have one or more actions that execute in order.

## Available actions

| action | description |
|--------|-------------|
| [move](move.md) | Move files to a new location |
| [copy](copy.md) | Copy files with a new name in the same directory |
| [rename](rename.md) | Rename files in place |
| [delete](delete.md) | Permanently delete files |
| [trash](trash.md) | Move files to system trash |
| [log](log.md) | Log a message (for debugging/testing) |

## Action syntax

Actions are defined as a list under the `actions` key:

```yaml
rules:
  - name: Example Rule
    locations: ~/Downloads
    actions:
      - move: ~/Documents
      - log: "Moved ${name}"
```

## Multiple actions

Actions execute in order. This allows chaining operations:

```yaml
actions:
  - copy: "${name}_backup${ext}"   # First, create a backup copy
  - move: ~/Documents              # Then, move the original
  - log: "Processed ${name}"
```

## Template variables

Most actions support template variables in paths and messages:

| variable | description |
|----------|-------------|
| `${name}` | Filename without extension |
| `${ext}` | File extension (with dot) |
| `%Y` | Current year (4 digits) |
| `%m` | Current month (01-12) |
| `%d` | Current day (01-31) |
| `%H` | Current hour (00-23) |
| `%M` | Current minute (00-59) |
| `%S` | Current second (00-59) |

See [Templates](../templates.md) for full details.

## Conflict handling

Actions that create files (`move`, `copy`, `rename`) support conflict handling:

```yaml
actions:
  - move:
      dest: ~/Documents
      on_conflict: skip  # Don't move if destination exists
```

| mode | behavior |
|------|----------|
| `rename_with_suffix` | Add numeric suffix (file_2.txt, file_3.txt, etc.) |
| `skip` | Don't move/copy if destination exists |
| `overwrite` | Replace existing file |

Default is `rename_with_suffix`.

## Dry run mode

Use `--dry-run` to preview actions without executing them:

```bash
autotidy run --dry-run
```
