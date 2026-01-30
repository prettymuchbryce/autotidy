# rename

Renames files in place.

## Syntax

```yaml
# Simple form - new name with template
- rename: "${name}_archived.${ext}"

# Explicit form with options
- rename:
    new_name: "${name}_${year}${month}${day}.${ext}"
    on_conflict: rename_with_suffix
```

## Options

| option | type | required | default | description |
|--------|------|----------|---------|-------------|
| `new_name` | string | Yes | - | New filename (supports templates) |
| `on_conflict` | string | No | `rename_with_suffix` | How to handle existing files |

## Conflict handling

| mode | behavior |
|------|----------|
| `rename_with_suffix` | Add numeric suffix (file_2.txt, file_3.txt, etc.) |
| `skip` | Don't rename if a file with the new name exists |
| `overwrite` | Replace existing file with the new name |

## Examples

### Add date suffix
```yaml
- rename: "${name}_${year}${month}${day}.${ext}"
```

### Lowercase extension
```yaml
- rename: "${name}.${ext}"
```

### Add prefix
```yaml
- rename: "archived_${name}.${ext}"
```

### Rename with skip on conflict
```yaml
- rename:
    new_name: "processed_${name}.${ext}"
    on_conflict: skip
```

### Standardize filenames
```yaml
rules:
  - name: Rename Screenshots
    locations: ~/Desktop
    filters:
      - name: "Screen Shot*"
    actions:
      - rename: "screenshot_${year}${month}${day}_${hour}${minute}${second}.${ext}"
```

## Template variables

The `new_name` field supports all template variables:

| variable | example output |
|----------|----------------|
| `${name}` | `document` |
| `${ext}` | `pdf` |
| `${year}` | `2024` |
| `${month}` | `03` |
| `${day}` | `15` |
| `${hour}` | `14` |
| `${minute}` | `30` |
| `${second}` | `45` |

See [Templates](../templates.md) for full details.

## Notes

- File stays in the same directory
- Only the filename changes, not the location
- Default conflict handling adds a numeric suffix (file_2.txt)
- Use rename and [move](move.md) in succession to rename a file and move it to another directory
- Template variables are evaluated at the time of action execution
