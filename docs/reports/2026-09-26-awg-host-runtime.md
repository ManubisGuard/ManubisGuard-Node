# AmneziaWG host runtime checkpoint — 2026-09-26

TEST verification established that native AmneziaWG requires a host kernel module; the Node container's `awg` userspace binary alone is not sufficient.

## TEST host state verified

- OS: Ubuntu 24.04 LTS (Noble)
- Running kernel: `6.8.0-142-generic`
- Kernel packages: `6.8.0-142.142`
- DKMS: `3.0.11-1ubuntu13`
- AmneziaWG DKMS package: `1.0.0-0~202609061402+4569c4c~ubuntu24.04.1`
- AmneziaWG tools package: `1.0.20210914-0~202608130144+ee0f0a9~ubuntu24.04.1`
- Loaded AmneziaWG/tool release: `3.1.20260812`
- Amnezia PPA: `ppa:amnezia/ppa`
- Native `ip link add <name> type amneziawg`: PASS

## Installer contract

`install-manubisguard-node.sh` now installs the host prerequisites before deploying the Node:

1. apt/dpkg host check
2. kernel compatibility check (AmneziaWG 3.1 requires kernel 6.7+ for the supported runtime path)
3. `linux-generic` bootstrap when the running kernel is too old, followed by an explicit reboot-and-rerun requirement
4. matching kernel headers, DKMS and build prerequisites
5. official Amnezia PPA
6. `amneziawg-dkms` and `amneziawg-tools`
7. `modprobe amneziawg` and boot-time module loading
8. native interface creation smoke test

The Node Docker image pins AmneziaWG tools to `v3.1.20260812` so the container userspace matches the TEST-validated runtime line.

No private keys, API keys, PSKs, or other secrets are stored in this report.
