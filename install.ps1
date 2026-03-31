#Requires -Version 5.1
<#
.SYNOPSIS
    OctAi Windows Installer — installs all dependencies, Ollama, and OctAi.
.DESCRIPTION
    Downloads and installs OctAi on Windows with full dependency chain.
.EXAMPLE
    iwr -useb https://raw.githubusercontent.com/raynaythegreat/OctAI/master/install.ps1 | iex
.PARAMETER SkipDeps
    Skip installing system dependencies.
.PARAMETER SkipOllama
    Skip installing Ollama.
.PARAMETER Source
    Build from source instead of downloading a release.
#>

param(
    [switch]$SkipDeps,
    [switch]$SkipOllama,
    [switch]$Source
)

$ErrorActionPreference = "Stop"
$GitHubRepo = "raynaythegreat/OctAI"
$BinaryName = "octai.exe"
$LauncherName = "octai-launcher.exe"
$InstallDir = if ($env:OCTAI_INSTALL_DIR) { $env:OCTAI_INSTALL_DIR } else { "$env:USERPROFILE\.local\bin" }
$OctAiHome = if ($env:OCTAI_HOME) { $env:OCTAI_HOME } else { "$env:USERPROFILE\.octai" }

function Write-Info($msg)  { Write-Host "  [info] " -ForegroundColor Cyan -NoNewline; Write-Host $msg }
function Write-Ok($msg)    { Write-Host "  [ok] " -ForegroundColor Green -NoNewline; Write-Host $msg }
function Write-Warn($msg)  { Write-Host "  [warn] " -ForegroundColor Yellow -NoNewline; Write-Host $msg }
function Write-Err($msg)   { Write-Host "  [error] " -ForegroundColor Red -NoNewline; Write-Host $msg; exit 1 }

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x86_64" } elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "x86_64" }

Write-Host ""
Write-Host "  OctAi Windows Installer" -ForegroundColor Bold
Write-Host ""
Write-Info "Platform: windows/$arch"

$tmpDir = Join-Path $env:TEMP "octai-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
    # ── Dependencies ────────────────────────────────────────────────────────

    if (-not $SkipDeps) {
        Write-Host ""
        Write-Host "  Installing system dependencies..." -ForegroundColor Bold

        # Git
        if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
            Write-Info "Installing Git..."
            winget install --id Git.Git -e --source winget --accept-package-agreements --accept-source-agreements 2>$null
            if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
                Write-Warn "Git not installed. Download from https://git-scm.com/"
            }
        } else {
            Write-Ok "Git already installed"
        }

        # Go
        if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
            Write-Info "Installing Go 1.25..."
            $goArch = if ($arch -eq "arm64") { "arm64" } else { "amd64" }
            $goUrl = "https://go.dev/dl/go1.25.7.windows-${goArch}.zip"
            $goZip = Join-Path $tmpDir "go.zip"

            Write-Info "Downloading Go..."
            Invoke-WebRequest -Uri $goUrl -OutFile $goZip -UseBasicParsing

            Write-Info "Extracting to C:\Program Files\Go..."
            Expand-Archive -Path $goZip -DestinationPath "C:\Program Files" -Force
            $env:PATH = "C:\Program Files\Go\bin;$env:PATH"
            Write-Ok "Go installed"
        } else {
            Write-Ok "Go already installed ($(go version))"
        }

        # Node.js
        if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
            Write-Info "Installing Node.js 22 LTS..."
            winget install --id OpenJS.NodeJS.LTS -e --source winget --accept-package-agreements --accept-source-agreements 2>$null
            if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
                Write-Warn "Node.js not installed. Download from https://nodejs.org/"
            }
        } else {
            Write-Ok "Node.js already installed ($(node --version))"
        }

        # pnpm
        if (-not (Get-Command pnpm -ErrorAction SilentlyContinue)) {
            Write-Info "Installing pnpm..."
            npm install -g pnpm 2>$null
            Write-Ok "pnpm installed"
        } else {
            Write-Ok "pnpm already installed"
        }
    }

    # ── Ollama ──────────────────────────────────────────────────────────────

    if (-not $SkipOllama) {
        Write-Host ""
        Write-Host "  Installing Ollama (local LLM runner)..." -ForegroundColor Bold

        if (-not (Get-Command ollama -ErrorAction SilentlyContinue)) {
            Write-Info "Downloading Ollama..."
            winget install --id Ollama.Ollama -e --source winget --accept-package-agreements --accept-source-agreements 2>$null
            if (-not (Get-Command ollama -ErrorAction SilentlyContinue)) {
                Write-Warn "Ollama not installed. Download from https://ollama.com/download/windows"
            } else {
                Write-Ok "Ollama installed"
            }
        } else {
            Write-Ok "Ollama already installed"
        }
    }

    # ── Download / Build ────────────────────────────────────────────────────

    Write-Host ""
    Write-Host "  Installing OctAi..." -ForegroundColor Bold

    $tag = "latest"
    $builtFromSource = $false

    if (-not $Source) {
        try {
            $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$GitHubRepo/releases/latest" -UseBasicParsing
            $tag = $release.tag_name
            Write-Info "Latest release: $tag"

            $filename = "OctAi_Windows_${arch}.zip"
            $url = "https://github.com/$GitHubRepo/releases/download/$tag/$filename"
            $archive = Join-Path $tmpDir "octai.zip"

            Write-Info "Downloading OctAi $tag for windows/$arch..."
            Invoke-WebRequest -Uri $url -OutFile $archive -UseBasicParsing

            Write-Info "Extracting..."
            Expand-Archive -Path $archive -DestinationPath $tmpDir -Force

            $extracted = Get-ChildItem -Path $tmpDir -Directory -Filter "OctAi-*" | Select-Object -First 1
            if ($extracted) {
                foreach ($name in @($BinaryName, $LauncherName)) {
                    if (Test-Path "$($extracted.FullName)\$name") {
                        Move-Item "$($extracted.FullName)\$name" "$tmpDir\$name" -Force
                    }
                }
            }
        } catch {
            Write-Warn "Pre-built binary not found. Building from source..."
            $builtFromSource = $true
        }
    } else {
        $builtFromSource = $true
    }

    if ($builtFromSource) {
        if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
            Write-Err "Go is required to build from source. Install from https://go.dev/dl/"
        }

        Write-Info "Cloning repository..."
        & git clone --depth 1 "https://github.com/$GitHubRepo.git" "$tmpDir\OctAI" 2>$null

        Write-Info "Building $BinaryName..."
        Push-Location "$tmpDir\OctAI"
        & go build -mod=mod -tags "goolm,stdjson" -ldflags "-s -w" -o "$tmpDir\$BinaryName" ./cmd/aibhq
        Pop-Location

        Write-Info "Building frontend for launcher..."
        if ((Get-Command node -ErrorAction SilentlyContinue) -and (Get-Command pnpm -ErrorAction SilentlyContinue)) {
            Push-Location "$tmpDir\OctAI\web\frontend"
            & pnpm install 2>$null
            & pnpm run build:backend
            Pop-Location

            Push-Location "$tmpDir\OctAI"
            & go build -mod=mod -tags "goolm,stdjson" -ldflags "-s -w" -o "$tmpDir\$LauncherName" ./web/backend
            Pop-Location
        } else {
            Write-Warn "Node.js/pnpm not found — skipping launcher build."
        }
    }

    # ── Install ─────────────────────────────────────────────────────────────

    if (-not (Test-Path "$tmpDir\$BinaryName")) {
        Write-Err "Installation failed — binary not found"
    }

    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    foreach ($name in @($BinaryName, $LauncherName)) {
        if (Test-Path "$tmpDir\$name") {
            Move-Item "$tmpDir\$name" "$InstallDir\$name" -Force
            Write-Ok "Installed $name -> $InstallDir\$name"
        }
    }

    if (-not (Test-Path $OctAiHome)) {
        New-Item -ItemType Directory -Path $OctAiHome -Force | Out-Null
    }

    # PATH check
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
        Write-Warn "$InstallDir is not in your PATH."
        Write-Host ""
        Write-Host "  Add it permanently:" -ForegroundColor White
        Write-Host "    " -NoNewline
        Write-Host "[Environment]::SetEnvironmentVariable('Path', `"`$env:Path;$InstallDir`", 'User')" -ForegroundColor Cyan
        Write-Host "  Then restart your terminal." -ForegroundColor White
        Write-Host ""
        $env:PATH = "$InstallDir;$env:PATH"
    }

    Write-Host ""
    Write-Ok "OctAi installed successfully!"
    Write-Host ""
    Write-Host "  Next steps:" -ForegroundColor Bold
    Write-Host ""
    Write-Host "  1. Set up your AI provider:" -ForegroundColor Cyan
    Write-Host "     " -NoNewline; Write-Host "octai onboard" -ForegroundColor DarkGray -NoNewline; Write-Host "              # Interactive wizard"
    Write-Host "     " -NoNewline; Write-Host "octai auth login -p openai" -ForegroundColor DarkGray -NoNewline; Write-Host "  # OpenAI OAuth"
    Write-Host "     " -NoNewline; Write-Host "octai auth login -p anthropic --browser-oauth" -ForegroundColor DarkGray -NoNewline; Write-Host "  # Anthropic OAuth"
    Write-Host ""
    Write-Host "  2. Start using OctAi:" -ForegroundColor Cyan
    Write-Host "     " -NoNewline; Write-Host "octai serve" -ForegroundColor DarkGray -NoNewline; Write-Host "                # Web dashboard & gateway (http://localhost:18800)"
    Write-Host "     " -NoNewline; Write-Host "octai tui" -ForegroundColor DarkGray -NoNewline; Write-Host "                  # Terminal UI"
    Write-Host "     " -NoNewline; Write-Host "octai agent" -ForegroundColor DarkGray -NoNewline; Write-Host "                # AI chat session"
    Write-Host ""
    Write-Host "  3. Local models (Ollama):" -ForegroundColor Cyan
    Write-Host "     " -NoNewline; Write-Host "ollama pull llama3" -ForegroundColor DarkGray -NoNewline; Write-Host "         # Download a model"
    Write-Host "     " -NoNewline; Write-Host "ollama serve" -ForegroundColor DarkGray -NoNewline; Write-Host "               # Start Ollama server"
    Write-Host ""

} finally {
    Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
