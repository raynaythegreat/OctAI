#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="raynaythegreat"
REPO_NAME="OctAi"
GITHUB_REPO="raynaythegreat/OctAI"
BINARY_NAME="octai"
LAUNCHER_NAME="octai-launcher"
INSTALL_DIR="${OCTAI_INSTALL_DIR:-$HOME/.local/bin}"
OCTAI_HOME="${OCTAI_HOME:-$HOME/.octai}"

NO_COLOR="${NO_COLOR:-}"
if [ -z "$NO_COLOR" ]; then
  RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
  CYAN='\033[0;36m'; BOLD='\033[1m'; DIM='\033[2m'; RESET='\033[0m'
else
  RED='' GREEN='' YELLOW='' CYAN='' BOLD='' DIM='' RESET=''
fi

info()  { printf "${CYAN}  [info]${RESET} %s\n" "$1"; }
ok()    { printf "${GREEN}  [ok]${RESET} %s\n" "$1"; }
warn()  { printf "${YELLOW}  [warn]${RESET} %s\n" "$1"; }
err()   { printf "${RED}  [error]${RESET} %s\n" "$1" >&2; }
die()   { err "$1"; exit 1; }
step()  { printf "\n${BOLD}%s${RESET}\n" "$1"; }

cleanup() { [ -n "${TMPDIR:-}" ] && rm -rf "$TMPDIR"; }
trap cleanup EXIT

TMPDIR="$(mktemp -d)"

# ── OS / Arch Detection ──────────────────────────────────────────────────────────

detect_os() {
  local os
  os="$(uname -s)"
  case "$os" in
    Linux*)  echo "linux" ;;
    Darwin*) echo "darwin" ;;
    FreeBSD*) echo "freebsd" ;;
    NetBSD*) echo "netbsd" ;;
    MINGW*|MSYS*|CYGWIN*) echo "windows" ;;
    *)       die "Unsupported OS: $os" ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64)  echo "x86_64" ;;
    aarch64|arm64) echo "arm64" ;;
    armv7l|armv7)  echo "armv7" ;;
    armv6l|armv6)  echo "armv6" ;;
    riscv64)       echo "riscv64" ;;
    loongarch64)   echo "loong64" ;;
    *)             die "Unsupported architecture: $arch" ;;
  esac
}

detect_pkg_manager() {
  if command -v apt-get >/dev/null 2>&1; then echo "apt"
  elif command -v dnf >/dev/null 2>&1; then echo "dnf"
  elif command -v pacman >/dev/null 2>&1; then echo "pacman"
  elif command -v apk >/dev/null 2>&1; then echo "apk"
  elif command -v zypper >/dev/null 2>&1; then echo "zypper"
  elif command -v brew >/dev/null 2>&1; then echo "brew"
  else echo ""
  fi
}

download() {
  local url="$1" dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl --proto '=https' --tlsv1.2 -fsSL "$url" -o "$dest"
  elif command -v wget >/dev/null 2>&1; then
    wget --https-only --secure-protocol=TLSv1_2 -qO "$dest" "$url"
  else
    die "Neither curl nor wget found. Install one first."
  fi
}

# ── Dependency Installation ────────────────────────────────────────────────────

sudo_if_needed() {
  if [ "$(id -u)" -eq 0 ]; then
    "$@"
  elif command -v sudo >/dev/null 2>&1; then
    sudo "$@"
  else
    "$@"
  fi
}

install_build_deps_linux() {
  local pkgmgr
  pkgmgr="$(detect_pkg_manager)"

  step "Installing system dependencies via $pkgmgr..."

  case "$pkgmgr" in
    apt)
      sudo_if_needed apt-get update -qq
      sudo_if_needed apt-get install -y -qq ca-certificates curl git build-essential
      ;;
    dnf)
      sudo_if_needed dnf install -y ca-certificates curl git gcc make
      ;;
    pacman)
      sudo_if_needed pacman -Sy --noconfirm ca-certificates curl git base-devel
      ;;
    zypper)
      sudo_if_needed zypper install -y ca-certificates curl git gcc make
      ;;
    apk)
      sudo_if_needed apk add --no-progress ca-certificates curl git build-base
      ;;
    *)
      warn "Unknown package manager. Install ca-certificates, curl, git manually."
      ;;
  esac
}

install_build_deps_macos() {
  step "Checking macOS dependencies..."

  if ! command -v xcode-select >/dev/null 2>&1; then
    info "Installing Xcode Command Line Tools..."
    xcode-select --install 2>/dev/null || true
    warn "If prompted, accept the Xcode license in the GUI."
  fi

  if ! command -v brew >/dev/null 2>&1; then
    info "Installing Homebrew..."
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)" </dev/null

    if [ -f /opt/homebrew/bin/brew ]; then
      eval "$(/opt/homebrew/bin/brew shellenv)"
      printf "\n  ${DIM}Homebrew installed. You may need to add it to your PATH:${RESET}\n"
      printf "    ${CYAN}eval \"\$(/opt/homebrew/bin/brew shellenv)\"${RESET}\n\n"
    fi
  fi
}

install_build_deps_freebsd() {
  step "Installing FreeBSD dependencies..."
  sudo_if_needed pkg install -y ca-certificates curl git go node npm
}

install_go() {
  if command -v go >/dev/null 2>&1; then
    local go_version
    go_version="$(go version | grep -oP 'go\d+\.\d+' | head -1)"
    ok "Go already installed ($go_version)"
    return 0
  fi

  step "Installing Go 1.25..."
  local os arch go_url
  os="$(detect_os)"
  arch="$(detect_arch)"
  go_url="https://go.dev/dl/go1.25.7.${os}-${arch}.tar.gz"

  case "$os" in
    linux)   go_url="https://go.dev/dl/go1.25.7.linux-${arch}.tar.gz" ;;
    darwin)  go_url="https://go.dev/dl/go1.25.7.darwin-${arch}.tar.gz" ;;
    freebsd) go_url="https://go.dev/dl/go1.25.7.freebsd-${arch}.tar.gz" ;;
    *)       die "Cannot determine Go download URL for $os" ;;
  esac

  info "Downloading Go..."
  download "$go_url" "${TMPDIR}/go.tar.gz"

  info "Extracting to /usr/local..."
  sudo_if_needed rm -rf /usr/local/go
  sudo_if_needed tar -C /usr/local -xzf "${TMPDIR}/go.tar.gz"

  export PATH="/usr/local/go/bin:$PATH"
  ok "Go installed ($(go version))"
}

install_node() {
  if command -v node >/dev/null 2>&1; then
    local node_version
    node_version="$(node --version 2>/dev/null)"
    ok "Node.js already installed ($node_version)"
    return 0
  fi

  step "Installing Node.js..."
  local pkgmgr
  pkgmgr="$(detect_pkg_manager)"
  local os
  os="$(detect_os)"

  case "$os" in
    darwin)
      if command -v brew >/dev/null 2>&1; then
        brew install node
      else
        die "Homebrew required to install Node.js on macOS. Run the script again after installing Homebrew."
      fi
      ;;
    linux)
      case "$pkgmgr" in
        apt)
          curl -fsSL https://deb.nodesource.com/setup_22.x | sudo_if_needed bash -
          sudo_if_needed apt-get install -y nodejs
          ;;
        dnf)
          sudo_if_needed dnf install -y nodejs npm
          ;;
        pacman)
          sudo_if_needed pacman -S --noconfirm nodejs npm
          ;;
        apk)
          sudo_if_needed apk add --no-progress nodejs npm
          ;;
        *)
          warn "Cannot auto-install Node.js. Install Node.js 22+ manually from https://nodejs.org/"
          return 1
          ;;
      esac
      ;;
    freebsd)
      sudo_if_needed pkg install -y node npm
      ;;
    *)
      warn "Cannot auto-install Node.js. Install from https://nodejs.org/"
      return 1
      ;;
  esac

  ok "Node.js installed ($(node --version))"
}

install_pnpm() {
  if command -v pnpm >/dev/null 2>&1; then
    ok "pnpm already installed ($(pnpm --version))"
    return 0
  fi

  step "Installing pnpm..."
  npm install -g pnpm 2>/dev/null || corepack enable && corepack prepare pnpm@latest --activate
  ok "pnpm installed"
}

install_ollama() {
  if command -v ollama >/dev/null 2>&1; then
    ok "Ollama already installed"
    return 0
  fi

  step "Installing Ollama (local LLM runner)..."
  local os arch
  os="$(detect_os)"
  arch="$(detect_arch)"

  case "$os" in
    linux)
      local ollama_url
      case "$arch" in
        x86_64) ollama_url="https://ollama.com/download/ollama-linux-amd64.tgz" ;;
        arm64)  ollama_url="https://ollama.com/download/ollama-linux-arm64.tgz" ;;
        *)      warn "Ollama does not have a prebuilt binary for $arch on Linux. Install manually from https://ollama.com"
                return 1 ;;
      esac
      info "Downloading Ollama..."
      download "$ollama_url" "${TMPDIR}/ollama.tgz"
      sudo_if_needed tar -C /usr -xzf "${TMPDIR}/ollama.tgz"
      ;;
    darwin)
      if command -v brew >/dev/null 2>&1; then
        brew install ollama
      else
        info "Download Ollama from https://ollama.com/download/mac"
        warn "Ollama not installed. You can install it later from https://ollama.com"
        return 1
      fi
      ;;
    *)
      warn "Ollama auto-install not supported on $os. Install manually from https://ollama.com"
      return 1
      ;;
  esac

  ok "Ollama installed"
}

install_uv() {
  if command -v uv >/dev/null 2>&1 || command -v uvx >/dev/null 2>&1; then
    ok "uv/uvx already installed"
    return 0
  fi

  step "Installing uv (Python tool runner for MCP servers)..."
  curl -LsSf https://astral.sh/uv/install.sh 2>/dev/null | sh 2>/dev/null || {
    warn "uv install failed. MCP servers requiring Python may not work."
    return 1
  }

  export PATH="$HOME/.local/bin:$PATH"
  ok "uv installed"
}

# ── Binary Download / Build ────────────────────────────────────────────────────

get_latest_release() {
  local api_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
  local tag
  tag="$(download "$api_url" - 2>/dev/null | grep '"tag_name"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')"
  if [ -z "$tag" ]; then
    warn "Could not determine latest release, building from source"
    echo "source"
    return
  fi
  echo "$tag"
}

resolve_download_url() {
  local os="$1" arch="$2" tag="$3"
  local filename="OctAi_${os^}_${arch}.tar.gz"
  case "$os" in
    linux)   filename="OctAi_Linux_${arch}.tar.gz" ;;
    darwin)  filename="OctAi_Darwin_${arch}.tar.gz" ;;
    freebsd) filename="OctAi_FreeBSD_${arch}.tar.gz" ;;
  esac
  echo "https://github.com/${GITHUB_REPO}/releases/download/${tag}/${filename}"
}

install_prebuilt() {
  local os="$1" arch="$2" tag="$3"
  local url archive

  url="$(resolve_download_url "$os" "$arch" "$tag")"
  archive="${TMPDIR}/octai.tar.gz"

  info "Downloading OctAi ${tag} for ${os}/${arch}..."
  if ! download "$url" "$archive" 2>/dev/null; then
    warn "Pre-built binary not found for ${os}/${arch}. Building from source..."
    install_from_source
    return
  fi

  info "Extracting..."
  tar -xzf "$archive" -C "$TMPDIR"

  local extracted
  extracted="$(find "$TMPDIR" -maxdepth 1 -type d -name "OctAi-*" | head -1)"
  if [ -n "$extracted" ]; then
    for name in "$BINARY_NAME" "$LAUNCHER_NAME"; do
      for ext in "" ".exe"; do
        if [ -f "${extracted}/${name}${ext}" ]; then
          mv "${extracted}/${name}${ext}" "${TMPDIR}/${name}${ext}"
        fi
      done
    done
  fi

  if [ -f "${TMPDIR}/${BINARY_NAME}" ] || [ -f "${TMPDIR}/${BINARY_NAME}.exe" ]; then
    ok "Pre-built binary extracted"
  else
    warn "Binary not found in archive. Building from source..."
    install_from_source
  fi
}

install_from_source() {
  info "Building OctAi from source..."

  command -v go >/dev/null 2>&1 || die "Go is required. Install from https://go.dev/dl/"

  local src_dir="${TMPDIR}/${REPO_NAME}"
  if [ ! -d "$src_dir" ]; then
    info "Cloning repository..."
    git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$src_dir" 2>/dev/null || \
      git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$src_dir"
  fi

  local go_tags="goolm,stdjson"
  local goarch="$(uname -m)"

  case "$goarch" in
    mips*|loongarch64|riscv64|s390x)
      go_tags="stdjson"
      info "Excluding goolm tag for $goarch (libc compatibility)"
      ;;
  esac

  info "Building ${BINARY_NAME} (tags: ${go_tags})..."
  (cd "$src_dir" && go build -mod=mod -tags "$go_tags" -ldflags "-s -w" -o "${TMPDIR}/${BINARY_NAME}" ./cmd/aibhq)

  info "Building frontend for launcher..."
  if command -v node >/dev/null 2>&1 && command -v pnpm >/dev/null 2>&1; then
    (cd "$src_dir/web/frontend" && pnpm install --frozen-lockfile 2>/dev/null || pnpm install && pnpm run build:backend)
    (cd "$src_dir" && go build -mod=mod -tags "$go_tags" -ldflags "-s -w" -o "${TMPDIR}/${LAUNCHER_NAME}" ./web/backend)
  else
    warn "Node.js/pnpm not found — skipping launcher build. Run 'octai web' for the web console instead."
  fi
}

# ── PATH Management ────────────────────────────────────────────────────────────

ensure_path() {
  case ":$PATH:" in
    *":${INSTALL_DIR}:"*) return 0 ;;
  esac

  warn "${INSTALL_DIR} is not in your PATH."
  printf "\n  Add it to your shell profile:\n\n"

  local shell_profile=""
  if [ -n "${ZSH_VERSION:-}" ] || [ -f "$HOME/.zshrc" ]; then
    shell_profile="$HOME/.zshrc"
  else
    shell_profile="$HOME/.bashrc"
  fi

  printf "    ${CYAN}echo 'export PATH=\"%s:\$PATH\"' >> %s${RESET}\n" "$INSTALL_DIR" "$shell_profile"
  printf "    ${CYAN}source %s${RESET}\n\n" "$shell_profile"

  export PATH="${INSTALL_DIR}:${PATH}"
}

# ── Main ───────────────────────────────────────────────────────────────────────

main() {
  local skip_deps=false skip_ollama=false skip_uv=false tag="latest"

  for arg in "$@"; do
    case "$arg" in
      --skip-deps)    skip_deps=true ;;
      --skip-ollama)  skip_ollama=true ;;
      --skip-uv)      skip_uv=true ;;
      --source)       tag="source" ;;
      latest)         tag="latest" ;;
      v*)             tag="$arg" ;;
      *)              tag="$arg" ;;
    esac
  done

  printf "\n"
  printf "${BOLD}  ██████╗  ██████╗████████╗ █████╗ ██╗${RESET}\n"
  printf "${BOLD} ██╔═══██╗██╔════╝╚══██╔══╝██╔══██╗██║${RESET}\n"
  printf "${BOLD} ██║   ██║██║        ██║   ███████║██║${RESET}\n"
  printf "${BOLD} ╚██████╔╝╚██████╗   ██║   ██╔══██║██║${RESET}\n"
  printf "${BOLD}  ╚═════╝  ╚═════╝   ╚═╝   ╚═╝  ╚═╝╚═╝${RESET}\n"
  printf "\n"

  local os arch
  os="$(detect_os)"
  arch="$(detect_arch)"
  info "Platform: ${os}/${arch}"

  if [ "$skip_deps" = false ]; then
    case "$os" in
      linux)   install_build_deps_linux ;;
      darwin)  install_build_deps_macos ;;
      freebsd) install_build_deps_freebsd ;;
    esac

    install_go
    install_node
    install_pnpm
  fi

  if [ "$skip_ollama" = false ]; then
    install_ollama || true
  fi

  if [ "$skip_uv" = false ]; then
    install_uv || true
  fi

  mkdir -p "$INSTALL_DIR"

  if [ "$tag" = "source" ]; then
    install_from_source
  else
    local release_tag
    release_tag="$(get_latest_release)"
    if [ "$release_tag" = "source" ]; then
      install_from_source
    else
      install_prebuilt "$os" "$arch" "$release_tag"
    fi
  fi

  for name in "$BINARY_NAME" "$LAUNCHER_NAME"; do
    for ext in "" ".exe"; do
      if [ -f "${TMPDIR}/${name}${ext}" ]; then
        chmod +x "${TMPDIR}/${name}${ext}"
        mv "${TMPDIR}/${name}${ext}" "${INSTALL_DIR}/${name}${ext}"
        ok "Installed ${name}${ext} → ${INSTALL_DIR}/${name}${ext}"
      fi
    done
  done

  if [ ! -f "${INSTALL_DIR}/${BINARY_NAME}" ] && [ ! -f "${INSTALL_DIR}/${BINARY_NAME}.exe" ]; then
    die "Installation failed — binary not found"
  fi

  mkdir -p "$OCTAI_HOME"
  ensure_path

  printf "\n"
  ok "OctAi installed successfully!"
  printf "\n"
  printf "  ${BOLD}Next steps:${RESET}\n"
  printf "\n"
  printf "  ${CYAN}1. Set up your AI provider:${RESET}\n"
  printf "     ${DIM}octai onboard${RESET}              ${DIM}# Interactive wizard${RESET}\n"
  printf "     ${DIM}octai auth login -p openai${RESET}  ${DIM}# OpenAI OAuth${RESET}\n"
  printf "     ${DIM}octai auth login -p anthropic --browser-oauth${RESET}  ${DIM}# Anthropic OAuth${RESET}\n"
  printf "\n"
  printf "  ${CYAN}2. Start using OctAi:${RESET}\n"
  printf "     ${DIM}octai web${RESET}                  ${DIM}# Web dashboard (http://localhost:18800)${RESET}\n"
  printf "     ${DIM}octai tui${RESET}                  ${DIM}# Terminal UI${RESET}\n"
  printf "     ${DIM}octai agent${RESET}                ${DIM}# AI chat session${RESET}\n"
  printf "\n"
  printf "  ${CYAN}3. Local models (Ollama):${RESET}\n"
  printf "     ${DIM}ollama pull llama3${RESET}         ${DIM}# Download a model${RESET}\n"
  printf "     ${DIM}ollama serve${RESET}               ${DIM}# Start Ollama server${RESET}\n"
  printf "\n"
}

main "$@"
