# macOS

## Homebrew

```bash
brew install prettymuchbryce/tap/autotidy
brew services start autotidy
```

## Verify

```bash
autotidy status
```

## Service management

```bash
# Stop
brew services stop autotidy

# Restart
brew services restart autotidy

# View logs
tail -f /tmp/autotidy.out.log
tail -f /tmp/autotidy.err.log
```

## Uninstall

```bash
brew services stop autotidy
brew uninstall autotidy
```
