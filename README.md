
> **Note**: This is a community-maintained fork of HellPot to preserve the project and ensure its continued availability, maintenance and support. All credit for the original creation and design goes to [yunginnanet](https://github.com/yunginnanet). Original repository: [yunginnanet/HellPot](https://github.com/yunginnanet/HellPot)

<div align="center">
  <img src="https://tcp.ac/i/00ctL.gif" alt="HellPot"/>

[![Go Version](https://img.shields.io/github/go-mod/go-version/bdk38/HellPot)](https://github.com/bdk38/HellPot)
[![Go Report Card](https://goreportcard.com/badge/github.com/bdk38/HellPot)](https://goreportcard.com/report/github.com/bdk38/HellPot)
[![GoDoc](https://godoc.org/github.com/bdk38/HellPot?status.svg)](https://godoc.org/github.com/bdk38/HellPot)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/bdk38/HellPot/blob/main/LICENSE)

</div>

# Summary

### Grave Consequences

HellPot is an endless HTTP honeypot that sends unruly bots straight to hell.
It uses a Markov engine to endlessly feed clients that ignore `robots.txt` an
infinite stream of classic-literature nonsense while looking *just* real enough
to keep them hooked.

---

**Latest release:** **[HellPot 0.7.2-Community](https://github.com/bdk38/HellPot/releases/tag/v0.7.2)**  
*pre-built binaries for Linux and macOS*

See [CHANGELOG.md](CHANGELOG.md) and [docs/RELEASING.md](docs/RELEASING.md).

---

## Docker

No host config/log directories to create. Compose uses named volumes that are
seeded from the image on first start.

```bash
git clone https://github.com/bdk38/HellPot.git
cd HellPot
docker compose up -d --build
docker compose logs -f
```

HellPot listens on host port **8080**.

### Edit the running config

```bash
docker compose cp hellpot:/config/config.toml ./config.toml
$EDITOR ./config.toml
docker compose cp ./config.toml hellpot:/config/config.toml
docker compose restart
```

### Config files in this repo

| File | Purpose |
|------|---------|
| `config.toml` | Host/binary default (same content as the embedded `--genconfig` template) |
| `internal/config/default_config.toml` | Embedded template used by `--genconfig` |
| `docker_config.toml` | Image/compose default — full config with Docker-only overrides (`bind_addr = 0.0.0.0`, logs under `/logs`, `docker_logging = true`) |

### Bind-mount config/logs on the host (optional)

If you prefer editing files directly under `./config` and `./logs`:

```bash
mkdir -p config logs
cp docker_config.toml config/config.toml
```

Then in `docker-compose.yml` switch the volume lines to:

```yaml
- ./config:/config
- ./logs:/logs
```

Only if the container cannot write logs (permission denied), fix ownership once:

```bash
sudo chown -R 65532:65532 config logs
```

---

## Quick Start (Binary)

Download the latest release from
[HellPot 0.7.2-Community](https://github.com/bdk38/HellPot/releases/tag/v0.7.2).

```bash
./HellPot --genconfig
./HellPot --config config.toml
```

### Building from Source

Requires the Go version in [`go.mod`](go.mod) (currently **Go 1.26.5**).

```bash
git clone https://github.com/bdk38/HellPot.git
cd HellPot
make
```

---

## Maintained dependency forks

Upstream `github.com/mitchellh/reflectwalk` and
`github.com/mitchellh/copystructure` are archived. This fork pins maintained
replacements via `replace` in `go.mod`:

- https://github.com/bdk38/reflectwalk
- https://github.com/bdk38/copystructure

---

## Releasing / contributing

- Release process: [docs/RELEASING.md](docs/RELEASING.md)
- Security policy: [SECURITY.md](SECURITY.md)
- Changelog: [CHANGELOG.md](CHANGELOG.md)

While these upgrades represent meaningful progress, the fork remains a work in
progress and further enhancements are planned. Issues, feature requests, and
pull requests are warmly welcomed.

## Development Transparency

Some features and fixes in HellPot were developed with the assistance of AI
tools. See
[AI assistance in Development discussion](https://github.com/bdk38/HellPot/discussions/67)
for full details on process, background, and philosophy around transparency.
