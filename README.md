# ManubisGuard Node

<p align="center">
  <strong>Open-source network infrastructure, under your control.</strong>
</p>

<p align="center">
  <a href="#english">🇬🇧 English</a> · <a href="#فارسی">🇮🇷 فارسی</a>
</p>

> **Fork / upstream:** ManubisGuard Node is maintained as a feature-enhanced fork of [PasarGuard/node](https://github.com/PasarGuard/node), with ManubisGuard-specific Node management, Multi-Core behavior, AmneziaWG runtime support and CLI/installer integration.

> **Current development branch:** `feature/amnezia-wg`

---

<a id="english"></a>

# 🇬🇧 English

## What is ManubisGuard Node?

**ManubisGuard Node** is the infrastructure-side component of the ManubisGuard Panel + Node architecture.

The Node runs network Cores and services assigned by the Panel and provides the runtime/control surface required for managing remote infrastructure.

The project is based on the **PasarGuard Node** codebase and keeps the upstream-compatible installation and operational concepts where required while extending them for ManubisGuard.

### Upstream project

- **Upstream Node:** [PasarGuard/node](https://github.com/PasarGuard/node)
- **ManubisGuard Node:** [ManubisGuard/ManubisGuard-Node](https://github.com/ManubisGuard/ManubisGuard-Node)
- **Matching Panel:** [ManubisGuard/ManubisGuard-Panel](https://github.com/ManubisGuard/ManubisGuard-Panel)

---

## ✨ Highlights

- Panel + Node architecture
- Multi-Core Node support
- Multiple WireGuard/AmneziaWG Cores on one Node
- Xray Core support with one Xray process per Node
- Native WireGuard / AmneziaWG runtime integration
- AmneziaWG host preparation
- DKMS/kernel integration for AWG runtime
- Native AWG interface validation
- Node Peer synchronization
- TLS/certificate support
- REST / gRPC transport options
- API key authentication
- Native `manubis` CLI
- `manubis-node` lifecycle command
- systemd service integration
- Core update tooling
- GeoIP/GeoSite update tooling
- Certificate renewal tooling
- Docker-based runtime

---

## 🧩 Multi-Core on one Node

A single ManubisGuard Node can be assigned to multiple Core configurations from the ManubisGuard Panel.

### WireGuard / AmneziaWG

Multiple WireGuard and AmneziaWG Core instances can run concurrently on the same Node.

Each running interface is managed independently by the Node runtime.

### Xray

Xray uses a single Xray process for its inbounds, so only one Xray Core configuration can be selected per Node.

### Compatibility

Existing Nodes that only use the legacy `core_config_id` representation remain compatible while the Panel can use `core_config_ids` for Multi-Core assignments.

### Additive synchronization

Adding another compatible Core is designed to use additive synchronization where possible so already-running compatible Cores are not unnecessarily stopped.

---

## 🛡️ AmneziaWG Runtime

AmneziaWG is a first-class runtime target in the current development branch.

The Node-side work includes:

- Native AmneziaWG interface support
- AWG kernel module preparation
- DKMS support
- Kernel/header checks
- AmneziaWG tools installation
- Runtime/tool version pinning for the validated environment
- Boot-time module handling
- Native interface smoke testing
- WireGuard/AWG peer operation
- Node-side runtime validation

The development/test workflow has been used to validate real external AmneziaWG connectivity, including peer mapping, endpoint delivery, UDP traffic, handshake state and RX/TX counters.

> **Support note:** the current implementation targets **AmneziaWG 2.x / the validated AWG runtime**. Do not assume AWG3 compatibility unless explicitly documented and tested.

### Kernel compatibility

AWG depends on host kernel capabilities and DKMS/module availability. The installer performs kernel/header and build prerequisite checks and prepares the host runtime where required.

A reboot may be required on systems where the installed kernel/module state cannot be activated immediately.

---

## 🔐 TLS, Certificates & Transports

The Node installer supports the interactive upstream-style configuration flow for:

- Service/API port
- API key
- TLS/certificate mode
- Certificate/key paths or generated certificates
- SAN entries where supported
- REST transport
- gRPC transport
- Service installation

Certificate renewal is also available through the native CLI.

---

## 🚀 One-command installation

Install the matching development branch with:

```bash
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/feature/amnezia-wg/install-manubisguard-node.sh)" @ install
```

The interactive installer follows the upstream-compatible Node installation workflow.

A normal interactive installation asks for the Node service port, TLS/certificate mode, API key and transport options.

Use `-y` only when you intentionally want non-interactive defaults.

---

## 🧰 Manubis Node CLI

The Node fork ships the ManubisGuard Node CLI under the `manubis` command.

Common commands:

```bash
sudo manubis status
sudo manubis restart
sudo manubis logs
sudo manubis update
sudo manubis core-update --version latest
sudo manubis geofiles --iran
sudo manubis renew-cert
```

### CLI command set

| Command | Purpose |
|---|---|
| `install` | Install or reinstall a Node instance |
| `update` | Update the Node deployment and optionally its service |
| `uninstall` | Remove the Node instance |
| `up` / `down` | Start or stop the Node stack |
| `restart` | Restart the Node stack |
| `status` | Show Node, port, certificate and Core status |
| `logs` | Follow Node container logs |
| `core-update` | Install or switch Xray-core |
| `geofiles` | Update regional geoip/geosite assets |
| `renew-cert` | Regenerate/renew the Node TLS certificate |
| `edit` / `edit-env` | Edit Compose or environment configuration |
| `install-script` / `uninstall-script` | Install/remove the global CLI |
| `completion` | Install Bash/Zsh completion |
| `version-script` / `script-version` | Show CLI version and commit information |
| `service-install` | Install the Node systemd service |
| `service-uninstall` | Remove the Node systemd service |
| `service-start` / `service-stop` | Start or stop the systemd service |
| `service-restart` | Restart the systemd service |
| `service-status` | Show systemd service status |
| `service-logs` | Follow or inspect service logs |
| `service-update` | Update the service helper |

Global options include `-y/--yes` and `--name`.

Install supports version selection, REST/gRPC selection, service/API ports, API key, TLS certificate/key, SAN entries, self-signed certificates, service installation and override mode.

---

## 📦 CLI installation

The CLI can be installed directly from the repository's installer:

```bash
curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/feature/amnezia-wg/install.sh | sudo bash -s -- install
```

Then verify:

```bash
sudo manubis status
```

The CLI is intentionally named `manubis`.

Existing Node deployments retain their current runtime layout unless the operator explicitly selects another instance configuration.

---

## 🧭 Node lifecycle

The Node also provides the `manubis-node` lifecycle command:

```bash
sudo manubis-node status
sudo manubis-node restart
sudo manubis-node logs
sudo manubis-node uninstall
```

The `manubis` CLI is the preferred full command surface for current ManubisGuard deployments.

---

## 🐳 Runtime

The Node uses a Docker-based runtime with host-side components where required by the selected networking Core.

The installer is responsible for preparing the required runtime files and service configuration.

Persistent data remains under the configured Node data directory.

Existing installations commonly use paths under:

```text
/opt/manubisguard-node
/var/lib/pg-node
```

Do not assume these paths for every custom instance; use the generated deployment configuration as the source of truth.

---

## 🔄 Panel Integration

The Node is designed to be managed from the matching ManubisGuard Panel.

Panel:

https://github.com/ManubisGuard/ManubisGuard-Panel

Example matching Panel installation:

```bash
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Panel/feature/amnezia-wg/install-manubisguard.sh)" @ install --database timescaledb
```

The Panel manages Nodes and Core assignments while the Node provides the infrastructure-side runtime.

---

## 🔐 Security

- Never publish real API keys.
- Never commit private keys or certificates.
- Use TLS for remote Node administration where supported.
- Keep the Node API accessible only from intended management paths.
- Use RBAC and least-privilege administration from the Panel.
- Treat Node credentials as infrastructure secrets.
- Do not expose generated temporary credentials in logs or screenshots.

---

## 🧪 Validation

Node validation includes:

- Installer syntax and execution checks
- Kernel/header checks
- DKMS/module preparation
- Native AWG interface smoke validation
- Core synchronization checks
- Multi-Core behavior checks
- Node lifecycle checks
- TLS/certificate checks
- REST/gRPC configuration checks
- Real external AmneziaWG connectivity validation in TEST

The exact validated environment and remaining gates are maintained in the project development records.

---

## 🤝 Contributing

Contributions are welcome.

Useful areas include:

- Networking
- AmneziaWG
- WireGuard
- Xray
- Node runtime
- Linux/kernel compatibility
- DKMS
- Docker
- CLI tooling
- TLS/certificates
- Testing
- Documentation

Before opening a PR, review the repository's current contribution guidance and project development notes.

---

## 📚 Project Links

- Node: https://github.com/ManubisGuard/ManubisGuard-Node
- Panel: https://github.com/ManubisGuard/ManubisGuard-Panel
- Upstream Node: https://github.com/PasarGuard/node
- Upstream Panel: https://github.com/PasarGuard/panel
- Issues: https://github.com/ManubisGuard/ManubisGuard-Node/issues

---

<a id="فارسی"></a>

# 🇮🇷 فارسی

## ManubisGuard Node چیست؟

**ManubisGuard Node** بخش زیرساختی معماری **Panel + Node** در ManubisGuard است.

Node وظیفه اجرای Coreها و سرویس‌های شبکه‌ای تخصیص‌داده‌شده توسط Panel را بر عهده دارد و Runtime لازم برای مدیریت زیرساخت Remote را فراهم می‌کند.

### این پروژه Fork کدام Node است؟

ManubisGuard Node بر پایه کدبیس **PasarGuard Node** توسعه داده شده است.

- **Upstream Node:** [PasarGuard/node](https://github.com/PasarGuard/node)
- **ManubisGuard Node:** [ManubisGuard/ManubisGuard-Node](https://github.com/ManubisGuard/ManubisGuard-Node)
- **Matching Panel:** [ManubisGuard/ManubisGuard-Panel](https://github.com/ManubisGuard/ManubisGuard-Panel)

قابلیت‌های اختصاصی ManubisGuard مانند Multi-Core، Runtime مربوط به AmneziaWG و CLI/Installer روی این پایه توسعه داده شده‌اند.

---

## ✨ قابلیت‌ها

- معماری Panel + Node
- پشتیبانی از Multi-Core
- چند WireGuard/AmneziaWG Core روی یک Node
- پشتیبانی از Xray با یک Process برای Coreهای Xray
- Native WireGuard / AmneziaWG Runtime
- آماده‌سازی Host برای AmneziaWG
- DKMS و Kernel integration برای AWG
- Native AWG Interface Validation
- Node Peer Synchronization
- TLS و Certificate
- REST / gRPC
- API Key Authentication
- Native `manubis` CLI
- `manubis-node` lifecycle command
- Systemd Service
- Core Update
- GeoIP/GeoSite Update
- Certificate Renewal
- Docker Runtime

---

## 🧩 Multi-Core روی یک Node

یک Node در ManubisGuard می‌تواند به چند Core Configuration متصل شود.

### WireGuard / AmneziaWG

چند Core از نوع WireGuard و AmneziaWG می‌توانند به صورت همزمان روی یک Node اجرا شوند.

هر Interface به صورت مستقل توسط Runtime مدیریت می‌شود.

### Xray

Xray برای Inboundها از یک Process استفاده می‌کند؛ بنابراین در هر Node فقط یک Xray Core Configuration قابل انتخاب است.

### Compatibility

Nodeهای قدیمی که فقط `core_config_id` دارند همچنان قابل استفاده هستند و Panel جدید می‌تواند از `core_config_ids` برای Multi-Core استفاده کند.

### Additive Synchronization

در صورت امکان اضافه شدن Core جدید به شکل Additive انجام می‌شود تا Coreهای سازگار در حال اجرا بدون دلیل متوقف نشوند.

---

## 🛡️ AmneziaWG Runtime

AmneziaWG یکی از اهداف اصلی Runtime در Branch فعلی است.

موارد سمت Node شامل:

- Native AmneziaWG Interface
- آماده‌سازی Kernel Module
- DKMS
- بررسی Kernel و Header
- نصب AWG Tools
- Pin کردن Runtime/Tools در محیط تأییدشده
- مدیریت Load شدن Module
- Native Interface Smoke Test
- WireGuard/AWG Peer Runtime
- Runtime Validation

در محیط TEST اتصال واقعی خارجی AmneziaWG نیز بررسی شده است؛ از جمله Peer Mapping، Endpoint، UDP Traffic، Handshake و RX/TX Counters.

> **نکته:** پیاده‌سازی فعلی روی **AmneziaWG 2.x و Runtime تأییدشده** متمرکز است. سازگاری با AWG3 نباید بدون تست صریح فرض شود.

### Kernel Compatibility

AmneziaWG به وضعیت Kernel و امکان ساخت/Load شدن Module وابسته است. Installer موارد لازم مربوط به Kernel، Header و DKMS را بررسی و در صورت نیاز آماده می‌کند.

در بعضی سیستم‌ها برای فعال شدن Kernel/Module جدید ممکن است Reboot لازم باشد.

---

## 🔐 TLS، Certificate و Transport

Installer Node Workflow مربوط به موارد زیر را پوشش می‌دهد:

- Service/API Port
- API Key
- TLS/Certificate Mode
- Certificate/Key
- SAN در صورت پشتیبانی
- REST
- gRPC
- Service Installation

Certificate Renewal نیز از طریق CLI قابل انجام است.

---

## 🚀 نصب یک‌دستوری

برای نصب Branch توسعه فعلی:

```bash
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/feature/amnezia-wg/install-manubisguard-node.sh)" @ install
```

Installer از Workflow سازگار با PasarGuard استفاده می‌کند.

در حالت Interactive مواردی مانند Port، TLS/Certificate، API Key و Transport از کاربر دریافت می‌شوند.

گزینه `-y` فقط برای نصب Non-interactive با Defaultهای مورد نظر استفاده شود.

---

## 🧰 Manubis CLI

Node دارای CLI اصلی با نام `manubis` است.

```bash
sudo manubis status
sudo manubis restart
sudo manubis logs
sudo manubis update
sudo manubis core-update --version latest
sudo manubis geofiles --iran
sudo manubis renew-cert
```

### دستورات

| دستور | کاربرد |
|---|---|
| `install` | نصب یا نصب مجدد Node |
| `update` | به‌روزرسانی Node و Service |
| `uninstall` | حذف Node |
| `up` / `down` | Start/Stop کردن Stack |
| `restart` | Restart کردن Node |
| `status` | نمایش وضعیت Node، Port، Certificate و Core |
| `logs` | مشاهده Logهای Container |
| `core-update` | نصب یا تغییر Xray-core |
| `geofiles` | به‌روزرسانی GeoIP/GeoSite |
| `renew-cert` | Renewal/Regeneration Certificate |
| `edit` / `edit-env` | ویرایش Compose یا Environment |
| `install-script` / `uninstall-script` | نصب/حذف CLI سراسری |
| `completion` | نصب Bash/Zsh Completion |
| `version-script` / `script-version` | نمایش Version و Commit |
| `service-install` | نصب Systemd Service |
| `service-uninstall` | حذف Systemd Service |
| `service-start` / `service-stop` | Start/Stop Service |
| `service-restart` | Restart Service |
| `service-status` | وضعیت Systemd Service |
| `service-logs` | مشاهده Logهای Service |
| `service-update` | Update کردن Service Helper |

Global Optionهای مهم شامل `-y/--yes` و `--name` هستند.

---

## 📦 نصب CLI

```bash
curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/feature/amnezia-wg/install.sh | sudo bash -s -- install
```

بعد از نصب:

```bash
sudo manubis status
```

نام CLI عمداً `manubis` انتخاب شده است.

---

## 🧭 Node Lifecycle

Command قدیمی‌تر `manubis-node` نیز برای Lifecycle در دسترس است:

```bash
sudo manubis-node status
sudo manubis-node restart
sudo manubis-node logs
sudo manubis-node uninstall
```

برای Deploymentهای فعلی، `manubis` سطح کامل‌تری از CLI را ارائه می‌دهد.

---

## 🐳 Runtime

Node از Docker Runtime به همراه اجزای Host-side مورد نیاز Core انتخاب‌شده استفاده می‌کند.

Installer فایل‌های Runtime و Service Configuration لازم را آماده می‌کند.

داده‌های دائمی در Data Directory مربوط به Instance نگهداری می‌شوند.

در Deploymentهای فعلی مسیرهایی مانند موارد زیر دیده می‌شوند:

```text
/opt/manubisguard-node
/var/lib/pg-node
```

در Instanceهای سفارشی نباید این مسیرها را بدون بررسی Configuration قطعی فرض کرد.

---

## 🔄 اتصال به Panel

Node برای مدیریت شدن توسط ManubisGuard Panel طراحی شده است.

Panel:

https://github.com/ManubisGuard/ManubisGuard-Panel

نمونه نصب Panel هماهنگ:

```bash
sudo bash -c "$(curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Panel/feature/amnezia-wg/install-manubisguard.sh)" @ install --database timescaledb
```

Panel مدیریت Node و Core Assignment را انجام می‌دهد و Node Runtime سمت زیرساخت را اجرا می‌کند.

---

## 🔐 امنیت

- API Key واقعی را منتشر نکنید.
- Private Key و Certificate را Commit نکنید.
- برای مدیریت Remote در صورت امکان TLS استفاده کنید.
- API Node را فقط از مسیرهای مدیریتی مورد نیاز در دسترس قرار دهید.
- از RBAC و Least Privilege استفاده کنید.
- Credentialهای Node را Secret زیرساختی در نظر بگیرید.
- Temporary Credential را در Log یا Screenshot منتشر نکنید.

---

## 🧪 Validation

موارد Validation شامل:

- بررسی Syntax و اجرای Installer
- Kernel/Header checks
- DKMS/Module preparation
- Native AWG interface smoke test
- Core synchronization
- Multi-Core behavior
- Node lifecycle
- TLS/Certificate
- REST/gRPC
- اتصال واقعی خارجی AmneziaWG در TEST

جزئیات محیط تأییدشده و موارد باقی‌مانده در گزارش‌های توسعه پروژه نگهداری می‌شوند.

---

## 🤝 مشارکت

Contributionها استقبال می‌شوند.

زمینه‌های مناسب:

- Networking
- AmneziaWG
- WireGuard
- Xray
- Node Runtime
- Linux/Kernel Compatibility
- DKMS
- Docker
- CLI
- TLS/Certificates
- Testing
- Documentation

قبل از PR، راهنمای Repository و گزارش‌های توسعه را بررسی کنید.

---

## 📚 لینک‌های پروژه

- Node: https://github.com/ManubisGuard/ManubisGuard-Node
- Panel: https://github.com/ManubisGuard/ManubisGuard-Panel
- Upstream Node: https://github.com/PasarGuard/node
- Upstream Panel: https://github.com/PasarGuard/panel
- Issues: https://github.com/ManubisGuard/ManubisGuard-Node/issues
