# autotidy uninstallation script for Windows
# Usage: irm https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/windows/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

$taskName = "autotidy"
$installDir = "$env:LOCALAPPDATA\autotidy"

Write-Host "Uninstalling autotidy..." -ForegroundColor Yellow

# Stop and remove scheduled task
$existingTask = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
if ($existingTask) {
    if ($existingTask.State -eq "Running") {
        Stop-ScheduledTask -TaskName $taskName
        Write-Host "Stopped scheduled task"
    }
    Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
    Write-Host "Removed scheduled task: $taskName"
} else {
    Write-Host "No scheduled task found"
}

# Remove installation directory
if (Test-Path $installDir) {
    Remove-Item -Recurse -Force $installDir
    Write-Host "Removed directory: $installDir"
} else {
    Write-Host "No installation directory found"
}

Write-Host ""
Write-Host "Uninstallation complete!" -ForegroundColor Green
