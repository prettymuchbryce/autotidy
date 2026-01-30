# autotidy uninstallation script for Windows
# Usage: irm https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/windows/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$installDir = "$env:LOCALAPPDATA\autotidy"
$registryPath = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Run"
$registryName = "autotidy"

Write-Host "Uninstalling autotidy..." -ForegroundColor Yellow

# Stop any running autotidy process
$process = Get-Process -Name "autotidy" -ErrorAction SilentlyContinue
if ($process) {
    Stop-Process -Name "autotidy" -Force -ErrorAction SilentlyContinue
    Write-Host "Stopped autotidy process"
}

# Remove startup registry entry
$existingValue = Get-ItemProperty -Path $registryPath -Name $registryName -ErrorAction SilentlyContinue
if ($existingValue) {
    Remove-ItemProperty -Path $registryPath -Name $registryName -ErrorAction SilentlyContinue
    Write-Host "Removed startup registry entry"
}

# Remove old startup shortcut if present (from previous install method)
$oldShortcutPath = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup\autotidy.lnk"
if (Test-Path $oldShortcutPath) {
    Remove-Item $oldShortcutPath -Force
    Write-Host "Removed old startup shortcut"
}

# Remove installation directory
if (Test-Path $installDir) {
    Remove-Item -Recurse -Force $installDir
    Write-Host "Removed directory: $installDir"
}

Write-Host ""
Write-Host "Uninstallation complete!" -ForegroundColor Green
