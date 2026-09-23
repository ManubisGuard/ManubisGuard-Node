# ManubisGuard Node

AmneziaWG-enabled ManubisGuard Node based on the `feature/amnezia-wg` branch.

## One-command installation

On a clean Ubuntu/Debian server:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/arsamnikzaad/ManubisGuard-Node/feature/amnezia-wg/install-manubisguard-node.sh) install
```

The installer:

- installs Docker when required
- clones the `feature/amnezia-wg` branch
- generates a persistent API key
- generates a TLS certificate for the public IPv4 address
- builds the Node image locally from this repository
- builds the bundled AmneziaWG tools
- enables host WireGuard routing
- starts the Node with gRPC on port `62050`
- verifies that the Node is listening before reporting success

## Status

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/arsamnikzaad/ManubisGuard-Node/feature/amnezia-wg/install-manubisguard-node.sh) status
```

## Uninstall

The uninstall action stops and removes the Node container but keeps the source and persistent data:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/arsamnikzaad/ManubisGuard-Node/feature/amnezia-wg/install-manubisguard-node.sh) uninstall
```

Persistent Node data is stored under `/var/lib/pg-node` and source under `/opt/manubisguard-node`.

## Panel

Install the matching Panel with TimescaleDB:

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/arsamnikzaad/ManubisGuard-Panel/feature/amnezia-wg/install-manubisguard.sh) install --database timescaledb
```
