#Requires -Version 5.1
<#
.SYNOPSIS
    OctAi Windows Installer
.DESCRIPTION
    Downloads and installs OctAi on Windows.
.EXAMPLE
    iwr -useb https://raw.githubusercontent.com/raynaythegreat/OctAI/master/install.ps1 | iex
#>

$ErrorActionPreference = "Stop"
$GitHubRepo = "raynaythegreat/OctAI"
$BinaryName = "octai.exe"
$InstallDir = if ($env:OCTAI_INSTALL_DIR) { $env:OCTAI_INSTALL_DIR } else { "$env:USERPROFILE\.local\bin" }

function Write-Info($msg)  { Write-Host "  [info] " -ForegroundColor Cyan -NoNewline; Write-Host $msg }
function Write-Ok($msg)    { Write-Host "  [ok] " -ForegroundColor Green -NoNewline; Write-Host $msg }
function Write-Warn($msg)  { Write-Host "  [warn] " -ForegroundColor Yellow -NoNewline; Write-Host $msg }
function Write-Err($msg)   { Write-Host "  [error] " -ForegroundColor Red -NoNewline; Write-Host $msg }

$os = "windows"
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x86_64" } elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "x86_64" }

Write-Host ""
Write-Host "  OctAi Windows Installer" -ForegroundColor Bold
Write-Host ""
Write-Info "Detected: ${os}/${arch}"

$tag = "latest"
try {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$GitHubRepo/releases/latest" -UseBasicParsing
    $tag = $release.tag_name
    Write-Info "Latest release: $tag"
} catch {
    Write-Warn "Could not determine latest release, building from source"
    $tag = "master"
}

$tmpDir = Join-Path $env:TEMP "octai-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
    if ($tag -ne "master") {
        $filename = "OctAi_Windows_${arch}.zip"
        $url = "https://github.com/$GitHubRepo/releases/download/$tag/$filename"

        Write-Info "Downloading OctAi $tag..."
        try {
            Invoke-WebRequest -Uri $url -OutFile "$tmpDir\octai.zip" -UseBasicParsing
        } catch {
            Write-Warn "Pre-built binary not found. Building from source..."
            $tag = "master"
        }
    }

    if ($tag -eq "master") {
        if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
            Write-Err "Go is required to build from source. Install from https://go.dev/dl/"
            exit 1
        }
        if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
            Write-Err "Git is required. Install from https://git-scm.com/"
            exit 1
        }

        Write-Info "Cloning repository..."
        & git clone --depth 1 "https://github.com/$GitHubRepo.git" "$tmpDir\OctAI" 2>$null

        Write-Info "Building $BinaryName..."
        Push-Location "$tmpDir\OctAI"
        & go build -mod=mod -tags "goolm,stdjson" -o "$tmpDir\$BinaryName" ./cmd/aibhq
        Pop-Location
    } else {
        Write-Info "Extracting..."
        Expand-Archive -Path "$tmpDir\octai.zip" -DestinationPath $tmpDir -Force

        $extracted = Get-ChildItem -Path $tmpDir -Directory -Filter "OctAi-*" | Select-Object -First 1
        if ($extracted -and (Test-Path "$($extracted.FullName)\$BinaryName")) {
            Move-Item "$($extracted.FullName)\$BinaryName" "$tmpDir\$BinaryName" -Force
        } elseif ($extracted -and (Test-Path "$($extracted.FullName)\octai.exe")) {
            Move-Item "$($extracted.FullName)\octai.exe" "$tmpDir\$BinaryName" -Force
        }
    }

    if (-not (Test-Path "$tmpDir\$BinaryName")) {
        Write-Err "Build failed — binary not found"
        exit 1
    }

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    Move-Item "$tmpDir\$BinaryName" "$InstallDir\$BinaryName" -Force
    Write-Ok "Installed to $InstallDir\$BinaryName"

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
        Write-Warn "$InstallDir is not in your PATH."
        Write-Host ""
        Write-Host "  Add it:" -ForegroundColor White
        Write-Host "    " -NoNewline
        Write-Host "[Environment]::SetEnvironmentVariable('Path', `"`$env:Path;$InstallDir`", 'User')" -ForegroundColor Cyan
        Write-Host "  Then restart your terminal." -ForegroundColor White
        Write-Host ""
    }

    Write-Host ""
    Write-Ok "OctAi installed successfully!"
    Write-Host ""
    Write-Host "  Getting started:" -ForegroundColor Bold
    Write-Host "    " -NoNewline; Write-Host "octai onboard" -ForegroundColor Cyan -NoNewline; Write-Host "          Interactive setup wizard"
    Write-Host "    " -NoNewline; Write-Host "octai web" -ForegroundColor Cyan -NoNewline; Write-Host "              Start web dashboard"
    Write-Host "    " -NoNewline; Write-Host "octai tui" -ForegroundColor Cyan -NoNewline; Write-Host "              Start terminal UI"
    Write-Host "    " -NoNewline; Write-Host "octai agent" -ForegroundColor Cyan -NoNewline; Write-Host "            Start AI chat session"
    Write-Host ""
} finally {
    Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
