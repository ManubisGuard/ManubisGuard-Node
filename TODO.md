# ManubisGuard Node TODO / Incident Report

Updated: 2026-10-18

## Fixed: AmneziaWG node installation stopped after a false reboot requirement

### Root cause
- The installer generated a probe interface named `manubis-awg-check-$$`.
- Linux network interface names are limited to 15 bytes, so this name was rejected even though the AmneziaWG kernel module was loaded and functional.
- This false negative caused the installer to request a reboot and stop.

### Source fixes committed on `feature/amnezia-wg`
- `e1cb917509d257314e540f666a7ea13df47f08d7`: use a short AmneziaWG probe interface name.
- `6d972d7b495868e42e07f905ea683bdc8fd0b8a9`: pass API key and runtime paths into the container through Compose interpolation.
- `e7f8eb95b5b97d5d345fd950ffbee1e2b65d296e`: align SSL and generated-config paths with the selected instance data directory.

### Additional deployment issue found and fixed on the server
- Compose did not pass `API_KEY` into the container, and it retained legacy `/var/lib/pg-node` paths while the instance used `/var/lib/manubis-node`.
- Updated `/opt/manubis-node/docker-compose.yml` to pass `API_KEY` and use values from `.env`.
- Set `GENERATED_CONFIG_PATH` to `/var/lib/manubis-node/generated`.
- Backed up the prior Compose file under `/var/backups/manubis-node/` before replacing it.

### Validation performed on 2026-10-18
- Running kernel: `6.8.0-147-generic`.
- DKMS module for the running kernel: installed.
- `modprobe amneziawg`: successful.
- Native interface creation test: successful.
- Installer passed the host-runtime check after correcting the probe interface name.
- Container image pulled and `manubis-node` started successfully.
- `manubis-node status`: Up.
- gRPC listener: `0.0.0.0:62050`.
- Container restart policy: `always`; restart count was zero at validation.
- Recent container logs confirm: `Starting Node: v0.5.4` and `gRPC Server listening on 0.0.0.0:62050`.

### Remaining checks
- Register the node in the ManubisGuard panel using the node's API key and generated certificate; do not place the API key in public logs or documentation.
- Verify a real panel-to-node connection and confirm traffic/config generation end to end.
- The container logs a benign warning that its working-directory `.env` file is absent. Docker environment variables are present and the node starts; clean up this warning only if a supported container configuration is identified.
- Test a clean installation on a fresh host to verify the installer and Compose fixes without relying on packages already present on this server.


## Repository audit: installer, CLI bootstrap, CI and image publishing

Updated: 2026-10-18

### Fixed in this audit
- The installer was looking for shared libraries under `$SCRIPT_DIR/lib`, although this repository stores them under `scripts/lib`. It now checks the actual in-repository path.
- The installer fallback/bootstrap previously used the PasarGuard shared-library repository and legacy install directory. It now uses ManubisGuard-Node's `scripts/lib` and `/usr/local/lib/manubisguard-node/lib`.
- The installer now points its shared-library installation source at this repository instead of `PasarGuard/scripts`.
- The installed CLI fallback had the same wrong raw URL shape (`/lib` instead of `/scripts/lib`). Corrected it to the ManubisGuard repository path.
- Updated installer output text that still referred to the PasarGuard Panel.
- CI now checks shell syntax, Docker Compose configuration, all Go packages, and builds the Node.
- The development Docker workflow no longer tries to publish images under the PasarGuard namespace or requires Docker Hub secrets; it is build-only on main/manual dispatch.
- The release workflow now publishes release images to `ghcr.io/manubisguard/manubisguard-node` instead of the old PasarGuard image namespace.

### Current verification status
- GitHub Actions runs were queued/in progress immediately after these commits. Do not mark CI or image publication as passed until the latest commit's workflow runs finish successfully.
- The running production-like node on host `8017638839-819f55` was verified separately to have a running gRPC listener on TCP 62050; this does not substitute for a clean-host install or panel-to-node integration test.

### Known follow-up work
- The Go module path remains `github.com/pasarguard/node`; a full module-path migration needs a dedicated audited change across imports, generated code and release artifacts.
- `Makefile` still calls upstream PasarGuard scripts for Xray/core installation; node-serviced is still sourced from the upstream release repository. These are remaining upstream dependencies, not fixed by this audit.
- `Dockerfile.wireguard` and `Dockerfile.xray` retain upstream image-source labels and need consistent ManubisGuard metadata if those images remain supported.
- Need run a clean-host installation, register the node in ManubisGuard Panel, and verify certificate/API-key authentication plus config generation and traffic accounting end to end.
