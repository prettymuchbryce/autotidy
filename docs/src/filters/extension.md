# extension

Matches files by their file extension.

## Syntax

```yaml
# Single extension
- extension: pdf

# With dot (also works)
- extension: .pdf

# Multiple extensions
- extension: [pdf, doc, docx]

# Explicit form
- extension:
    extensions: [pdf, doc]
```

## Options

| option | type | description |
|--------|------|-------------|
| `extensions` | string/list | One or more extensions to match |

## Glob support

Extensions support glob patterns:

```yaml
# Match doc and docx
- extension: "doc*"

# Match any single-character extension
- extension: "?"
```

## Examples

### Match PDF files
```yaml
- extension: pdf
```

### Match image files
```yaml
- extension: [jpg, jpeg, png, gif, webp]
```

### Match document files
```yaml
- extension: [pdf, doc, docx, xls, xlsx, ppt, pptx]
```

### Match video files
```yaml
- extension: [mp4, mkv, avi, mov, wmv]
```

### Match all Microsoft Office formats
```yaml
- extension: ["doc*", "xls*", "ppt*"]
```
