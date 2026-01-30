# Windows

> **Note:** Windows support is experimental. [Please open an issue](https://github.com/prettymuchbryce/autotidy/issues) if you run into problems.

## Install

```powershell
irm https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/windows/install.ps1 | iex
```

This downloads the latest release, installs it to `%LOCALAPPDATA%\autotidy`, registers it to start at login, and starts the daemon.

## Verify

```powershell
& "$env:LOCALAPPDATA\autotidy\autotidy.exe" status
```

## Service management

```powershell
# Check if running
Get-Process -Name "autotidy" -ErrorAction SilentlyContinue

# Stop
Stop-Process -Name "autotidy" -Force

# Start
Start-Process -FilePath "$env:LOCALAPPDATA\autotidy\autotidy.exe" -ArgumentList "daemon" -WindowStyle Hidden
```

## Uninstall

```powershell
irm https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/windows/uninstall.ps1 | iex
```
