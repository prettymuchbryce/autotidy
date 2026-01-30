# autotidy

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="assets/icon-dark-128.png">
    <source media="(prefers-color-scheme: light)" srcset="assets/icon-light-128.png">
    <img src="assets/icon-light-128.png" alt="autotidy" width="128" height="128">
  </picture>
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

**autotidy** consists of both a background process (daemon) and a CLI tool. It allows you to write declarative rules that are applied automatically when files in a directory change. Rules are defined in a YAML file. Rules specify the directories that should be watched for changes and the actions that should be performed when files in those directories change. Actions include moving, copying, renaming, and deleting files. Additionally, filters can be applied to target only files that meet certain criteria, such as a particular name, extension, MIME type, or size.

autotidy aims to be cross-platform, though Windows support is currently experimental.

## Configuration example

Rules are defined in a `config.yaml` file. The below configuration contains an example rule which organizes the `~/Downloads` directory.

```yaml
# config.yaml
rules:
  - name: Organize Downloaded Images
    locations: ~/Downloads
    filters:
      - mime_type: "image/*"
    actions:
      - move: ~/Pictures/Downloads
```

When the contents of the `~/Downloads` directory change, image files in this directory will be moved to `~/Pictures/Downloads`.

You can find an exhaustive list of configuration options along with more examples [here](https://prettymuchbryce.github.io/autotidy/configuration.html).

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

Additional installation options can be found [here](https://prettymuchbryce.github.io/autotidy/installation/).

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
- [Getting Started](https://prettymuchbryce.github.io/autotidy/getting-started.html)
- [Configuration](https://prettymuchbryce.github.io/autotidy/configuration.html)
- [Filters](https://prettymuchbryce.github.io/autotidy/filters/)
- [Actions](https://prettymuchbryce.github.io/autotidy/actions/)
- [Installation](https://prettymuchbryce.github.io/autotidy/installation/)

## Inspired By

autotidy was inspired by tools like [Hazel](https://www.noodlesoft.com/), [Maid](https://github.com/maid/maid), and [organize](https://github.com/tfeldmann/organize).
