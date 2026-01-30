# mime_type

Matches files by their MIME type, detected from file content (not extension).

## Syntax

```yaml
# Single MIME type
- mime_type: "image/png"

# Wildcard pattern
- mime_type: "image/*"

# Multiple types
- mime_type: ["image/*", "video/*"]

# Explicit form
- mime_type:
    mime_types: ["application/pdf", "application/msword"]
```

## Glob patterns

MIME type patterns support glob matching:

| pattern | matches |
|---------|---------|
| `image/*` | All image types |
| `video/*` | All video types |
| `text/*` | All text types |
| `application/*` | All application types |

## Common MIME types

### Images
- `image/jpeg` - JPEG images
- `image/png` - PNG images
- `image/gif` - GIF images
- `image/webp` - WebP images
- `image/svg+xml` - SVG images

### Documents
- `application/pdf` - PDF documents
- `application/msword` - Word documents (.doc)
- `application/vnd.openxmlformats-officedocument.wordprocessingml.document` - Word (.docx)

### Video
- `video/mp4` - MP4 video
- `video/x-matroska` - MKV video
- `video/quicktime` - QuickTime video

### Audio
- `audio/mpeg` - MP3 audio
- `audio/wav` - WAV audio
- `audio/flac` - FLAC audio

### Archives
- `application/zip` - ZIP archives
- `application/x-tar` - TAR archives
- `application/gzip` - Gzip compressed

## Examples

### Match all images
```yaml
- mime_type: "image/*"
```

### Match videos and images
```yaml
- mime_type: ["image/*", "video/*"]
```

### Match PDF files
```yaml
- mime_type: "application/pdf"
```

### Match text files (including code)
```yaml
- mime_type: "text/*"
```

## Notes

- MIME type is detected from file content, not the file extension
- This is more accurate than extension-based filtering, but slower as the file needs to be read
- Directories always return `false` (no MIME type)
- Detection uses the first few bytes of the file (magic numbers)
