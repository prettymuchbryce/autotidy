# copy

Copies a file to a new name in the same directory.

## Syntax

```yaml
# Simple form - just the new filename
- copy: backup.txt

# With template
- copy: "${name}_backup${ext}"

# With conflict handling
- copy:
    new_name: "${name}_backup${ext}"
    on_conflict: overwrite
```

## Options

| option | type | required | default | description |
|--------|------|----------|---------|-------------|
| `new_name` | string | Yes | - | New filename (supports templates) |
| `on_conflict` | string | No | `rename_with_suffix` | How to handle existing files |

## Conflict handling

| mode | behavior |
|------|----------|
| `rename_with_suffix` | Add numeric suffix (file_2.txt, file_3.txt, etc.). This is the default |
| `skip` | Don't copy if destination file exists |
| `overwrite` | Replace existing destination file |

## Examples

### Simple copy
```yaml
- copy: backup.txt
```

### Copy with original name preserved
```yaml
- copy: "${name}_copy${ext}"
```

### Copy with timestamp
```yaml
- copy: "${name}_%Y%m%d${ext}"
```

### Backup with overwrite
```yaml
- copy:
    new_name: "${name}_backup${ext}"
    on_conflict: overwrite
```

## Template variables

The `new_name` field supports template variables:

```yaml
- copy: "${name}_%H%M${ext}"
```

See [Templates](../templates.md) for all available variables.

## Notes

- Copies to the **same directory** as the source file
- The `new_name` must not contain path separators
- Original file remains unchanged
- To copy to a different directory, first `copy` the file and then use the [move](move.md) action
