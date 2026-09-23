#!/usr/bin/env bash
set -Eeuo pipefail

REPO_URL="${REPO_URL:-https://github.com/arsamnikzaad/ManubisGuard-Node.git}"
REPO_BRANCH="${REPO_BRANCH:-feature/amnezia-wg}"
INSTALL_DIR="${INSTALL_DIR:-/opt/manubisguard-node}"
DATA_DIR="${DATA_DIR:-/var/lib/pg-node}"
COMPOSE_FILE="$INSTALL_DIR/docker-compose.yml"
CONTAINER_NAME="manubisguard-node"
IMAGE_NAME="manubisguard-node:feature-amnezia-wg"
SERVICE_PORT="${SERVICE_PORT:-62050}"

log(){ echo "[ManubisGuard Node] $*"; }
die(){ echo "[ManubisGuard Node] ERROR: $*" >&2; exit 1; }
need_root(){ [ "$(id -u)" -eq 0 ] || die "Run as root."; }

install_docker(){
  command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1 && return 0
  log "Installing Docker..."
  if command -v apt-get >/dev/null 2>&1; then
    apt-get update
    apt-get install -y ca-certificates curl git openssl
    install -m 0755 -d /etc/apt/keyrings
    if [ ! -f /etc/apt/keyrings/docker.asc ]; then
      curl -fsSL https://download.docker.com/linux/$(. /etc/os-release && echo "$ID")/gpg -o /etc/apt/keyrings/docker.asc
      chmod a+r /etc/apt/keyrings/docker.asc
    fi
    . /etc/os-release
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/$ID $VERSION_CODENAME stable" >/etc/apt/sources.list.d/docker.list
    apt-get update
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    systemctl enable --now docker
  else
    die "Unsupported OS: apt-get is required."
  fi
}

prepare_tools(){
  for x in curl git openssl; do command -v "$x" >/dev/null 2>&1 || die "Missing required command: $x"; done
  docker info >/dev/null 2>&1 || die "Docker is not running."
}

prepare_source(){
  if [ -d "$INSTALL_DIR/.git" ]; then
    git -C "$INSTALL_DIR" fetch --depth 1 origin "$REPO_BRANCH"
    git -C "$INSTALL_DIR" checkout -q "$REPO_BRANCH"
    git -C "$INSTALL_DIR" reset --hard -q "origin/$REPO_BRANCH"
  else
    rm -rf "$INSTALL_DIR"
    git clone --depth 1 --branch "$REPO_BRANCH" "$REPO_URL" "$INSTALL_DIR"
  fi
}

prepare_data(){
  mkdir -p "$DATA_DIR/certs" "$DATA_DIR/generated"
  chmod 700 "$DATA_DIR" "$DATA_DIR/certs" "$DATA_DIR/generated"

  if [ ! -s "$DATA_DIR/api_key" ]; then
    cat /proc/sys/kernel/random/uuid > "$DATA_DIR/api_key"
    chmod 600 "$DATA_DIR/api_key"
  fi
  API_KEY="$(cat "$DATA_DIR/api_key")"

  if [ ! -s "$DATA_DIR/certs/ssl_key.pem" ] || [ ! -s "$DATA_DIR/certs/ssl_cert.pem" ]; then
    NODE_IP="$(curl -4fsS https://api.ipify.org)" || die "Could not detect public IPv4."
    log "Generating TLS certificate for $NODE_IP..."
    openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 \
      -keyout "$DATA_DIR/certs/ssl_key.pem" \
      -out "$DATA_DIR/certs/ssl_cert.pem" \
      -days 3650 -nodes -subj "/CN=$NODE_IP" \
      -addext "subjectAltName = IP:$NODE_IP"
    chmod 600 "$DATA_DIR/certs/ssl_key.pem"
    chmod 644 "$DATA_DIR/certs/ssl_cert.pem"
  fi
}

write_compose(){
  cat > "$COMPOSE_FILE" <<EOF
services:
  node:
    build:
      context: .
      dockerfile: Dockerfile
    image: $IMAGE_NAME
    container_name: $CONTAINER_NAME
    restart: always
    network_mode: host
    cap_add:
      - NET_ADMIN
    environment:
      SERVICE_PORT: $SERVICE_PORT
      SERVICE_PROTOCOL: grpc
      API_KEY: $API_KEY
      NODE_HOST: 0.0.0.0
      PG_NODE_WG_HOST_ROUTING: "1"
      SSL_CERT_FILE: /var/lib/pg-node/certs/ssl_cert.pem
      SSL_KEY_FILE: /var/lib/pg-node/certs/ssl_key.pem
      GENERATED_CONFIG_PATH: /var/lib/pg-node/generated
    volumes:
      - $DATA_DIR:/var/lib/pg-node
EOF
}

install_node(){
  docker compose -f "$COMPOSE_FILE" down --remove-orphans >/dev/null 2>&1 || true
  log "Building ManubisGuard Node from $REPO_BRANCH..."
  docker compose -f "$COMPOSE_FILE" build --pull node
  docker compose -f "$COMPOSE_FILE" up -d node
  sleep 5
  docker compose -f "$COMPOSE_FILE" ps
  if ! ss -lntp 2>/dev/null | grep -q ":$SERVICE_PORT "; then
    docker logs --tail 100 "$CONTAINER_NAME" || true
    die "Node is not listening on port $SERVICE_PORT."
  fi
  NODE_IP="$(curl -4fsS https://api.ipify.org)" || NODE_IP="unknown"
  echo
  echo "=============================================="
  echo " ManubisGuard Node + AmneziaWG installation"
  echo "=============================================="
  echo "Address:     $NODE_IP"
  echo "Port:        $SERVICE_PORT"
  echo "Protocol:    grpc"
  echo "API Key:     $API_KEY"
  echo "Certificate: $DATA_DIR/certs/ssl_cert.pem"
  echo "Source:      $REPO_URL"
  echo "Branch:      $REPO_BRANCH"
  echo "Data:        $DATA_DIR"
  echo "Compose:     $COMPOSE_FILE"
  echo "Container:   $CONTAINER_NAME"
  echo "=============================================="
}

status_node(){
  docker compose -f "$COMPOSE_FILE" ps 2>/dev/null || true
  ss -lntp 2>/dev/null | grep ":$SERVICE_PORT " || true
}

uninstall_node(){
  if [ -f "$COMPOSE_FILE" ]; then docker compose -f "$COMPOSE_FILE" down --remove-orphans || true; fi
  docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
  log "Container stopped and removed. Data kept at $DATA_DIR and source kept at $INSTALL_DIR."
}

main(){
  need_root
  ACTION="${1:-install}"
  case "$ACTION" in
    install)
      install_docker
      prepare_tools
      prepare_source
      prepare_data
      write_compose
      install_node
      ;;
    status)
      status_node
      ;;
    uninstall)
      uninstall_node
      ;;
    *)
      echo "Usage: $0 {install|status|uninstall}"
      exit 2
      ;;
  esac
}
main "$@"
