# ManubisGuard Node

AmneziaWG-enabled ManubisGuard Node from the `feature/amnezia-wg` branch.

## One-command installation

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/arsamnikzaad/ManubisGuard-Node/feature/amnezia-wg/install-manubisguard-node.sh) install
```

The installer uses the prebuilt GHCR image first, so a normal installation does **not** compile Go, Xray, or AmneziaWG on the VPS. If the image is unavailable, it automatically falls back to building from this repository.

It also:
- installs Docker when required
- creates a persistent API key
- creates the Node TLS certificate for the public IPv4
- enables host WireGuard routing
- starts gRPC on port `62050`
- verifies that the Node is listening before reporting success

## Status

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/arsamnikzaad/ManubisGuard-Node/feature/amnezia-wg/install-manubisguard-node.sh) status
```

## Uninstall

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/arsamnikzaad/ManubisGuard-Node/feature/amnezia-wg/install-manubisguard-node.sh) uninstall
```

Persistent data: `/var/lib/pg-node`  
Source: `/opt/manubisguard-node`

## Matching Panel

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/arsamnikzaad/ManubisGuard-Panel/feature/amnezia-wg/install-manubisguard.sh) install --database timescaledb
```
