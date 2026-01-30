# Linux

## Install

```bash
curl -fsSL https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/linux/install.sh | sh
```

This downloads the latest release, installs it to `~/.local/bin`, sets up a systemd user service, and starts the daemon.

## Verify

```bash
autotidy status
```

> **Note:** If `autotidy` is not recognized, restart your shell or run: `export PATH="$HOME/.local/bin:$PATH"`

## Service management

```bash
# Check status
systemctl --user status autotidy

# View logs
journalctl --user -u autotidy -f

# Restart
systemctl --user restart autotidy

# Stop
systemctl --user stop autotidy
```

## Uninstall

```bash
curl -fsSL https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/linux/uninstall.sh | sh
```

## Alternative: deb/rpm packages

You can also install via `.deb` or `.rpm` packages from [GitHub Releases](https://github.com/prettymuchbryce/autotidy/releases):

```bash
# Debian/Ubuntu
sudo dpkg -i autotidy_*.deb
systemctl --user enable --now autotidy

# Fedora/RHEL
sudo rpm -i autotidy-*.rpm
systemctl --user enable --now autotidy
```
