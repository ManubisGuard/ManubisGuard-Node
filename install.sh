#!/usr/bin/env bash
set -Eeuo pipefail

REPO="${MANUBISGUARD_NODE_REPO:-https://github.com/ManubisGuard/ManubisGuard-Node.git}"
BRANCH="${MANUBISGUARD_NODE_BRANCH:-main}"
APP_NAME="${MANUBISGUARD_NODE_NAME:-pg-node}"
APP_DIR="${MANUBISGUARD_NODE_DIR:-/opt/manubisguard-node}"
DATA_DIR="${MANUBISGUARD_NODE_DATA_DIR:-/var/lib/pg-node}"
ENV_FILE="$APP_DIR/.env"
COMPOSE_FILE="$APP_DIR/docker-compose.yml"
CERT_DIR="$DATA_DIR/certs"
CERT_FILE="$CERT_DIR/ssl_cert.pem"
KEY_FILE="$CERT_DIR/ssl_key.pem"
SERVICE_NAME="${APP_NAME}-service"
SERVICE_UNIT="/etc/systemd/system/$SERVICE_NAME.service"

AUTO_CONFIRM=false
OVERRIDE=false
SERVICE_PORT=""
API_PORT=""
API_KEY="${NODE_API_KEY:-}"
USE_REST=""
CERT_PATH=""
KEY_PATH=""
SELF_SIGNED=false
SAN_ENTRIES=""
NODE_IP=""

log(){ printf '[manubisguard-node] %s\n' "$*"; }
die(){ printf '[manubisguard-node] ERROR: %s\n' "$*" >&2; exit 1; }
tty_prompt(){
  local __v="$1" prompt="$2" default="${3:-}" answer
  printf '%s' "$prompt" >/dev/tty
  IFS= read -r answer </dev/tty || die "could not read input from terminal"
  [[ -z "$answer" ]] && answer="$default"
  printf -v "$__v" '%s' "$answer"
}
run_root(){ [[ "$EUID" -eq 0 ]] || die "run as root"; }

usage(){
cat <<EOF
ManubisGuard Node installer

Usage:
  install.sh install [options]

Interactive install follows the PasarGuard Node flow:
  1. system validation
  2. Docker/package preparation
  3. node name
  4. TLS certificate / SAN
  5. API key
  6. REST or gRPC
  7. SERVICE_PORT
  8. API_PORT
  9. Docker deployment
 10. systemd service
 11. runtime validation

Options:
  --name NAME
  --service-port PORT
  --api-port PORT
  --api-key UUID
  --use-rest | --use-grpc
  --self-signed
  --cert-path FILE --key-path FILE
  --san-entries LIST
  --yes|-y
  --override

Defaults:
  repository: $REPO
  branch:     $BRANCH
  install:    $APP_DIR
  data:       $DATA_DIR
  service:    62050
  api:        62051
EOF
}

parse(){
  local a
  while [[ $# -gt 0 ]]; do
    a="$1"
    case "$a" in
      --name) [[ $# -ge 2 ]] || die "--name requires a value"; APP_NAME="$2"; shift 2;;
      --name=*) APP_NAME="${a#*=}"; shift;;
      --service-port) [[ $# -ge 2 ]] || die "--service-port requires a value"; SERVICE_PORT="$2"; shift 2;;
      --service-port=*) SERVICE_PORT="${a#*=}"; shift;;
      --api-port) [[ $# -ge 2 ]] || die "--api-port requires a value"; API_PORT="$2"; shift 2;;
      --api-port=*) API_PORT="${a#*=}"; shift;;
      --api-key) [[ $# -ge 2 ]] || die "--api-key requires a value"; API_KEY="$2"; shift 2;;
      --api-key=*) API_KEY="${a#*=}"; shift;;
      --use-rest) USE_REST=true; shift;;
      --use-grpc) USE_REST=false; shift;;
      --self-signed) SELF_SIGNED=true; shift;;
      --cert-path) [[ $# -ge 2 ]] || die "--cert-path requires a value"; CERT_PATH="$2"; shift 2;;
      --key-path) [[ $# -ge 2 ]] || die "--key-path requires a value"; KEY_PATH="$2"; shift 2;;
      --san-entries) [[ $# -ge 2 ]] || die "--san-entries requires a value"; SAN_ENTRIES="$2"; shift 2;;
      --yes|-y) AUTO_CONFIRM=true; shift;;
      --override) OVERRIDE=true; shift;;
      -h|--help) usage; exit 0;;
      install) shift;;
      *) die "unknown option: $a";;
    esac
  done
}

validate_port(){
  local p="$1" label="$2"
  [[ "$p" =~ ^[0-9]+$ && "$p" -ge 1 && "$p" -le 65535 ]] || die "$label must be between 1 and 65535"
}

occupied(){
  ss -tuln 2>/dev/null | awk '{print $5}' | grep -Eo '[0-9]+$' | sort -u | grep -qx "$1"
}

choose_port(){
  local __v="$1" label="$2" default="$3" value
  if [[ -n "${!__v:-}" ]]; then
    value="${!__v}"
    validate_port "$value" "$label"
    occupied "$value" && die "$label $value is already in use"
    printf -v "$__v" '%s' "$value"
    return
  fi
  if "$AUTO_CONFIRM"; then
    value="$default"
    occupied "$value" && die "$label $value is already in use; run interactively to choose another"
    printf -v "$__v" '%s' "$value"
    return
  fi
  while true; do
    tty_prompt value "Enter the $label (default $default): " "$default"
    if ! [[ "$value" =~ ^[0-9]+$ && "$value" -ge 1 && "$value" -le 65535 ]]; then
      log "Invalid port. Enter 1-65535."
    elif occupied "$value"; then
      log "Port $value is already in use. Choose another."
    else
      printf -v "$__v" '%s' "$value"
      return
    fi
  done
}

install_docker(){
  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    log "Docker and Compose plugin already installed."
    return
  fi
  log "Docker/Compose not ready. Installing Docker Engine from the official Docker repository."
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install -y ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  if [[ ! -s /etc/apt/keyrings/docker.asc ]]; then
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    chmod a+r /etc/apt/keyrings/docker.asc
  fi
  . /etc/os-release
  cat >/etc/apt/sources.list.d/docker.list <<EOF
deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu ${VERSION_CODENAME} stable
EOF
  apt-get update
  apt-cache policy docker-ce | grep -q 'Candidate:' || die "Docker CE repository is unavailable for ${VERSION_CODENAME}"
  if ! command -v docker >/dev/null 2>&1; then
    if dpkg-query -W -f='${db:Status-Abbrev}' containerd 2>/dev/null | grep -q '^ii'; then
      log "Removing conflicting distro containerd package before Docker CE installation."
      DEBIAN_FRONTEND=noninteractive apt-get remove -y containerd
    fi
    DEBIAN_FRONTEND=noninteractive apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  else
    DEBIAN_FRONTEND=noninteractive apt-get install -y docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  fi
  systemctl enable --now docker
}

install_packages(){
  . /etc/os-release
  [[ "${ID:-}" == "ubuntu" || "${ID:-}" == "debian" ]] || die "Ubuntu/Debian is required"
  log "System: ${PRETTY_NAME:-unknown}"
  apt-get update
  DEBIAN_FRONTEND=noninteractive apt-get install -y git curl ca-certificates openssl wireguard-tools python3 jq uuid-runtime iproute2 lsof
  modprobe wireguard 2>/dev/null || true
  install_docker
}

get_ip(){
  NODE_IP="$(curl -4fsS --max-time 8 https://ifconfig.io 2>/dev/null || true)"
  [[ -n "$NODE_IP" ]] || NODE_IP="$(curl -4fsS --max-time 8 https://api.ipify.org 2>/dev/null || true)"
  [[ -n "$NODE_IP" ]] || NODE_IP="127.0.0.1"
}

ensure_dirs(){
  mkdir -p "$CERT_DIR" "$DATA_DIR/generated" "$APP_DIR"
  chmod 700 "$CERT_DIR"
}

generate_cert(){
  local sans="DNS:localhost,IP:127.0.0.1,IP:$NODE_IP"
  if [[ -n "$SAN_ENTRIES" ]]; then
    local x
    IFS=',' read -ra xs <<<"$SAN_ENTRIES"
    for x in "${xs[@]}"; do
      x="$(printf '%s' "$x" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
      [[ -z "$x" ]] && continue
      if [[ "$x" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]]; then
        sans="$sans,IP:$x"
      else
        sans="$sans,DNS:$x"
      fi
    done
  fi
  log "Generating self-signed TLS certificate."
  openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 -nodes     -keyout "$KEY_FILE" -out "$CERT_FILE" -days 3650     -subj "/CN=$NODE_IP" -addext "subjectAltName = $sans" >/dev/null 2>&1
  chmod 600 "$KEY_FILE"
  chmod 644 "$CERT_FILE"
}

configure_tls(){
  if [[ "$SELF_SIGNED" == true || ( -z "$CERT_PATH" && -z "$KEY_PATH" && "$AUTO_CONFIRM" == true ) ]]; then
    generate_cert
    return
  fi
  local answer
  if [[ -n "$CERT_PATH" || -n "$KEY_PATH" ]]; then
    answer="y"
  else
    tty_prompt answer "Do you want to use your own public certificate instead? (Y/n): " "n"
  fi
  if [[ "$answer" =~ ^[Yy]$ ]]; then
    [[ -n "$CERT_PATH" ]] || tty_prompt CERT_PATH "Certificate file path: "
    [[ -n "$KEY_PATH" ]] || tty_prompt KEY_PATH "Private key file path: "
    [[ -f "$CERT_PATH" && -f "$KEY_PATH" ]] || die "certificate/key file not found"
    openssl x509 -in "$CERT_PATH" -noout >/dev/null 2>&1 || die "invalid certificate file"
    openssl pkey -in "$KEY_PATH" -noout >/dev/null 2>&1 || die "invalid private key file"
    cp "$CERT_PATH" "$CERT_FILE"
    cp "$KEY_PATH" "$KEY_FILE"
    chmod 600 "$KEY_FILE"
    chmod 644 "$CERT_FILE"
  else
    generate_cert
  fi
}

configure_inputs(){
  local answer
  if [[ "$APP_NAME" == "pg-node" && "$AUTO_CONFIRM" != true ]]; then
    tty_prompt APP_NAME "Node name (default pg-node): " "pg-node"
  fi
  [[ "$APP_NAME" =~ ^[A-Za-z0-9][A-Za-z0-9_-]{0,62}$ ]] || die "invalid node name"
  if [[ -z "$API_KEY" ]]; then
    if "$AUTO_CONFIRM"; then
      API_KEY="$(cat /proc/sys/kernel/random/uuid)"
    else
      tty_prompt API_KEY "Enter API Key (UUID, ENTER to generate): " ""
      [[ -n "$API_KEY" ]] || API_KEY="$(cat /proc/sys/kernel/random/uuid)"
    fi
  fi
  [[ "$API_KEY" =~ ^[0-9a-fA-F]{8}-[0-9a-fA-F-]{27,}$ ]] || die "API key must be a UUID"
  if [[ -z "$USE_REST" ]]; then
    if "$AUTO_CONFIRM"; then
      USE_REST=false
    else
      answer="n"
      tty_prompt answer "Do you want to use REST protocol instead of gRPC? (y/N): " "n"
      answer="${answer:-n}"
      if [[ "$answer" =~ ^[Yy]$ ]]; then
        USE_REST=true
      else
        USE_REST=false
      fi
    fi
  fi
  choose_port SERVICE_PORT "SERVICE_PORT" 62050
  choose_port API_PORT "API_PORT" 62051
  [[ "$SERVICE_PORT" != "$API_PORT" ]] || die "SERVICE_PORT and API_PORT cannot be the same"
}

prepare_source(){
  if [[ -d "$APP_DIR/.git" ]]; then
    if "$OVERRIDE"; then
      git -C "$APP_DIR" fetch origin "$BRANCH"
      git -C "$APP_DIR" reset --hard "origin/$BRANCH"
    else
      git -C "$APP_DIR" pull --ff-only origin "$BRANCH" || true
    fi
  else
    rm -rf "$APP_DIR"
    git clone --depth 1 --branch "$BRANCH" "$REPO" "$APP_DIR"
  fi
}

write_config(){
  cp "$APP_DIR/.env.example" "$ENV_FILE"
  python3 - "$ENV_FILE" "$SERVICE_PORT" "$API_KEY" "$USE_REST" "$CERT_DIR" "$DATA_DIR" <<'PY'
from pathlib import Path
import re, sys
p=Path(sys.argv[1]); service=sys.argv[2]; key=sys.argv[3]; rest=sys.argv[4]=="true"; cert=sys.argv[5]; data=sys.argv[6]
s=p.read_text()
def put(s,k,v):
    pat=r'(?m)^#?\s*'+re.escape(k)+r'\s*=.*$'
    repl=f'{k} = {v}'
    return re.sub(pat,repl,s,count=1) if re.search(pat,s) else s+'\n'+repl+'\n'
s=put(s,"SERVICE_PORT",service)
s=put(s,"API_KEY",key)
s=put(s,"SSL_CERT_FILE",cert)
s=put(s,"SSL_KEY_FILE",cert.replace("ssl_cert.pem","ssl_key.pem"))
s=put(s,"SERVICE_PROTOCOL",'"rest"' if rest else '"grpc"')
s=put(s,"GENERATED_CONFIG_PATH",data+"/generated")
p.write_text(s)
PY
  python3 - "$COMPOSE_FILE" "$SERVICE_PORT" "$USE_REST" "$DATA_DIR" "$APP_NAME" <<'PY'
from pathlib import Path
import re, sys
p=Path(sys.argv[1]); port=sys.argv[2]; rest=sys.argv[3]=="true"; data=sys.argv[4]; name=sys.argv[5]
s=p.read_text()
s=re.sub(r'(?m)^\s*SERVICE_PORT:\s*.*$', f'      SERVICE_PORT: {port}', s, count=1)
s=re.sub(r'(?m)^\s*SERVICE_PROTOCOL:\s*.*$', f'      SERVICE_PROTOCOL: "{"rest" if rest else "grpc"}"', s, count=1)
s=re.sub(r'(?m)^\s*- /var/lib/pg-node:/var/lib/pg-node$', f'      - {data}:/var/lib/pg-node', s, count=1)
if name != "pg-node" and "container_name:" not in s:
    s=s.replace("  node:\n", "  node:\n    container_name: "+name+"\n", 1)
p.write_text(s)
PY
  chmod 600 "$ENV_FILE"
}

install_service(){
  cat >/usr/local/bin/"$APP_NAME" <<EOF
#!/usr/bin/env bash
set -e
cd "$APP_DIR"
exec docker compose -f "$COMPOSE_FILE" "$@"
EOF
  chmod 755 /usr/local/bin/"$APP_NAME"
  cat >"$SERVICE_UNIT" <<EOF
[Unit]
Description=ManubisGuard Node ($APP_NAME)
After=network-online.target docker.service
Wants=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$APP_DIR
ExecStart=/usr/bin/docker compose -f $COMPOSE_FILE up -d
ExecStop=/usr/bin/docker compose -f $COMPOSE_FILE down
ExecReload=/usr/bin/docker compose -f $COMPOSE_FILE up -d --build
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF
  systemctl daemon-reload
  systemctl enable --now "$SERVICE_NAME"
}

validate(){
  log "Waiting for node container."
  local i
  for i in $(seq 1 30); do
    if docker compose -f "$COMPOSE_FILE" ps --status running --services 2>/dev/null | grep -qx node; then
      break
    fi
    sleep 2
  done
  docker compose -f "$COMPOSE_FILE" ps
  occupied "$SERVICE_PORT" || die "SERVICE_PORT $SERVICE_PORT is not listening"
  systemctl is-enabled "$SERVICE_NAME" >/dev/null 2>&1 || die "systemd service is not enabled"
}

install_node(){
  run_root
  log "=== 1/11 System validation ==="
  install_packages
  get_ip
  log "Detected public IP: $NODE_IP"
  log "=== 2/11 Installation options ==="
  configure_inputs
  log "=== 3/11 Directories ==="
  ensure_dirs
  log "=== 4/11 TLS certificate ==="
  configure_tls
  log "=== 5/11 Source ==="
  prepare_source
  log "=== 6/11 Environment ==="
  write_config
  log "=== 7/11 Docker build ==="
  docker compose -f "$COMPOSE_FILE" build --pull=false
  log "=== 8/11 Docker start ==="
  docker compose -f "$COMPOSE_FILE" up -d --remove-orphans
  log "=== 9/11 Runtime ==="
  docker compose -f "$COMPOSE_FILE" ps
  log "=== 10/11 systemd service ==="
  install_service
  log "=== 11/11 Validation ==="
  validate
  printf '\n==============================================\n'
  printf ' ManubisGuard Node installation completed\n'
  printf '==============================================\n'
  printf 'Node name   : %s\n' "$APP_NAME"
  printf 'Address     : %s\n' "$NODE_IP"
  printf 'Protocol    : %s\n' "$([[ "$USE_REST" == true ]] && echo REST || echo gRPC)"
  printf 'Service port: %s\n' "$SERVICE_PORT"
  printf 'API port    : %s\n' "$API_PORT"
  printf 'Certificate : %s\n' "$CERT_FILE"
  printf 'API Key     : %s\n' "$API_KEY"
  printf 'Env         : %s\n' "$ENV_FILE"
  printf '\nUse this certificate and API key in ManubisGuard Panel.\n'
}

main(){
  run_root
  parse "$@"
  install_node
}
main "$@"
