#!/usr/bin/env bash
set -euo pipefail

REPO_OWNER="raynaythegreat"
REPO_NAME="OctAI"
GITHUB_REPO="raynaythegreat/OctAI"
BINARY_NAME="octai"
INSTALL_DIR="${OCTAI_INSTALL_DIR:-$HOME/.local/bin}"

NO_COLOR="${NO_COLOR:-}"
if [ -z "$NO_COLOR" ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  CYAN='\033[0;36m'
  BOLD='\033[1m'
  RESET='\033[0m'
else
  RED='' GREEN='' YELLOW='' CYAN='' BOLD='' RESET=''
fi

info()  { printf "${CYAN}  [info]${RESET} %s\n" "$1"; }
ok()    { printf "${GREEN}  [ok]${RESET} %s\n" "$1"; }
warn()  { printf "${YELLOW}  [warn]${RESET} %s\n" "$1"; }
err()   { printf "${RED}  [error]${RESET} %s\n" "$1" >&2; }
die()   { err "$1"; exit 1; }

cleanup() {
  [ -n "${TMPDIR:-}" ] && rm -rf "$TMPDIR"
}
trap cleanup EXIT

TMPDIR="$(mktemp -d)"

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
    mipsel|mipsle) echo "mipsle" ;;
    s390x)         echo "s390x" ;;
    *)             die "Unsupported architecture: $arch" ;;
  esac
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

get_latest_release() {
  local api_url="https://api.github.com/repos/${GITHUB_REPO}/releases/latest"
  local tag
  tag="$(download "$api_url" - 2>/dev/null | grep '"tag_name"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')"
  if [ -z "$tag" ]; then
    warn "Could not determine latest release, falling back to master"
    tag="master"
  fi
  echo "$tag"
}

resolve_download_url() {
  local os="$1" arch="$2" tag="$3"
  local ext="tar.gz"

  if [ "$os" = "windows" ]; then
    ext="zip"
  fi

  if [ "$tag" = "master" ]; then
    echo "https://github.com/${GITHUB_REPO}/archive/refs/heads/master.${ext}"
    return
  fi

  local filename
  case "$os" in
    linux)   filename="OctAi_Linux_${arch}" ;;
    darwin)  filename="OctAi_Darwin_${arch}" ;;
    freebsd) filename="OctAi_FreeBSD_${arch}" ;;
    netbsd)  filename="OctAi_NetBSD_${arch}" ;;
    *)       die "Cannot determine download filename for OS: $os" ;;
  esac

  echo "https://github.com/${GITHUB_REPO}/releases/download/${tag}/${filename}.${ext}"
}

install_from_source() {
  info "Building from source..."

  command -v go >/dev/null 2>&1 || die "Go is required to build from source. Install from https://go.dev/dl/"
  command -v git >/dev/null 2>&1 || die "Git is required. Install from https://git-scm.com/"

  local src_dir="${TMPDIR}/${REPO_NAME}"
  if [ ! -d "$src_dir" ]; then
    info "Cloning repository..."
    git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$src_dir" 2>/dev/null || \
      git clone --depth 1 "https://github.com/${GITHUB_REPO}.git" "$src_dir"
  fi

  info "Building ${BINARY_NAME}..."
  (cd "$src_dir" && go build -mod=mod -tags "goolm,stdjson" -o "${TMPDIR}/${BINARY_NAME}" ./cmd/aibhq)
}

install_prebuilt() {
  local os="$1" arch="$2" tag="$3"

  local url
  url="$(resolve_download_url "$os" "$arch" "$tag")"

  local archive="${TMPDIR}/octai.${url##*.}"

  info "Downloading OctAi ${tag} for ${os}/${arch}..."
  download "$url" "$archive" || {
    warn "Pre-built binary not found for ${os}/${arch}. Building from source..."
    install_from_source
    return
  }

  info "Extracting..."
  case "$archive" in
    *.tar.gz) tar -xzf "$archive" -C "$TMPDIR" ;;
    *.zip)     unzip -qo "$archive" -d "$TMPDIR" ;;
  esac

  local extracted_dir
  extracted_dir="$(find "$TMPDIR" -maxdepth 1 -type d -name "OctAi-*" -o -name "${REPO_NAME}-*" | head -1)"
  if [ -n "$extracted_dir" ] && [ -f "${extracted_dir}/${BINARY_NAME}" ]; then
    mv "${extracted_dir}/${BINARY_NAME}" "${TMPDIR}/${BINARY_NAME}"
  elif [ -n "$extracted_dir" ] && [ -f "${extracted_dir}/octai" ]; then
    mv "${extracted_dir}/octai" "${TMPDIR}/${BINARY_NAME}"
  elif [ -f "${TMPDIR}/${BINARY_NAME}" ]; then
    :
  else
    warn "Could not find binary in archive. Building from source..."
    install_from_source
  fi
}

main() {
  local tag="${1:-latest}"
  local no_path_check=false

  for arg in "$@"; do
    case "$arg" in
      --no-path-check) no_path_check=true ;;
      latest)          tag="latest" ;;
      *)               tag="$arg" ;;
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

  info "Detected: ${os}/${arch}"

  if [ "$tag" = "latest" ]; then
    tag="$(get_latest_release)"
    info "Latest release: ${tag}"
  fi

  if [ "${FORCE_SOURCE:-}" = "1" ] || [ "$os" = "windows" ]; then
    install_from_source
  else
    install_prebuilt "$os" "$arch" "$tag"
  fi

  [ -f "${TMPDIR}/${BINARY_NAME}" ] || die "Build failed — binary not found"

  chmod +x "${TMPDIR}/${BINARY_NAME}"

  mkdir -p "$INSTALL_DIR"
  mv "${TMPDIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"

  ok "Installed to ${INSTALL_DIR}/${BINARY_NAME}"

  if [ "$no_path_check" = false ]; then
    case ":$PATH:" in
      *":${INSTALL_DIR}:"*) ;;
      *)
        warn "${INSTALL_DIR} is not in your PATH."
        printf "\n  Add it to your shell profile:\n\n"
        if [ -n "${ZSH_VERSION:-}" ] || [ -f "$HOME/.zshrc" ]; then
          printf "    ${CYAN}echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.zshrc${RESET}\n"
          printf "    ${CYAN}source ~/.zshrc${RESET}\n\n"
        else
          printf "    ${CYAN}echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc${RESET}\n"
          printf "    ${CYAN}source ~/.bashrc${RESET}\n\n"
        fi
        ;;
    esac
  fi

  printf "\n"
  ok "OctAi ${tag} installed successfully!"
  printf "\n"
  printf "  ${BOLD}Getting started:${RESET}\n"
  printf "    ${CYAN}octai onboard${RESET}          Interactive setup wizard\n"
  printf "    ${CYAN}octai web${RESET}              Start web dashboard (http://localhost:18800)\n"
  printf "    ${CYAN}octai tui${RESET}              Start terminal UI\n"
  printf "    ${CYAN}octai agent${RESET}            Start AI chat session\n"
  printf "    ${CYAN}octai auth login -p openai${RESET}   Login with OpenAI\n"
  printf "    ${CYAN}octai auth login -p anthropic --browser-oauth${RESET}  Login with Anthropic\n"
  printf "\n"
}

main "$@"
