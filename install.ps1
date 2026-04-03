$ErrorActionPreference = "Stop"

# Repository metadata for release downloads.
$Repo = "bentos-lab/parley"
$AppName = "parley"
$InstallDir = "$env:USERPROFILE\bin"

# Windows-only installer, so the OS is fixed.
$OS = "windows"

# Map the processor architecture to the release artifact naming.
switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { $Arch = "amd64" }
    "ARM64" { $Arch = "arm64" }
    default {
        Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
        exit 1
    }
}

Write-Host "Detect: $OS/$Arch"

# Pull the latest release metadata from GitHub.
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Version = $Release.tag_name
Write-Host "Latest version: $Version"

# Build download URL for the Windows asset.
$FileName = "$AppName-$Version-$OS-$Arch.zip"
$Url = "https://github.com/$Repo/releases/download/$Version/$FileName"

# Create a temp dir for downloading and extraction.
$TempDir = New-Item -ItemType Directory -Path ([System.IO.Path]::GetTempPath()) -Name ("parley_" + [System.Guid]::NewGuid())
$ZipPath = Join-Path $TempDir $FileName

Write-Host "Downloading $FileName..."
Invoke-WebRequest -Uri $Url -OutFile $ZipPath

Write-Host "Extracting..."
Expand-Archive -Path $ZipPath -DestinationPath $TempDir -Force

# Ensure the install directory exists and move the binary.
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

Move-Item -Path (Join-Path $TempDir "$AppName.exe") -Destination (Join-Path $InstallDir "$AppName.exe") -Force

# Add the install directory to the user PATH if missing.
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
    Write-Host "Added $InstallDir to PATH (restart terminal to apply)"
}

# Clean up temp files and print success message.
Remove-Item $TempDir -Recurse -Force
Write-Host ""
Write-Host "Installed parley $Version!"
Write-Host "Run: $AppName --version"
