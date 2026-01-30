# Trash

Moves files to the system trash/recycle bin. This is a safer alternative to `delete` since files can be recovered.

## Syntax

```yaml
- trash
```

## Options

This action has no options.

## Examples

### Clean old downloads
```yaml
rules:
  - name: Trash Old Downloads
    locations: ~/Downloads
    filters:
      - date_modified:
          before:
            days_ago: 30
    actions:
      - trash
```

### Trash old backups
```yaml
rules:
  - name: Trash Old Backups
    locations: ~/Backup
    filters:
      - name: "*_backup_*"
      - date_modified:
          before:
            weeks_ago: 4
    actions:
      - trash
```

### Trash temporary files
```yaml
rules:
  - name: Clean Temp Files
    locations: ~/Projects
    subfolders: true
    filters:
      - extension: [tmp, temp, bak, swp]
    actions:
      - trash
```

## Platform behavior

| platform | trash location |
|----------|----------------|
| macOS | `~/.Trash` |
| Linux | `~/.local/share/Trash` (freedesktop.org spec) |
| Windows | Recycle Bin |

## Comparison with delete

| aspect | trash | delete |
|--------|-------|--------|
| Recoverable | Yes | No |
| Disk space | Still used until emptied | Freed immediately |
| Speed | Slightly slower | Faster |
| Safety | Safe | Dangerous |

## Notes

- Safer than `delete` for automated cleanup rules
- Uses the system's native trash mechanism
- On Linux, follows the freedesktop.org trash specification
