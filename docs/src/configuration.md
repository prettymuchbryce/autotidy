# Configuration

Configuration for autotidy can be found at the following location:

| platform | path |
|----------|------|
| Linux    | `~/.config/autotidy/config.yaml` |
| macOS    | `~/.config/autotidy/config.yaml` |
| Windows  | `%APPDATA%\autotidy\config.yaml` |

## Example

```yaml
# Moves PDFs and Word docs from Downloads/Desktop to ~/Documents
# organized by extension
rules:
  - name: Organize Downloads
    locations:
      - ~/Downloads
      - ~/Desktop
    filters:
      - extension: [pdf, doc, docx]
    actions:
      - move: ~/Documents/${ext}
```

## Reference

- [Rules](rules.md) - Rule properties and structure
- [Filters](filters/README.md) - Filter types and boolean operators
- [Actions](actions/README.md) - Available actions
- [Templates](templates.md) - Variables like `${name}`, `${ext}`, `${date}`
- [Additional Options](options.md) - Daemon and logging settings

## Hot Reload

The autotidy daemon supports hot-reloading of configuration:

```bash
autotidy reload
```
