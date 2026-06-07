#!/usr/bin/env sh
set -eu

REPO="devlopersabbir/tview-cli"
BINARY="tview"
VERSION="${TVIEW_VERSION:-latest}"
INSTALL_DIR="${TVIEW_INSTALL_DIR:-}"

need() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "tview installer: missing required command: $1" >&2
    exit 1
  }
}

detect_os() {
  case "$(uname -s)" in
    Darwin) echo "darwin" ;;
    Linux) echo "linux" ;;
    *)
      echo "tview installer: unsupported OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64 | amd64) echo "amd64" ;;
    arm64 | aarch64) echo "arm64" ;;
    *)
      echo "tview installer: unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

choose_install_dir() {
  if [ -n "$INSTALL_DIR" ]; then
    echo "$INSTALL_DIR"
  elif [ -d /usr/local/bin ] && [ -w /usr/local/bin ]; then
    echo "/usr/local/bin"
  else
    echo "$HOME/.local/bin"
  fi
}

need curl
need tar
need mktemp

os="$(detect_os)"
arch="$(detect_arch)"
dir="$(choose_install_dir)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT INT TERM

if [ "$VERSION" = "latest" ]; then
  url="https://github.com/${REPO}/releases/latest/download/${BINARY}_${os}_${arch}.tar.gz"
else
  case "$VERSION" in
    v*) tag="$VERSION" ;;
    *) tag="v$VERSION" ;;
  esac
  url="https://github.com/${REPO}/releases/download/${tag}/${BINARY}_${os}_${arch}.tar.gz"
fi

echo "Downloading $BINARY for $os/$arch..."
curl -fsSL "$url" -o "$tmp/${BINARY}.tar.gz"
tar -xzf "$tmp/${BINARY}.tar.gz" -C "$tmp"

mkdir -p "$dir"
install -m 0755 "$tmp/$BINARY" "$dir/$BINARY"

if [ "$os" = "darwin" ] && command -v xattr >/dev/null 2>&1; then
  xattr -d com.apple.quarantine "$dir/$BINARY" 2>/dev/null || true
  xattr -d com.apple.provenance "$dir/$BINARY" 2>/dev/null || true
fi

echo "Installed $BINARY to $dir/$BINARY"

case ":$PATH:" in
  *":$dir:"*) ;;
  *)
    echo
    echo "$dir is not in your PATH."
    echo "Add this to your shell profile, then restart your terminal:"
    echo "  export PATH=\"$dir:\$PATH\""
    ;;
esac

echo
"$dir/$BINARY" version
