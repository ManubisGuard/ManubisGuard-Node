# ManubisGuard Node

AmneziaWG-enabled ManubisGuard Node from the `feature/amnezia-wg` branch.

## One-command installation

```bash
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/feature/amnezia-wg/install-manubisguard-node.sh)" @ install
```

The installer follows the interactive PasarGuard Node installation workflow. A normal interactive install asks for the Node service port, TLS/certificate mode, API key, and gRPC/REST transport. Use `-y` only when you intentionally want the non-interactive defaults.

After installation, the same workflow installs the `manubis-node` host command for lifecycle management.

```bash
sudo manubis-node status
sudo manubis-node restart
sudo manubis-node logs
```

## Multi-Core on one Node

A single ManubisGuard Node can be assigned to multiple Core configurations from the Node settings in the ManubisGuard Panel. Select the desired Core configurations with the checkboxes next to the Node connection settings.

- Multiple WireGuard/AmneziaWG Core instances can run concurrently on one Node.
- Xray uses one process for all Xray inbounds, so only one Xray Core configuration can be selected per Node.
- Existing Nodes that only have the legacy `core_config_id` remain compatible.
- Adding another Core uses additive synchronization so already-running compatible Cores are not unnecessarily stopped.

## Status and uninstall

```bash
sudo manubis-node status
sudo manubis-node uninstall
```

Persistent data remains under the Node data directory configured by the installer.

## Matching Panel

```bash
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Panel/feature/amnezia-wg/install-manubisguard.sh)" @ install --database timescaledb
```
