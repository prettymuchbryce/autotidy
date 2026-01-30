# Test: Windows PowerShell installation
# Platform: Windows

$ErrorActionPreference = "Stop"

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-ErrorMessage {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

function Write-Success {
    param([string]$Message)
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  TEST PASSED: $Message" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
}

Write-Info "Testing: Windows PowerShell installation"

# Get paths
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoDir = Split-Path -Parent (Split-Path -Parent $ScriptDir)
$TestDir = Join-Path $env:TEMP "autotidy-test-$(Get-Random)"
$StartupDir = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup"
$ShortcutPath = "$StartupDir\autotidy.lnk"

Write-Info "Repository: $RepoDir"
Write-Info "Test directory: $TestDir"

# Create test directories
New-Item -ItemType Directory -Force -Path $TestDir | Out-Null
$ConfigDir = Join-Path $TestDir "config"
$WatchDir = Join-Path $TestDir "watch"
$DestDir = Join-Path $TestDir "dest"
New-Item -ItemType Directory -Force -Path $ConfigDir, $WatchDir, $DestDir | Out-Null

# Build
Write-Info "Building binary..."
Push-Location $RepoDir
try {
    go build -o autotidy.exe .
    if (-not (Test-Path "autotidy.exe")) {
        Write-ErrorMessage "Binary not found after build"
    }
}
finally {
    Pop-Location
}

$BinaryPath = Join-Path $RepoDir "autotidy.exe"
Write-Info "Built binary: $BinaryPath"

# Verify binary
Write-Info "Verifying binary..."
try {
    & $BinaryPath --help | Out-Null
    Write-Info "Binary verified successfully"
} catch {
    Write-ErrorMessage "Failed to run binary: $_"
}

# Create test config
$ConfigPath = Join-Path $ConfigDir "autotidy.yaml"
$ConfigContent = @"
rules:
  - name: test-rule
    locations:
      - $WatchDir
    filters:
      - extension: txt
    actions:
      - move:
          dest: $DestDir
"@
$ConfigContent | Out-File -FilePath $ConfigPath -Encoding utf8
Write-Info "Created config at: $ConfigPath"

# Install using the install script
Write-Info "Running install script..."
$InstallScript = Join-Path $RepoDir "install\windows\install.ps1"
& $InstallScript -BinaryPath $BinaryPath

# Check startup shortcut was created
if (-not (Test-Path $ShortcutPath)) {
    Write-ErrorMessage "Startup shortcut was not created at $ShortcutPath"
}
Write-Info "Startup shortcut created successfully"

# Check binary was installed
$InstalledBinary = "$env:LOCALAPPDATA\autotidy\autotidy.exe"
if (-not (Test-Path $InstalledBinary)) {
    Write-ErrorMessage "Binary was not installed at $InstalledBinary"
}
Write-Info "Binary installed successfully"

# Stop the daemon started by install script - we'll start our own with a custom config
Write-Info "Stopping daemon..."
Get-Process -Name "autotidy" -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 1

# For functional test, start daemon with test config
Write-Info "Starting daemon for functional test..."
$DaemonProcess = Start-Process -FilePath $InstalledBinary -ArgumentList "daemon", "--config", $ConfigPath -PassThru -WindowStyle Hidden

Start-Sleep -Seconds 2

if ($DaemonProcess.HasExited) {
    Write-ErrorMessage "Daemon process exited unexpectedly"
}
Write-Info "Daemon started with PID: $($DaemonProcess.Id)"

# Run functional test
Write-Info "Creating test file..."
$TestFile = Join-Path $WatchDir "test_file.txt"
"test content" | Out-File -FilePath $TestFile -Encoding utf8

$ExpectedDest = Join-Path $DestDir "test_file.txt"
Write-Info "Waiting for file to be moved..."

$MaxAttempts = 10
$Attempts = 0
while ($Attempts -lt $MaxAttempts) {
    if (Test-Path $ExpectedDest) {
        Write-Info "File successfully moved to: $ExpectedDest"
        break
    }
    Start-Sleep -Seconds 1
    $Attempts++
}

if (-not (Test-Path $ExpectedDest)) {
    Write-ErrorMessage "File was not moved within ${MaxAttempts}s"
}

# Cleanup
Write-Info "Cleaning up..."

# Stop daemon process
if (-not $DaemonProcess.HasExited) {
    Stop-Process -Id $DaemonProcess.Id -Force -ErrorAction SilentlyContinue
}

# Uninstall
$UninstallScript = Join-Path $RepoDir "install\windows\uninstall.ps1"
& $UninstallScript

# Verify uninstall
if (Test-Path $ShortcutPath) {
    Write-ErrorMessage "Startup shortcut was not removed"
}
if (Test-Path $InstalledBinary) {
    Write-ErrorMessage "Binary was not removed"
}
Write-Info "Uninstall verified"

# Clean test directory
Remove-Item -Recurse -Force $TestDir -ErrorAction SilentlyContinue

Write-Success "Windows PowerShell installation"
