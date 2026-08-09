
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

**Latest release:** **[HellPot 0.7.1-Community](https://github.com/bdk38/HellPot/releases/tag/v0.7.1)**  
*pre-built binaries for Linux and macOS*

See [CHANGELOG.md](CHANGELOG.md) and [docs/RELEASING.md](docs/RELEASING.md).

---

## Docker

HellPot includes a modern, secure Dockerfile and docker-compose.yml.

```bash
git clone https://github.com/bdk38/HellPot.git
cd HellPot
```

Create folders + copy default config with correct ownership:

```bash
sudo install -d -m 0755 -o 65532 -g 65532 config logs
sudo install -m 0644 -o 65532 -g 65532 docker_config.toml ./config/config.toml
```

Customize (optional):

```bash
sudo nano ./config/config.toml
```

Start and follow logs:

```bash
docker compose up -d
docker compose logs -f
```

---

## Quick Start (Binary)

Download the latest release from
[HellPot 0.7.1-Community](https://github.com/bdk38/HellPot/releases/tag/v0.7.1).

```bash
./HellPot --genconfig
./HellPot --config config.toml
```

### Building from Source

Requires the Go version in [`go.mod`](go.mod) (currently **Go 1.25.8**).

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
