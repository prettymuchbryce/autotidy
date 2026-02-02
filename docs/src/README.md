<p align="center">
  <img src="icon.png" alt="autotidy" width="256">
</p>

<p align="center">Automatically organize files using declarative rules</p>

## About

autotidy is a cross-platform file organization daemon. Define rules in YAML, and it watches your folders for changes, moving, renaming, or organizing files as they appear.

- **Automatic** - Runs in the background, watching directories. Triggers your rules when contents change.
- **Declarative** - Define your rules in YAML. No code required.
- **Filters** - Match files by name, extension, size, date, MIME type, file type.
- **Actions** - Move, copy, rename, delete, and trash files that pass your filters.
- **Standalone** - Standalone compiled binary. No runtime dependencies.
- **Dry-run** - Preview what your config _would_ do before running it with `autotidy run`.
- **Cross-platform** - Linux, macOS, Windows (experimental)
- **Open source** - MIT licensed
