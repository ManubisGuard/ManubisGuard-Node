#!/usr/bin/env bash
set -euo pipefail

REPO="${MANUBISGUARD_NODE_REPO:-https://github.com/ManubisGuard/ManubisGuard-Node.git}"
BRANCH="${MANUBISGUARD_NODE_BRANCH:-main}"
INSTALL_DIR="${MANUBISGUARD_NODE_DIR:-/opt/manubisguard-node}"
DATA_DIR="/var/lib/pg-node"

[[ "$EUID" -eq 0 ]] || { echo "ERROR: run as root"; exit 1; }

if command -v apt-get >/dev/null 2>&1; then
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install -y git curl ca-certificates openssl wireguard-tools python3 docker.io docker-compose-plugin
else
  echo "ERROR: this installer currently requires Debian/Ubuntu (apt-get)."
  exit 1
fi



modprobe wireguard 2>/dev/null || true
mkdir -p "$DATA_DIR/certs" "$DATA_DIR/generated"

if [[ ! -s "$DATA_DIR/certs/ssl_cert.pem" || ! -s "$DATA_DIR/certs/ssl_key.pem" ]]; then
  openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 -nodes     -keyout "$DATA_DIR/certs/ssl_key.pem"     -out "$DATA_DIR/certs/ssl_cert.pem"     -days 3650 -subj "/CN=pasarguard-node-bootstrap"     -addext "subjectAltName=DNS:pasarguard-node-bootstrap,IP:127.0.0.1"
  chmod 600 "$DATA_DIR/certs/ssl_key.pem"
  chmod 644 "$DATA_DIR/certs/ssl_cert.pem"
fi

if [[ -z "${NODE_API_KEY:-}" ]]; then
  NODE_API_KEY="$(cat /proc/sys/kernel/random/uuid)"
fi

rm -rf "$INSTALL_DIR"
git clone --depth 1 --branch "$BRANCH" "$REPO" "$INSTALL_DIR"
cd "$INSTALL_DIR"

if [[ ! -f .env ]]; then
  cp .env.example .env
fi

python3 - "$NODE_API_KEY" <<'PY'
from pathlib import Path
import sys
key=sys.argv[1]
p=Path(".env")
s=p.read_text()
lines=[]
seen=False
for line in s.splitlines():
    if line.startswith("API_KEY"):
        lines.append(f"API_KEY = {key}"); seen=True
    else:
        lines.append(line)
if not seen:
    lines.append(f"API_KEY = {key}")
p.write_text("\n".join(lines)+"\n")
PY

docker compose build --pull=false
docker compose up -d

echo "Node source: $REPO"
echo "Node branch: $BRANCH"
echo "Node API key: $NODE_API_KEY"
echo "Node service: 62050/grpc (bootstrap self-signed TLS)"
docker compose ps
