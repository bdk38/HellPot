# Changelog

All notable changes to **HellPot Community** (`bdk38/HellPot`) are documented here.

Format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed
- Repository housecleaning: single canonical release workflow, Dependabot bumps, docs hygiene
- Config surface refactored from many package-level globals into typed structs
  (`HTTPConfig`, `PerformanceConfig`, `LoggerConfig`, `DeceptionConfig`)
- Default `config.toml` documents logging and performance knobs more completely

### Dependencies
- Bump `github.com/knadh/koanf/*` and `github.com/valyala/fasthttp` (and transitive deps)
- Bump GitHub Actions `actions/checkout` and `actions/setup-go` to v7

## [0.7.0] - 2026-05-29

### Added
- Dual logging: system log + access log
- Pre-generated Markov chunk pool for lower per-connection CPU cost
- Per-connection and global bandwidth rate limiting
- Performance config validation with safe fallbacks

### Fixed
- Write granularity under rate limiting (smooth 4KB slices, faster disconnect detection)
- Chunk pool / rate limiter init-order bug (config was always zero during package `init`)
- Infinite loop in chunk pool generation when remaining buffer was too small for a token
- Logger pre-init guards

### Changed
- Release assets limited to Linux + macOS (Windows dropped from release matrix)

## [0.6.4] - 2026-04-12

### Changed
- Dependency maintenance
- Maintained forks for archived `mitchellh/reflectwalk` and `mitchellh/copystructure`

## [0.6.3] - 2026-03-11

### Changed
- Go toolchain updated to 1.25.8
- Docker base image / config refresh

## [0.6.2] - 2026-03-08

### Changed
- Release workflow cleanup (drop 386 builds / GHCR push path from legacy workflow)

## [0.6.1] - 2026-03-08

### Fixed
- `logger.debug` and `logger.trace` in config.toml now actually work
- CLI flag parsing rewritten with Go's standard `flag` package (more reliable)
- Debug/trace override logic now correctly defaults to INFO level

### Changed
- Configuration engine migrated toward koanf/TOML embed defaults

## [0.6.0] - 2026-03-07

### Added
- Community fork baseline under `bdk38/HellPot`
- Docker/distroless packaging improvements
- Community branding / banner updates
