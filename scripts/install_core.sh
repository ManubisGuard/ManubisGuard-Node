#!/usr/bin/env bash
set -euo pipefail

# Resolve the Xray release from this project's Go dependency so the binary
# and github.com/xtls/xray-core protobuf/API version cannot silently drift.
SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd -- "$SCRIPT_DIR/.." && pwd)"
RELEASE_TAG=""
TAG_WAS_SET=0
TARGET_OS="linux"
TARGET_ARCH=""

usage() {
  echo "Usage: install_core.sh [--tag <release-tag|latest>] [--os linux] [--arch <arch>]"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag) RELEASE_TAG="$2"; TAG_WAS_SET=1; shift 2 ;;
    --os) TARGET_OS="$2"; shift 2 ;;
    --arch) TARGET_ARCH="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1"; usage; exit 1 ;;
  esac
done

[[ "$EUID" -eq 0 ]] || { echo "error: run as root"; exit 1; }
[[ "$TARGET_OS" == "linux" ]] || { echo "error: only linux is supported"; exit 1; }

if [[ "$TAG_WAS_SET" -eq 0 ]]; then
  [[ -f "$PROJECT_ROOT/go.mod" ]] || {
    echo "error: go.mod not found; pass --tag explicitly"; exit 1;
  }

  MODULE_VERSION="$(awk '$1 == "github.com/xtls/xray-core" { print $2; exit }' "$PROJECT_ROOT/go.mod")"
  # Xray Go module tags use v1.YYMMDD.PATCH while release assets use vYY.M.DD.
  if [[ "$MODULE_VERSION" =~ ^v1\.([0-9]{2})([0-9]{2})([0-9]{2})\.([0-9]+)$ ]]; then
    RELEASE_TAG="v${BASH_REMATCH[1]}.$((10#${BASH_REMATCH[2]})).$((10#${BASH_REMATCH[3]}))"
  else
    echo "error: cannot map Xray Go module version '$MODULE_VERSION' to a release tag."
    echo "Update this resolver for the new version format or pass --tag explicitly."
    exit 1
  fi
fi

if [[ -z "$TARGET_ARCH" ]]; then
  case "$(uname -m)" in
    x86_64|amd64) TARGET_ARCH="64" ;;
    aarch64|arm64) TARGET_ARCH="arm64-v8a" ;;
    armv7|armv7l) TARGET_ARCH="arm32-v7a" ;;
    armv6l) TARGET_ARCH="arm32-v6" ;;
    armv5tel) TARGET_ARCH="arm32-v5" ;;
    i386|i686) TARGET_ARCH="32" ;;
    mips) TARGET_ARCH="mips32" ;;
    mipsle) TARGET_ARCH="mips32le" ;;
    mips64) TARGET_ARCH="mips64" ;;
    mips64le) TARGET_ARCH="mips64le" ;;
    ppc64) TARGET_ARCH="ppc64" ;;
    ppc64le) TARGET_ARCH="ppc64le" ;;
    riscv64) TARGET_ARCH="riscv64" ;;
    s390x) TARGET_ARCH="s390x" ;;
    *) echo "error: unsupported architecture: $(uname -m)"; exit 1 ;;
  esac
fi

if ! command -v curl >/dev/null 2>&1 || ! command -v unzip >/dev/null 2>&1; then
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update -qq
    DEBIAN_FRONTEND=noninteractive apt-get install -y curl unzip
  elif command -v dnf >/dev/null 2>&1; then
    dnf install -y curl unzip
  elif command -v yum >/dev/null 2>&1; then
    yum install -y curl unzip
  else
    echo "error: install curl and unzip first"; exit 1
  fi
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
zip="$tmp/xray.zip"

if [[ "$RELEASE_TAG" == "latest" ]]; then
  url="https://github.com/XTLS/Xray-core/releases/latest/download/Xray-linux-$TARGET_ARCH.zip"
else
  url="https://github.com/XTLS/Xray-core/releases/download/$RELEASE_TAG/Xray-linux-$TARGET_ARCH.zip"
fi

echo "Installing Xray release $RELEASE_TAG for $TARGET_OS/$TARGET_ARCH (Go module: ${MODULE_VERSION:-explicit override})"
curl -fL --retry 3 -o "$zip" "$url"
unzip -q "$zip" -d "$tmp/extracted"
install -m 0755 "$tmp/extracted/xray" /usr/local/bin/xray
install -d -m 0755 /usr/local/share/xray
install -m 0644 "$tmp/extracted/geoip.dat" /usr/local/share/xray/geoip.dat
install -m 0644 "$tmp/extracted/geosite.dat" /usr/local/share/xray/geosite.dat

echo "Xray installed from release $RELEASE_TAG."
