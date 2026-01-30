# Quick Start

## 1. Install

See [Installation](installation/README.md)

## 2. Verify installation

Verify autotidy is installed and running.

```sh
❯ autotidy status
status      🟢 running
config      ~/.config/autotidy/config.yaml
watching    none
rules
  ⚠ none
```

## 3. Create your first rule

Edit `~/.config/autotidy/config.yaml`:

```yaml
rules:
  - name: Organize PDFs
    locations: ~/Downloads
    filters:
      - extension: pdf
    actions:
      - move: ~/Documents/PDFs
```

## 4. Dry run your rule

Use `autotidy run` to preview what your rules would do without making changes:

```sh
❯ autotidy run
Dry-run mode enabled (pass --dry-run=false to perform a one-off run of all rules)

━━━ Rule: Clean up Desktop images ━━━
~/Downloads/document.pdf
├── filters:
│   └── extension:     ✓
└── actions:
    └── move:          ✓ → ~/Documents/PDFs
```

## 5. Reload your rules

Tell the daemon to pick up your config changes:

```sh
❯ autotidy reload
Reloaded ~/.config/autotidy/config.yaml
```

Then verify your rule is active:

```sh
❯ autotidy status
status      🟢 running
config      ~/.config/autotidy/config.yaml
watching    1 directories
rules
  🟢 Organize PDFs
    last run: 1 hour ago (2ms, 1 files)
```

Your rule is now running. Any PDFs added to `~/Downloads` will be moved to `~/Documents/PDFs`.

## Next steps

- [Configuration](configuration.md) - Full configuration reference
- [Filters](filters/README.md) - All available filters
- [Actions](actions/README.md) - All available actions
