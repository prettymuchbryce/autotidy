<p align="center">
  <img src="icon.png" alt="autotidy" width="128" height="128">
</p>

<p align="center">Automatically organize files using declarative rules</p>

## About

**autotidy** allows you to write declarative rules and apply them automatically when files in a directory change. It consists of both a background process (daemon), and a CLI tool. Rules are defined in a yaml file. They include the directories that should be watched, and actions to be performed when files in those directories change. Available actions include moving, copying, renaming, and deleting files. Additionally, files in a watched directory can be filtered further to include only files with a particular name, extension, mime-type, or size.

## Features

- **Automatic** - Runs in the background, watching directories and triggers your rules when their contents change
- **Declarative** - Define your rules in yaml
- **Filters** - Match files by name, extension, size, date, MIME type, file type
- **Actions** - move, copy, rename, delete, and trash files that pass your filters
- **Dry-run** - Preview what your config _would_ do before running it with `autotidy run`
- **Cross-platform** - Linux, macOS, Windows
- **Open source** - MIT licensed
