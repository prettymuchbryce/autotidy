# delete

Permanently deletes files and directories. For safer deletes consider [trash](trash.md) instead.


## Syntax

```yaml
- delete
```

## Options

This action has no options.

## Examples

### Delete old temp files
```yaml
rules:
  - name: Clean Temp Files
    locations: ~/tmp
    filters:
      - date_modified:
          before:
            days_ago: 7
    actions:
      - delete
```

### Delete by extension
```yaml
rules:
  - name: Remove Log Files
    locations: ~/Projects
    subfolders: true
    filters:
      - extension: log
      - date_modified:
          before:
            days_ago: 30
    actions:
      - delete
```

### Delete with size filter
```yaml
rules:
  - name: Remove Large Temp Files
    locations: /tmp
    filters:
      - size:
          gt: 100MB
      - extension: [tmp, temp, cache]
    actions:
      - delete
```

## Safety recommendations

### Always use dry run first

```bash
autotidy run --dry-run
```

### Combine with specific filters

Don't use `delete` without filters:

```yaml
# DANGEROUS - deletes everything!
rules:
  - name: Bad Rule
    locations: ~/Documents
    actions:
      - delete

# SAFE - specific filters
rules:
  - name: Good Rule
    locations: ~/Documents
    filters:
      - extension: tmp
      - date_modified:
          before:
            days_ago: 30
    actions:
      - delete
```

### Consider using trash instead

For most cases, [trash](trash.md) is safer:

```yaml
# Recoverable
- trash

# Permanent
- delete
```

## Notes

- **Permanent**: Files cannot be recovered after deletion
- **No confirmation**: Executes immediately when rules match
