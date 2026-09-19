# PasarGuard-Node
<p align="center">
    <a href="#">
        <img src="https://img.shields.io/github/actions/workflow/status/PasarGuard/node/docker-build.yml?style=flat-square" />
    </a>
    <a href="https://hub.docker.com/r/pasarguard/node" target="_blank">
        <img src="https://img.shields.io/docker/pulls/pasarguard/node?style=flat-square&logo=docker" />
    </a>
    <a href="#">
        <img src="https://img.shields.io/github/license/PasarGuard/node?style=flat-square" />
    </a>
    <a href="#">
        <img src="https://img.shields.io/github/stars/PasarGuard/node?style=social" />
    </a>
</p>

# Documentation
You can find a full guide in docs https://docs.pasarguard.org/en/node/

# One-Click Installation (Recommended)
The easiest way to install PasarGuard Node is using our automated installation script:

```bash
sudo bash -c "$(curl -sL https://github.com/PasarGuard/scripts/raw/main/pg-node.sh)" @ install
```

## AmneziaWG

This fork supports AmneziaWG v2-style kernel interfaces in addition to standard WireGuard. The node uses the AmneziaWG-aware awgctrl-go library and creates interfaces with the Linux `amneziawg` link type when the panel sends an AWG backend config.

### Host requirement

The Linux host running the node must have an AmneziaWG kernel module installed and loaded. The maintained Advanced-WG kernel module documents DKMS installation for Ubuntu/Debian and exposes the `amneziawg` Linux link type.

After installing the module, verify it before starting the node:

```bash
sudo modprobe amneziawg
lsmod | grep amneziawg
ip link add awg-test type amneziawg
ip link del awg-test
```

In Docker, the container needs `CAP_NET_ADMIN` (normally configured as `NET_ADMIN`) and access to the host network namespace if the deployment uses host networking. The kernel module is provided by the host, not by the container image.

For an opt-in real-kernel integration check on a Linux test host, run the node package tests as root (or with `CAP_NET_ADMIN`) after loading the module:

```bash
export PASARGUARD_AWG_INTEGRATION=yesreallydoit
go test ./backend/wireguard -run TestAmneziaWGKernelIntegration -v
```

This test creates a temporary `amneziawg` interface, configures J/S/H/I parameters and a peer through `awgctrl-go`, reads the values back through generic netlink, verifies peer `AdvancedSecurity`, removes the peer, and cleans up the interface. It is intentionally not enabled in ordinary CI because GitHub-hosted runners do not provide the required AmneziaWG kernel module.

AWG parameters (`Jc`, `Jmin`, `Jmax`, `S1-S4`, `H1-H4`, `I1-I5`) are carried in the node backend configuration and applied through the AmneziaWG generic-netlink family. Compatible clients must understand the same AWG version/configuration; ordinary WireGuard clients do not understand the AWG-specific parameters.

# Donation
You can help PasarGuard team with your donations, [Click Here](https://donate.pasarguard.org/)

# Contributors

We ❤️‍🔥 contributors! If you'd like to contribute, please check out our [Contributing Guidelines](CONTRIBUTING.md) and feel free to submit a pull request or open an issue. We also welcome you to join our [Telegram](https://t.me/Pasar_Guard) group for either support or contributing guidance.

Check [open issues](https://github.com/PasarGuard/node/issues) to help the progress of this project.

## Stargazers over time
[![Stargazers over time](https://starchart.cc/PasarGuard/node.svg?variant=adaptive)](https://starchart.cc/PasarGuard/node)
                    
<p align="center">
Thanks to the all contributors who have helped improve PasarGuard Node:
</p>
<p align="center">
<a href="https://github.com/PasarGuard/node/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=PasarGuard/node" />
</a>
</p>
<p align="center">
  Made with <a rel="noopener noreferrer" target="_blank" href="https://contrib.rocks">contrib.rocks</a>
</p>
