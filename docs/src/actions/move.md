# move

Moves files to a specified destination directory.

## Syntax

```yaml
# Simple form - just the destination
- move: ~/Documents

# Explicit form with options
- move:
    dest: ~/Documents
    on_conflict: skip
```

## Options

| option | type | required | default | description |
|--------|------|----------|---------|-------------|
| `dest` | string | Yes | - | Destination directory path |
| `on_conflict` | string | No | `rename_with_suffix` | How to handle existing files |

## Conflict handling

| mode | behavior |
|------|----------|
| `rename_with_suffix` | Add numeric suffix (file_2.txt, file_3.txt, etc.) |
| `skip` | Don't move if destination file exists |
| `overwrite` | Replace existing destination file |

## Examples

### Move downloads to documents
```yaml
- move: ~/Documents
```

### Move, skipping files on conflict
```yaml
- move:
    dest: ~/Documents
    on_conflict: skip
```

### Organize by date
```yaml
- move: ~/Photos/%Y/%m
```

### Move to categorized folders
```yaml
rules:
  - name: Organize Images
    locations: ~/Downloads
    filters:
      - extension: [jpg, png, gif]
    actions:
      - move: ~/Pictures/Downloads

  - name: Organize Documents
    locations: ~/Downloads
    filters:
      - extension: [pdf, doc, docx]
    actions:
      - move: ~/Documents/Downloads
```

## Template variables

The destination directory supports template variables:

```yaml
- move: ~/Archive/%Y/%m
```

See [Templates](../templates.md) for all available variables.

## Notes

- The destination must be a directory, not a file path
- The original filename is preserved (use [rename](rename.md) to change filenames)
- Creates destination directories if they don't exist
- Default conflict handling adds a numeric suffix (file_2.txt)
