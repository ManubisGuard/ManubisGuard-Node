# AmneziaWG runtime requirements

The node runtime supports the panel's AmneziaWG 2.x profile through the dedicated `amneziawg` Linux link type and `awgctrl-go` control API. It does not silently fall back to a plain WireGuard interface.

## Host prerequisite

The node container uses `network_mode: host` and the host kernel, so the **host** must have an AmneziaWG 2.0-compatible kernel module loaded. Installing tools inside the container alone is not sufficient. The current `awgctrl-go` integration targets the AWG 2.x netlink interface; an AWG 3.x module/tools profile is not a compatible substitute for this runtime.

The inspected host is Ubuntu 24.04.5 with kernel `6.8.0-147-generic`. The `amneziawg` module is neither installed nor loaded, and the configured APT sources show no package candidate. For a disposable staging host, use a pinned AWG 2.0 module build such as [Advanced-WG's patched AWG 2.0 kernel module](https://github.com/Advanced-WG/amneziawg-linux-kernel-module-awg), which documents Linux 5.4–7.2 support. Pin matching AWG 2.x tools as well; do not mix AWG 3.x tools/module with this profile.

Do not install or replace kernel modules automatically during a normal panel/node update. Module installation can affect host networking and kernel compatibility and must be an explicit operator action with a maintenance window. The live host was not modified.

## Validation after deployment

1. Build and deploy node runtime version `0.6.0` or newer to a disposable staging node.
2. Confirm the host has the expected AWG 2.x module loaded and inspect `/sys/module/amneziawg/version`.
3. Connect an AmneziaWG core from the panel. If the module is missing or incompatible, startup must fail with an explicit error.
4. Inspect the interface with `ip -d link show <interface-name>`; its link kind must be `amneziawg`, not `wireguard`.
5. Confirm `latest handshake` advances and RX/TX counters increase for a real AWG client, then verify traffic through the tunnel.

The ordinary WireGuard path remains on the existing `wireguard` link type. Do not use the absence of a node API error as proof of an AWG handshake; the end-to-end traffic check is required.
