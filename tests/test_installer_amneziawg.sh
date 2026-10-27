#!/usr/bin/env bash
set -Eeuo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALLER="$ROOT/install.sh"
COMPOSE="$ROOT/docker-compose.yml"
README="$ROOT/README.md"

assert_contains() {
  local file="$1" pattern="$2"
  if ! grep -Fq -- "$pattern" "$file"; then
    printf 'missing expected text in %s: %s\n' "$file" "$pattern" >&2
    exit 1
  fi
}

bash -n "$INSTALLER"
assert_contains "$INSTALLER" 'install_amneziawg_module'
assert_contains "$INSTALLER" 'linux-headers-$kernel'
assert_contains "$INSTALLER" 'build-essential dkms git'
assert_contains "$INSTALLER" 'dkms install "amneziawg/$expected_version"'
assert_contains "$INSTALLER" 'modprobe amneziawg'
assert_contains "$INSTALLER" '/etc/modules-load.d/amneziawg.conf'
assert_contains "$INSTALLER" 'type amneziawg'
assert_contains "$INSTALLER" 'cat "$CERT_FILE"'
assert_contains "$INSTALLER" 'Private key is intentionally not printed.'
assert_contains "$COMPOSE" 'ghcr.io/manubisguard/manubisguard-node:v0.6.0'
assert_contains "$README" 'v1.0.20260929+awg'
if grep -Fq 'cat "$KEY_FILE"' "$INSTALLER"; then
  echo 'installer must never print the private TLS key' >&2
  exit 1
fi
printf 'AmneziaWG installer prerequisites/output checks passed.\n'
