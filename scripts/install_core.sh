#!/usr/bin/env bash
set -euo pipefail

RELEASE_TAG="latest"
TARGET_OS="linux"
TARGET_ARCH=""

usage() {
  echo "Usage: install_core.sh [--tag <release-tag>] [--os linux] [--arch <arch>]"
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag) RELEASE_TAG="$2"; shift 2 ;;
    --os) TARGET_OS="$2"; shift 2 ;;
    --arch) TARGET_ARCH="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1"; usage; exit 1 ;;
  esac
done

[[ "$EUID" -eq 0 ]] || { echo "error: run as root"; exit 1; }
[[ "$TARGET_OS" == "linux" ]] || { echo "error: only linux is supported"; exit 1; }

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

curl -fL --retry 3 -o "$zip" "$url"
unzip -q "$zip" -d "$tmp/extracted"
install -m 0755 "$tmp/extracted/xray" /usr/local/bin/xray
install -d -m 0755 /usr/local/share/xray
install -m 0644 "$tmp/extracted/geoip.dat" /usr/local/share/xray/geoip.dat
install -m 0644 "$tmp/extracted/geosite.dat" /usr/local/share/xray/geosite.dat

echo "Xray installed from the Node-AWG fork installer."
