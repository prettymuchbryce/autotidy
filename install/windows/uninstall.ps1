# autotidy uninstallation script for Windows
# Usage: irm https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/windows/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$installDir = "$env:LOCALAPPDATA\autotidy"
$startupDir = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup"
$shortcutPath = "$startupDir\autotidy.lnk"

Write-Host "Uninstalling autotidy..." -ForegroundColor Yellow

# Stop any running autotidy process
$process = Get-Process -Name "autotidy" -ErrorAction SilentlyContinue
if ($process) {
    Stop-Process -Name "autotidy" -Force -ErrorAction SilentlyContinue
    Write-Host "Stopped autotidy process"
}

# Remove startup shortcut
if (Test-Path $shortcutPath) {
    Remove-Item $shortcutPath -Force
    Write-Host "Removed startup shortcut"
}

# Remove installation directory
if (Test-Path $installDir) {
    Remove-Item -Recurse -Force $installDir
    Write-Host "Removed directory: $installDir"
}

Write-Host ""
Write-Host "Uninstallation complete!" -ForegroundColor Green
