<!-- PasarGuard Pro Node | Bilingual documentation -->

<p align="center">
  <a href="https://github.com/ManubisGuard/ManubisGuard-Node">
    <img src="https://img.shields.io/github/license/ManubisGuard/ManubisGuard-Node?style=flat-square" alt="License">
  </a>
  <a href="https://github.com/ManubisGuard/ManubisGuard-Node/actions">
    <img src="https://img.shields.io/github/actions/workflow/status/ManubisGuard/ManubisGuard-Node/docker-build.yml?style=flat-square" alt="Build status">
  </a>
  <img src="https://img.shields.io/badge/AmneziaWG-2.x-blue?style=flat-square" alt="AmneziaWG 2.x">
</p>

<h1 align="center">PasarGuard Pro Node</h1>
<p align="center">A secure, containerized worker node for PasarGuard Pro</p>

<p align="center">
  <a href="#english">🇬🇧 English</a> ·
  <a href="#فارسی">🇮🇷 فارسی</a>
</p>

> **Project identity:** The runtime is published from the ManubisGuard GitHub organization and is intended to pair with [PasarGuard Pro Panel](https://github.com/ManubisGuard/ManubisGuard-Panel). Repository names and image paths remain unchanged for compatibility.

---

<a id="english"></a>

# 🇬🇧 English

## Overview

**PasarGuard Pro Node** is the worker component managed by the PasarGuard Pro Panel. It receives node configuration from the Panel and runs the supported networking runtimes on the host. The node is containerized, while networking features that depend on the Linux kernel must be installed on the host.

### Key capabilities

- Panel-managed node configuration and runtime lifecycle
- WireGuard and native **AmneziaWG 2.x** runtime support
- Multiple WireGuard/AmneziaWG interfaces where configured by the Panel
- Xray runtime support
- REST or gRPC node transport
- TLS certificate support, including self-signed certificates and user-provided certificate/key pairs
- Interactive and unattended installation
- Docker Compose deployment and systemd service integration
- Host-side AmneziaWG kernel module setup with DKMS and module-load persistence
- Startup validation that fails clearly when the required AWG kernel support is unavailable

## Compatibility

- **Operating system:** Linux host with a supported kernel and matching kernel headers. The installer targets Debian/Ubuntu-family environments; review the installer before adapting it to another distribution.
- **Container runtime:** Docker Engine and the Docker Compose plugin.
- **AmneziaWG:** Current runtime work targets AmneziaWG 2.x. Do not assume AWG3 compatibility unless it is separately documented and tested.
- **Panel compatibility:** Use a Panel build that supports the Node runtime version required by the selected Core. The Panel's AWG runtime guard rejects older node runtimes rather than treating an unsupported runtime as ready.
- **Privileges:** Installation requires root or sudo access because it installs system packages, configures services, and may install a kernel module.

## Quick installation

Review the installer and its options before running it on a production host. The command below downloads the current installer from the repository's `main` branch and runs the interactive installation:

```bash
curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/main/install.sh | sudo bash -s -- install
```

The installer validates the host, prepares Docker if needed, configures the node, installs the required host-side AmneziaWG kernel module when selected/needed by the installer, pulls the published image, configures the service, and checks the selected ports.

### Interactive setup

The installer can prompt for:

- Node name
- TLS mode: self-signed certificate or a supplied public certificate and private key
- Additional certificate Subject Alternative Name (SAN) entries
- API key or automatic UUID generation
- REST or gRPC transport
- Service port (default: `62050`)
- API port (default: `62051`)
- systemd service installation and startup

At the end of installation, the public TLS certificate is printed with `cat` so it can be copied into the Panel's node configuration. **The private key is not printed.** Treat the API key as a secret and do not publish it.

### Unattended installation

Example using a self-signed certificate and gRPC:

```bash
curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/main/install.sh | sudo bash -s -- install \
  --name node-de1 \
  --service-port 62050 \
  --api-port 62051 \
  --use-grpc \
  --self-signed \
  --yes
```

For an existing certificate and private key, run the local installer and provide the paths supported by that version:

```bash
sudo bash install.sh install \
  --cert-path /path/to/cert.pem \
  --key-path /path/to/key.pem \
  --service-port 62050 \
  --api-port 62051
```

Check `install.sh --help` or the installer's built-in usage output if an option differs in your checked-out revision.

## Prebuilt image and updates

The Compose configuration uses the published image:

```text
ghcr.io/manubisguard/manubisguard-node:latest
```

The installer pulls this image rather than compiling the Go runtime on the node. The published image is built for supported multi-architecture targets. The host kernel module is separate from the container image because the container shares the host network and kernel.

To pull the latest image and recreate the node service using the installed CLI:

```bash
sudo pg-node pull
sudo pg-node up -d
```

Plan node updates during a maintenance window if active users depend on the node. Confirm Panel/Node compatibility and verify connectivity after an update.

## AmneziaWG host prerequisites

AmneziaWG requires compatible kernel support on the **host**. The installer prepares build tools and matching kernel headers, installs the project's pinned AWG kernel module through DKMS, loads the module, configures it for automatic loading at boot, and performs a link-type preflight check.

Important notes:

- The kernel module is not supplied by the container image.
- Kernel headers must match the running kernel, and the host must be able to build/load the module.
- Secure Boot may block an unsigned module. If module loading fails, follow the installer's signing-key/MOK guidance instead of bypassing host security controls.
- A successful install-time preflight confirms that the host exposes the expected AWG link type; it does not by itself prove end-to-end client connectivity.
- Keep the host kernel and module update process aligned. Re-check module status after kernel upgrades.

## CLI and service management

The default installed CLI name is `pg-node`:

```bash
sudo pg-node status
sudo pg-node logs
sudo pg-node restart
sudo pg-node pull
sudo pg-node up -d
```

The systemd service is:

```bash
sudo systemctl status pg-node-service
sudo systemctl restart pg-node-service
sudo journalctl -u pg-node-service -f
```

Use the CLI's own help output to confirm commands supported by the installed version.

## Runtime files

Typical installation paths:

| Purpose | Path |
|---|---|
| Source and Compose files | `/opt/manubisguard-node` |
| Environment file | `/opt/manubisguard-node/.env` |
| Persistent node data | `/var/lib/pg-node` |
| Public TLS certificate | `/var/lib/pg-node/certs/ssl_cert.pem` |
| TLS private key | `/var/lib/pg-node/certs/ssl_key.pem` |
| Compose definition | `/opt/manubisguard-node/docker-compose.yml` |

The self-signed certificate includes localhost, loopback, the detected public IP, and any additional SAN entries supplied during setup. Verify the certificate's names and trust model before using it in a production environment.

## Configuration overrides

The installer supports these environment overrides:

```bash
MANUBISGUARD_NODE_REPO=https://github.com/ManubisGuard/ManubisGuard-Node.git
MANUBISGUARD_NODE_BRANCH=main
MANUBISGUARD_NODE_NAME=pg-node
MANUBISGUARD_NODE_DIR=/opt/manubisguard-node
MANUBISGUARD_NODE_DATA_DIR=/var/lib/pg-node
```

Use overrides deliberately and verify the selected branch before installing on production systems.

## Troubleshooting

- **AWG interface cannot start:** Check `dkms status`, `lsmod | grep amneziawg`, kernel/header compatibility, and the service logs.
- **Secure Boot prevents module loading:** Follow the installer's module-signing and MOK enrollment instructions.
- **Node is unreachable:** Confirm the service/API ports are free and allowed by the host firewall and provider firewall; verify the selected REST/gRPC mode and TLS configuration in the Panel.
- **Image update does not take effect:** Pull the image, recreate the Compose service, and confirm the running container's image ID.
- **Panel rejects an AWG node:** Check the runtime version and use a compatible Panel/Node pair.

Never share API keys, private keys, or complete environment files in public issue reports. Redact secrets before attaching logs.

## Development and tests

The repository includes runtime and installer regression tests. Useful local checks include:

```bash
bash -n install.sh
bash tests/test_installer_amneziawg.sh
go test ./backend/wireguard -count=1
git diff --check
```

Run tests from the correct project/container environment where Go dependencies and required system libraries are available. Integration tests that require kernel AWG support should be run on an isolated test host, not against a production node.

## Related projects

- **Panel:** [ManubisGuard/ManubisGuard-Panel](https://github.com/ManubisGuard/ManubisGuard-Panel)
- **Node source:** [ManubisGuard/ManubisGuard-Node](https://github.com/ManubisGuard/ManubisGuard-Node)
- **AmneziaWG kernel module:** [Advanced-WG/amneziawg-linux-kernel-module-awg](https://github.com/Advanced-WG/amneziawg-linux-kernel-module-awg)

## Security

Please report vulnerabilities privately to the repository maintainers when possible. Do not publish credentials, certificates containing private material, production IP allowlists, or unredacted environment files in issues.

---

<a id="فارسی"></a>

# 🇮🇷 فارسی

## معرفی

**PasarGuard Pro Node** بخش اجرایی و Worker پلتفرم PasarGuard Pro است. پنل تنظیمات نود را ارسال می‌کند و نود Runtimeهای شبکه را روی میزبان اجرا می‌کند. خود برنامه داخل Container اجرا می‌شود، اما قابلیت‌هایی که به Kernel لینوکس وابسته‌اند باید روی Host آماده شوند.

### قابلیت‌های اصلی

- دریافت تنظیمات و مدیریت چرخه اجرای نود از پنل
- پشتیبانی از WireGuard و Runtime بومی **AmneziaWG 2.x**
- اجرای هم‌زمان چند Interface از نوع WireGuard/AmneziaWG در صورت تنظیم در پنل
- پشتیبانی از Runtime مربوط به Xray
- ارتباط نود با REST یا gRPC
- TLS با Certificate خودامضا یا Certificate/Private Key ارائه‌شده توسط مدیر
- نصب تعاملی و نصب بدون دخالت کاربر
- استقرار با Docker Compose و مدیریت سرویس با systemd
- آماده‌سازی Kernel Module مربوط به AmneziaWG روی Host با DKMS و فعال‌سازی بارگذاری هنگام بوت
- بررسی اولیه Runtime و ارائه خطای واضح در صورت نبود پشتیبانی لازم AWG در Kernel

## سازگاری و پیش‌نیازها

- **سیستم‌عامل:** میزبان لینوکسی با Kernel پشتیبانی‌شده و Headerهای منطبق. Installer برای محیط‌های خانواده Debian/Ubuntu طراحی شده است؛ پیش از استفاده روی توزیع دیگر، اسکریپت را بررسی کنید.
- **Container Runtime:** Docker Engine و افزونه Docker Compose.
- **AmneziaWG:** تمرکز نسخه فعلی روی AmneziaWG 2.x است. سازگاری AWG3 را بدون مستندات و تست مستقل فرض نکنید.
- **سازگاری با پنل:** نسخه‌ای از Panel استفاده کنید که Runtime نود موردنیاز Core انتخابی را پشتیبانی کند. بررسی سازگاری AWG در پنل، Runtime قدیمی را به‌عنوان نود آماده قبول نمی‌کند.
- **دسترسی:** نصب به دسترسی root یا sudo نیاز دارد؛ چون ممکن است Package، سرویس سیستم و Kernel Module نصب یا تنظیم شوند.

## نصب سریع

پیش از اجرا روی سرور Production، Installer و گزینه‌های آن را بررسی کنید. فرمان زیر Installer شاخه `main` را دریافت و نصب تعاملی را اجرا می‌کند:

```bash
curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/main/install.sh | sudo bash -s -- install
```

Installer میزبان را بررسی می‌کند، در صورت نیاز Docker را آماده می‌کند، تنظیمات نود را می‌گیرد، Kernel Module لازم AmneziaWG را در صورت نیاز نصب می‌کند، Image منتشرشده را دریافت می‌کند، سرویس را تنظیم می‌کند و پورت‌های انتخاب‌شده را بررسی می‌کند.

### گزینه‌های نصب تعاملی

در فرایند نصب ممکن است این موارد از شما پرسیده شوند:

- نام نود
- روش TLS: گواهی خودامضا یا Certificate و Private Key معتبر
- مقادیر اضافی Subject Alternative Name یا SAN
- API Key یا تولید خودکار UUID
- روش ارتباط REST یا gRPC
- پورت سرویس با مقدار پیش‌فرض `62050`
- پورت API با مقدار پیش‌فرض `62051`
- نصب و فعال‌سازی سرویس systemd

در پایان نصب، محتوای Certificate عمومی TLS با دستور `cat` نمایش داده می‌شود تا بتوانید آن را مستقیماً در فرم نود پنل کپی کنید. **Private Key نمایش داده نمی‌شود.** API Key محرمانه است و نباید منتشر شود.

### نصب بدون دخالت کاربر

نمونه نصب با Certificate خودامضا و ارتباط gRPC:

```bash
curl -fsSL https://raw.githubusercontent.com/ManubisGuard/ManubisGuard-Node/main/install.sh | sudo bash -s -- install \
  --name node-de1 \
  --service-port 62050 \
  --api-port 62051 \
  --use-grpc \
  --self-signed \
  --yes
```

برای Certificate و Private Key موجود، Installer محلی را اجرا کنید و مسیرهای مورد پشتیبانی همان نسخه را بدهید:

```bash
sudo bash install.sh install \
  --cert-path /path/to/cert.pem \
  --key-path /path/to/key.pem \
  --service-port 62050 \
  --api-port 62051
```

اگر گزینه‌ای در نسخه مورد استفاده متفاوت است، راهنمای داخلی Installer یا خروجی Help همان نسخه را بررسی کنید.

## Image آماده و به‌روزرسانی

فایل Compose از Image منتشرشده زیر استفاده می‌کند:

```text
ghcr.io/manubisguard/manubisguard-node:latest
```

Installer به‌جای کامپایل Runtime با Go روی نود، Image آماده را دریافت می‌کند. Image منتشرشده برای معماری‌های پشتیبانی‌شده ساخته می‌شود. Kernel Module مستقل از Image است، چون Container از Network و Kernel میزبان استفاده می‌کند.

برای دریافت Image جدید و بازسازی سرویس با CLI نصب‌شده:

```bash
sudo pg-node pull
sudo pg-node up -d
```

اگر کاربران فعال به نود وابسته‌اند، به‌روزرسانی را در بازه نگهداری انجام دهید. سازگاری نسخه Panel/Node و اتصال پس از به‌روزرسانی را بررسی کنید.

## پیش‌نیازهای AmneziaWG روی Host

AmneziaWG به پشتیبانی سازگار Kernel روی **Host** نیاز دارد. Installer ابزارهای ساخت و Headerهای منطبق را آماده می‌کند، Kernel Module نسخه پین‌شده را با DKMS نصب می‌کند، Module را بارگذاری می‌کند، بارگذاری خودکار هنگام بوت را تنظیم می‌کند و نوع Interface مورد انتظار AWG را بررسی می‌کند.

نکات مهم:

- Kernel Module داخل Container ارائه نمی‌شود.
- Headerها باید با Kernel در حال اجرا منطبق باشند و Host باید امکان Build و Load کردن Module را داشته باشد.
- Secure Boot ممکن است از بارگذاری Module امضانشده جلوگیری کند. در صورت شکست، راهنمای امضای Module و ثبت MOK را دنبال کنید و کنترل‌های امنیتی Host را دور نزنید.
- موفقیت بررسی اولیه نصب فقط وجود نوع Interface مورد انتظار را تأیید می‌کند و به‌تنهایی اثبات‌کننده اتصال کامل کلاینت نیست.
- پس از به‌روزرسانی Kernel، وضعیت Module را دوباره بررسی کنید.

## مدیریت با CLI و systemd

نام پیش‌فرض CLI نصب‌شده `pg-node` است:

```bash
sudo pg-node status
sudo pg-node logs
sudo pg-node restart
sudo pg-node pull
sudo pg-node up -d
```

سرویس systemd:

```bash
sudo systemctl status pg-node-service
sudo systemctl restart pg-node-service
sudo journalctl -u pg-node-service -f
```

برای اطمینان از فرمان‌های قابل استفاده در نسخه نصب‌شده، خروجی Help همان CLI را بررسی کنید.

## مسیرهای نصب

| کاربرد | مسیر |
|---|---|
| سورس و فایل‌های Compose | `/opt/manubisguard-node` |
| فایل Environment | `/opt/manubisguard-node/.env` |
| داده‌های پایدار نود | `/var/lib/pg-node` |
| Certificate عمومی TLS | `/var/lib/pg-node/certs/ssl_cert.pem` |
| Private Key TLS | `/var/lib/pg-node/certs/ssl_key.pem` |
| تعریف Compose | `/opt/manubisguard-node/docker-compose.yml` |

Certificate خودامضا شامل localhost، loopback، IP عمومی تشخیص‌داده‌شده و SANهای اضافی واردشده هنگام نصب است. پیش از استفاده Production، نام‌های Certificate و مدل اعتماد آن را بررسی کنید.

## متغیرهای تنظیمات Installer

```bash
MANUBISGUARD_NODE_REPO=https://github.com/ManubisGuard/ManubisGuard-Node.git
MANUBISGUARD_NODE_BRANCH=main
MANUBISGUARD_NODE_NAME=pg-node
MANUBISGUARD_NODE_DIR=/opt/manubisguard-node
MANUBISGUARD_NODE_DATA_DIR=/var/lib/pg-node
```

از Overrideها آگاهانه استفاده کنید و قبل از نصب روی Production مطمئن شوید Branch انتخاب‌شده همان نسخه موردنظر است.

## عیب‌یابی

- **Interface مربوط به AWG اجرا نمی‌شود:** خروجی `dkms status`، نتیجه `lsmod | grep amneziawg`، سازگاری Kernel/Header و Log سرویس را بررسی کنید.
- **Secure Boot مانع بارگذاری Module شده است:** راهنمای امضای Module و ثبت MOK در Installer را دنبال کنید.
- **نود در دسترس نیست:** آزاد بودن پورت‌های سرویس و API، Firewall میزبان و ارائه‌دهنده، حالت REST/gRPC و تنظیمات TLS در پنل را بررسی کنید.
- **Image جدید فعال نشده است:** Image را Pull کنید، سرویس Compose را بازسازی کنید و Image ID کانتینر در حال اجرا را بررسی کنید.
- **پنل نود AWG را قبول نمی‌کند:** نسخه Runtime و سازگاری نسخه Panel/Node را بررسی کنید.

API Key، Private Key یا فایل کامل Environment را در Issue عمومی منتشر نکنید. پیش از ارسال Log، Secretها را حذف کنید.

## توسعه و تست

مخزن دارای تست‌های Regression برای Runtime و Installer است. بررسی‌های محلی مفید:

```bash
bash -n install.sh
bash tests/test_installer_amneziawg.sh
go test ./backend/wireguard -count=1
git diff --check
```

تست‌ها را در محیط صحیح پروژه و با Dependencyها و کتابخانه‌های لازم اجرا کنید. تست‌های Integration وابسته به Kernel AWG باید روی میزبان آزمایشی ایزوله اجرا شوند، نه روی نود Production.

## پروژه‌های مرتبط

- **پنل:** [ManubisGuard/ManubisGuard-Panel](https://github.com/ManubisGuard/ManubisGuard-Panel)
- **سورس نود:** [ManubisGuard/ManubisGuard-Node](https://github.com/ManubisGuard/ManubisGuard-Node)
- **Kernel Module مربوط به AmneziaWG:** [Advanced-WG/amneziawg-linux-kernel-module-awg](https://github.com/Advanced-WG/amneziawg-linux-kernel-module-awg)

## امنیت

در صورت امکان، آسیب‌پذیری‌ها را به‌صورت خصوصی به نگهدارندگان مخزن گزارش کنید. اطلاعات ورود، Private Key، فهرست IPهای Production یا فایل Environment بدون حذف Secretها را در Issue منتشر نکنید.
