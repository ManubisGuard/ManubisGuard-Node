# ManubisGuard Node
<p align="center">
    <a href="https://github.com/ManubisGuard/ManubisGuard-Node/actions">
        <img src="https://img.shields.io/github/actions/workflow/status/ManubisGuard/ManubisGuard-Node/docker-build.yml?style=flat-square" />
    </a>
    <a href="https://github.com/ManubisGuard/ManubisGuard-Node">
        <img src="https://img.shields.io/github/license/ManubisGuard/ManubisGuard-Node?style=flat-square" />
    </a>
</p>

# Documentation
ManubisGuard Node is the worker node used by ManubisGuard Panel.

## One-Click Installation

The installer follows the current PasarGuard Node installation flow, adapted for the ManubisGuard fork. It validates the host, prepares Docker safely, asks for node/TLS/protocol/port settings, builds the node, installs the systemd service, and validates the selected service port. The interactive prompts are read from the controlling terminal so the command also works with `curl | sudo bash`.

```bash
curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/main/install.sh | sudo bash -s -- install
```

### Interactive choices

During a normal install you can choose:

- Node name
- Self-signed TLS certificate or your own public certificate/key
- Additional certificate SAN entries
- API key or automatic UUID generation
- REST or gRPC protocol
- `SERVICE_PORT`, default `62050`
- `API_PORT`, default `62051`
- systemd service startup

The installer checks that selected ports are free before deployment.

### Unattended installation

```bash
curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/main/install.sh | sudo bash -s -- install \
  --name node-de1 \
  --service-port 62050 \
  --api-port 62051 \
  --use-grpc \
  --self-signed \
  --yes
```

For a public certificate:

```bash
sudo bash install.sh install \
  --cert-path /path/to/cert.pem \
  --key-path /path/to/key.pem \
  --service-port 62050 \
  --api-port 62051
```

### Runtime files

- Source: `/opt/manubisguard-node`
- Environment: `/opt/manubisguard-node/.env`
- Data: `/var/lib/pg-node`
- Certificate: `/var/lib/pg-node/certs/ssl_cert.pem`
- Private key: `/var/lib/pg-node/certs/ssl_key.pem`
- Compose: `/opt/manubisguard-node/docker-compose.yml`

The installer generates a self-signed certificate by default in interactive mode unless you select a public certificate. The certificate includes localhost, loopback, the detected public IP, and any additional SAN entries you provide.

## CLI

After installation, the node name is installed as a global command. With the default name:

```bash
sudo pg-node status
sudo pg-node restart
sudo pg-node logs
sudo pg-node
```

The systemd unit is:

```bash
sudo systemctl status pg-node-service
sudo systemctl restart pg-node-service
```

## Docker installation safety

The installer does not blindly install Ubuntu's `docker.io` package alongside Docker's `containerd.io`. If Docker/Compose is missing, it configures the official Docker repository and installs Docker CE plus the Compose plugin. This avoids the `containerd.io : Conflicts: containerd` failure caused by mixing the two package families.

## Source and customization

Environment overrides:

```bash
MANUBISGUARD_NODE_REPO=https://github.com/ManubisGuard/ManubisGuard-Node.git
MANUBISGUARD_NODE_BRANCH=main
MANUBISGUARD_NODE_NAME=pg-node
MANUBISGUARD_NODE_DIR=/opt/manubisguard-node
MANUBISGUARD_NODE_DATA_DIR=/var/lib/pg-node
```

## Upstream compatibility

The installation flow is intentionally aligned with the PasarGuard Node installer: system validation, Docker preparation, TLS/SAN handling, API key, REST/gRPC choice, occupied-port checks, node deployment, service installation, and final certificate/API-key output. The source and runtime identity point to ManubisGuard rather than the upstream PasarGuard repository. PasarGuard's current installer documents the same service/API port selection and certificate/API-key handoff flow. citeturn6view0turn3view1
