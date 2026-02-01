# autotidy

<p align="center">
  <img src="assets/icon-512.png" alt="autotidy" width="256">
</p>
<p align="center">
  Automatically organize files using declarative rules
</p>


<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License: MIT"></a>
  <a href="https://github.com/prettymuchbryce/autotidy/actions/workflows/test.yaml"><img src="https://github.com/prettymuchbryce/autotidy/actions/workflows/test.yaml/badge.svg" alt="unit tests"></a>
  <a href="https://github.com/prettymuchbryce/autotidy/actions/workflows/install-tests.yml"><img src="https://github.com/prettymuchbryce/autotidy/actions/workflows/install-tests.yml/badge.svg" alt="install tests"></a>
  <a href="https://github.com/prettymuchbryce/autotidy/releases"><img src="https://img.shields.io/github/v/release/prettymuchbryce/autotidy?include_prereleases" alt="Release"></a>
</p>

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

## Examples

Below are some example rules. Rules are defined in a `config.yaml` file.

```yaml
# Organize downloads by extension
rules:
 - name: Organize Downloads
   locations: ~/Downloads
   filters:
     - extension: [pdf, doc, docx]
   actions:
     - move: ~/Documents
```

```yaml
# Sort images by date taken
rules:
 - name: Sort Photos
   locations: ~/Downloads
   recursive: true
   filters:
     - extension: [jpg, jpeg, png, heic]
   actions:
     - move: ~/Pictures/%Y/%B  # e.g., ~/Pictures/2024/January
```

```yaml
# Trash old Downloads
rules:
 - name: Trash old downloads
   locations: ~/Downloads
   filters:
     - date_created:
         before:
           days_ago: 90
   actions:
     - trash
```

```yaml
# Backup images or videos (OR logic with any:)
# Copy the file, then move the copy to the backups directory
rules:
 - name: Backup Media
   locations: ~/Downloads
   filters:
     - any:
         - mime_type: "image/*"
         - mime_type: "video/*"
   actions:
     - copy: "${name}_backup${ext}"
     - move: ~/backups
```

You can find an exhaustive list of configuration options [here](https://prettymuchbryce.github.io/autotidy/configuration.html).

## Quick start

### Homebrew (macOS)

```bash
brew install prettymuchbryce/tap/autotidy
brew services start autotidy
autotidy status
```

### Nix (Home Manager)

```nix
{
  inputs.autotidy.url = "github:prettymuchbryce/autotidy";

  imports = [ inputs.autotidy.homeModules.default ];

  services.autotidy.enable = true;
}
```

### Linux

```bash
curl -fsSL https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/linux/install.sh | sh
```

### Windows

```powershell
irm https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/windows/install.ps1 | iex
```

Additional installation information can be found [here](https://prettymuchbryce.github.io/autotidy/installation/).

## Configuration path

By default, rules are defined in the following locations:

| Platform | Default Path |
|----------|------|
| Linux | `~/.config/autotidy/config.yaml` |
| macOS | `~/.config/autotidy/config.yaml` |
| Windows | `%APPDATA%\autotidy\config.yaml` |

The [default rule file](internal/config/config-example.yaml) contains no enabled rules, but does include some commented-out examples.

## CLI Usage

```sh
autotidy status      Print status information
autotidy help        Show available commands
autotidy disable     Temporarily pause rule execution
autotidy enable      Resume rule execution (if it was previously disabled)
autotidy reload      Reload the daemon\'s rules from configuration
autotidy run         Perform a one-off run or dry run of rules
```

## Documentation

Full documentation is available [here](https://prettymuchbryce.github.io/autotidy/).

### Quick links
- [Quick Start](https://prettymuchbryce.github.io/autotidy/quick-start.html)
- [Configuration](https://prettymuchbryce.github.io/autotidy/configuration.html)
- [Filters](https://prettymuchbryce.github.io/autotidy/filters/)
- [Actions](https://prettymuchbryce.github.io/autotidy/actions/)
- [Installation](https://prettymuchbryce.github.io/autotidy/installation/)

## Inspired By

autotidy was inspired by tools like [Hazel](https://www.noodlesoft.com/), [Maid](https://github.com/maid/maid), and [organize](https://github.com/tfeldmann/organize).
