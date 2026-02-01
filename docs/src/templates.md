# Templates

Templates allow you to use dynamic values in action parameters like destination paths and new filenames. Variables are replaced with actual values when the action executes.

## Syntax

Template variables use the `${variable}` syntax:

```yaml
- move: ~/Archive/%Y/%m
- rename: "${name}_backup${ext}"
```

## File variables

| variable | description | example |
|----------|-------------|---------|
| `${name}` | Filename without extension | `document` |
| `${ext}` | File extension (with dot) | `.pdf` |

### Examples

For a file named `report.pdf`:

| template | result |
|----------|--------|
| `${name}` | `report` |
| `${ext}` | `.pdf` |
| `${name}${ext}` | `report.pdf` |
| `${name}_copy${ext}` | `report_copy.pdf` |
| `backup_${name}${ext}` | `backup_report.pdf` |

## Time variables

Time variables use [strftime](https://strftime.org/) format tokens:

| token | description | example |
|-------|-------------|---------|
| `%Y` | Year (4 digits) | `2024` |
| `%m` | Month (01-12) | `03` |
| `%d` | Day (01-31) | `15` |
| `%H` | Hour (00-23) | `14` |
| `%M` | Minute (00-59) | `30` |
| `%S` | Second (00-59) | `45` |

### Common Patterns

| pattern | example output |
|---------|----------------|
| `%Y-%m-%d` | `2024-03-15` |
| `%Y%m%d` | `20240315` |
| `%Y/%m/%d` | `2024/03/15` |
| `%H:%M:%S` | `14:30:45` |
| `%Y%m%d_%H%M%S` | `20240315_143045` |

### Additional Time Tokens

| token | description | example |
|-------|-------------|---------|
| `%y` | Year (2 digits) | `24` |
| `%B` | Month name (full) | `March` |
| `%b` | Month name (abbr) | `Mar` |
| `%A` | Weekday name (full) | `Friday` |
| `%a` | Weekday name (abbr) | `Fri` |
| `%j` | Day of year (001-366) | `074` |
| `%U` | Week number (00-53) | `11` |
| `%W` | Week number (Monday start) | `10` |

## Examples

### Organize by date

```yaml
rules:
  - name: Organize Downloads
    locations: ~/Downloads
    actions:
      - move: ~/Archive/%Y/%m/%d
```

Files are organized into folders like `~/Archive/2024/03/15/`.

### Add timestamp to filename

```yaml
rules:
  - name: Timestamp Files
    locations: ~/Documents
    filters:
      - extension: pdf
    actions:
      - rename: "${name}_%Y%m%d${ext}"
```

`report.pdf` becomes `report_20240315.pdf`.

### Create dated copies

```yaml
rules:
  - name: Daily Backup
    locations: ~/Documents
    actions:
      - copy: "${name}_%Y%m%d${ext}"
```

Creates a timestamped copy of each file in the same directory.

### Organize photos by date

```yaml
rules:
  - name: Organize Photos
    locations: ~/Downloads
    filters:
      - extension: [jpg, jpeg, png, heic]
    actions:
      - move: ~/Pictures/%Y/%B
```

Photos organized into folders like `~/Pictures/2024/March/`.

### Archive old files

```yaml
rules:
  - name: Archive Old Downloads
    locations: ~/Downloads
    filters:
      - date_modified:
          before:
            days_ago: 30
    actions:
      - move: ~/Archive/Downloads/%Y-%m
```

### Unique filenames with timestamp

```yaml
rules:
  - name: Rename Duplicates
    locations: ~/Downloads
    actions:
      - rename: "${name}_%Y%m%d_%H%M%S${ext}"
```

## Combining variables

Mix file and time variables:

```yaml
- move: ~/Archive/%Y/${name}/%m
- rename: "${name}_%Y%m%d_%H%M%S${ext}"
- copy: "${name}_backup_%Y%m%d${ext}"
```

## Where templates work

Templates are supported in:

| action | fields |
|--------|--------|
| `move` | `dest` |
| `copy` | `new_name` |
| `rename` | `new_name` |
| `log` | `msg` |

## Notes

- Time values are evaluated at action execution time
- File variables (`${name}`, `${ext}`) come from the matched file
- Unknown variables are left unchanged in the output
- Paths are created automatically if they don't exist
