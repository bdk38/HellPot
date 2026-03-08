
> **Note**: This is a community-maintained fork of HellPot to preserve the project and ensure its continued availability, maintenance and support. All credit for the original creation and design goes to [yunginnanet](https://github.com/yunginnanet). Original repository: [yunginnanet/HellPot](https://github.com/yunginnanet/HellPot)

<div align="center">
  <img src="https://tcp.ac/i/00ctL.gif" alt="HellPot"/>

[![Go Version](https://img.shields.io/github/go-mod/go-version/bdk38/HellPot)](https://github.com/bdk38/HellPot)
[![Go Report Card](https://goreportcard.com/badge/github.com/bdk38/HellPot)](https://goreportcard.com/report/github.com/bdk38/HellPot)
[![GoDoc](https://godoc.org/github.com/bdk38/HellPot?status.svg)](https://godoc.org/github.com/bdk38/HellPot)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/bdk38/HellPot/blob/main/LICENSE)

</div>

<div align="center">
<img width="669" height="214" alt="hellpotv060" src="https://github.com/user-attachments/assets/e8c0a44e-e4e7-4a75-afb8-36ae938fe16b" />
</div>

# Summary

### Grave Consequences

HellPot is an endless HTTP honeypot that sends unruly bots straight to hell. It uses a Markov engine to endlessly feed those that ignore robots.txt get an infinite stream of Nietzsche-themed nonsense while looking *just* real enough to keep them hooked.

---

**Latest release:** **[HellPot 0.6.0-Community](https://github.com/bdk38/HellPot/releases/tag/v0.6.0)** <br>
*pre-built binaries for Linux, macOS, Windows, FreeBSD*

---

## Docker

HellPot includes a modern, secure Dockerfile and docker-compose.yml.

```
git clone https://github.com/bdk38/HellPot.git
```
```
cd HellPot
```

Create folders + copy default config with correct ownership
```
sudo install -d -m 0755 -o 65532 -g 65532 config logs
```
```
sudo install -m 0644 -o 65532 -g 65532 docker_config.toml ./config/config.toml
```
Customize (Optional)
```
sudo nano ./config/config.toml
```
Start
```
docker compose up -d
```
Watch live logs — you'll see the redesigned retro ASCII banner!
```
docker compose logs -f
```

---

### Quick Start (Binary)

Download the latest release from [HellPot 0.6.0-Community](https://github.com/bdk38/HellPot/releases/tag/v0.6.0)

Generate default config:
```
./HellPot --genconfig
```
Customize (Optional)
```
sudo nano ./config/config.toml
```
Run:
```
./HellPot --config config.toml
```

### Building from Source (Requires Go 1.24+)

```
git clone https://github.com/bdk38/HellPot.git
```
```
cd HellPot
```
```
make
```

---
While these upgrades represent meaningful progress, the fork remains a work in progress and further enhancements are planned. Issues, feature requests, and pull requests are warmly welcomed.
