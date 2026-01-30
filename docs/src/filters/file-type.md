# file_type

Matches by file type: regular file, directory, or symlink.

## Syntax

```yaml
# Single type
- file_type: file

# Multiple types
- file_type: [file, directory]

# Explicit form
- file_type:
    types: [file, symlink]
```

## Types

| type | aliases | description |
|------|---------|-------------|
| `file` | - | Regular files |
| `directory` | `dir`, `folder` | Directories |
| `symlink` | - | Symbolic links |

## Examples

### Match only files (not directories)
```yaml
filters:
  - file_type: file
```

### Match directories only
```yaml
filters:
  - file_type: directory
```

### Match files and symlinks
```yaml
filters:
  - file_type: [file, symlink]
```

### Exclude directories
```yaml
filters:
  - not:
      - file_type: directory
```

## Notes

- Symlinks are checked without following them (uses `lstat`)
- This filter is useful when processing directories recursively to skip subdirectories
