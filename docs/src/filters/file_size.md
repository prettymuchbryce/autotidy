# file_size

Matches files by their size. Directories are ignored (filter returns false for directories).

## Syntax

### Shorthand
```yaml
- file_size: "> 10mb"
- file_size: "<= 500kb"
- file_size: ">= 1gb"
```

### Explicit
```yaml
- file_size:
    greater_than:
      mb: 10

- file_size:
    between:
      min:
        mb: 1
      max:
        gb: 1
```

## Operators

| shorthand | explicit | description |
|-----------|----------|-------------|
| `>` | `greater_than` | Strictly greater than |
| `>=` | `at_least` | Greater than or equal |
| `<` | `less_than` | Strictly less than |
| `<=` | `at_most` | Less than or equal |
| - | `between` | Within a range (inclusive) |

## Size units

| unit | bytes |
|------|-------|
| `b` | 1 |
| `kb` | 1,024 |
| `mb` | 1,048,576 |
| `gb` | 1,073,741,824 |
| `tb` | 1,099,511,627,776 |

Units are case-insensitive (`MB`, `mb`, `Mb` all work).

## Examples

### Match large files (> 100MB)
```yaml
- file_size: "> 100mb"
```

### Match small files (< 1KB)
```yaml
- file_size: "< 1kb"
```

### Match files in a range
```yaml
- file_size:
    between:
      min:
        mb: 10
      max:
        mb: 100
```

### Match empty files
```yaml
- file_size: "< 1b"
```

### Match files at least 1GB
```yaml
- file_size: ">= 1gb"
```
