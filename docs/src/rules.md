# Rules

Each rule defines directories (`locations`) to be watched. When the contents of those locations change, the containing files which match the filters will have the actions applied against them sequentially.

By default, rules evaluate both files and directories. Use the [file_type](filters/file-type.md) filter to limit to one or the other.

```yaml
# Trashes files (not directories) in ~/Downloads that haven't been modified in 30 days
rules:
  - name: Clean Old Downloads
    locations: ~/Downloads
    filters:
      - file_type: file
      - date_modified:
          before:
            days_ago: 30
    actions:
      - trash
```

## Properties

| property | type | default | description |
|----------|------|---------|-------------|
| `name` | string | required | Rule identifier (shown in logs and status) |
| `enabled` | bool | `true` | Whether the rule is active |
| `recursive` | bool | `false` | Process subdirectories |
| `traversal` | string | `depth-first` | `depth-first` or `breadth-first` |
| `locations` | string/list | required | Directories to watch |
| `filters` | list | - | Filter expressions |
| `actions` | list | required | Actions to execute |

## Locations

Locations can be a single path or a list:

```yaml
# Single location
locations: ~/Downloads

# Multiple locations
locations:
  - ~/Downloads
  - ~/Desktop
```

## Filters

Filters determine which files are processed. Filters are AND'd together by default.

```yaml
# Matches PDFs that haven't been modified in 30 days
filters:
  - extension: pdf
  - date_modified:
      before:
        days_ago: 30
```

Use `any:` for OR logic and `not:` for negation:

```yaml
# Matches PDFs or Word documents, excluding any with "_backup" in the name
filters:
  - any:
      - extension: pdf
      - extension: [doc, docx]
  - not:
      - name: "*_backup*"
```

| operator | behavior |
|----------|----------|
| (default) | All filters must match (AND) |
| `any:` | At least one child must match (OR) |
| `not:` | None of the children must match |

See [Filters](filters/README.md) for all available filter types.

## Actions

Actions are executed in order for each matching file:

```yaml
# Logs the file, renames it with a "_backup" suffix, then moves it to ~/Archive
actions:
  - log: "Processing ${name}"
  - rename: "${name}_backup${ext}"
  - move: ~/Archive
```

See [Actions](actions/README.md) for all available action types.

## Recursive

By default, rules only process files directly in the specified locations. Set `recursive: true` to also process files in subdirectories.

```yaml
# Moves all PDFs from ~/Downloads and its subdirectories to ~/Documents/PDFs
rules:
  - name: Organize All Downloads
    locations: ~/Downloads
    recursive: true
    filters:
      - extension: pdf
    actions:
      - move: ~/Documents/PDFs
```

> **Note:** Recursion is depth-first by default. This means children are processed before their parents.

## Traversal

When `recursive: true`, the `traversal` option controls the order files are processed:

- `depth-first` (default) - processes deepest files first, then works upward
- `breadth-first` - processes files at each level before going deeper

```yaml
# Deletes empty directories in ~/Downloads, processing deepest folders first
rules:
  - name: Clean Empty Folders
    locations: ~/Downloads
    recursive: true
    traversal: depth-first
    filters:
      - file_type: directory
    actions:
      - delete
```

> **Note:** Using `breadth-first` when moving, renaming, or copying a directory will result in subsequent actions being performed on the moved/renamed/copied contents.
