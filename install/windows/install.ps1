# autotidy installation script for Windows
# Usage: irm https://raw.githubusercontent.com/prettymuchbryce/autotidy/master/install/windows/install.ps1 | iex
#
# For testing with a local binary:
#   .\install.ps1 -BinaryPath .\autotidy.exe

param(
    [string]$BinaryPath = ""
)

$ErrorActionPreference = "Stop"

$repo = "prettymuchbryce/autotidy"
$installDir = "$env:LOCALAPPDATA\autotidy"
$binPath = "$installDir\autotidy.exe"
$startupDir = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup"
$shortcutPath = "$startupDir\autotidy.lnk"

Write-Host "Installing autotidy..." -ForegroundColor Green

# Create install directory
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
    Write-Host "Created directory: $installDir"
}

if ($BinaryPath -ne "") {
    # Use local binary (for testing)
    if (-not (Test-Path $BinaryPath)) {
        Write-Error "Binary not found at $BinaryPath"
        exit 1
    }
    Copy-Item $BinaryPath $binPath -Force
    Write-Host "Installed autotidy from local binary"
    $version = "local"
} else {
    # Download from GitHub releases
    Write-Host "Fetching latest release..."
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
    $version = $release.tag_name

    # Find Windows amd64 asset
    $asset = $release.assets | Where-Object { $_.name -like "*windows_amd64.zip" } | Select-Object -First 1
    if (-not $asset) {
        Write-Error "Could not find Windows release asset"
        exit 1
    }

    Write-Host "Downloading autotidy $version..."
    $zipPath = "$env:TEMP\autotidy.zip"
    Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $zipPath

    # Extract binary
    Write-Host "Extracting..."
    Expand-Archive -Path $zipPath -DestinationPath $env:TEMP -Force
    Move-Item -Path "$env:TEMP\autotidy.exe" -Destination $binPath -Force
    Remove-Item $zipPath -Force

    Write-Host "Installed autotidy $version to $binPath"
}

# Stop any running autotidy process
Get-Process -Name "autotidy" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue

# Remove existing startup shortcut if present
if (Test-Path $shortcutPath) {
    Remove-Item $shortcutPath -Force
}

# Create startup shortcut
$WshShell = New-Object -ComObject WScript.Shell
$shortcut = $WshShell.CreateShortcut($shortcutPath)
$shortcut.TargetPath = $binPath
$shortcut.Arguments = "daemon"
$shortcut.WorkingDirectory = $installDir
$shortcut.WindowStyle = 7  # Minimized
$shortcut.Save()

Write-Host "Created startup shortcut"

# Start the daemon now
Write-Host "Starting autotidy daemon..."
Start-Process -FilePath $binPath -ArgumentList "daemon" -WindowStyle Hidden

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host ""
Write-Host "autotidy $version is now running."
Write-Host ""
Write-Host "To check status:  & `"$binPath`" status"
Write-Host "To view config:   notepad $env:APPDATA\autotidy\config.yaml"
