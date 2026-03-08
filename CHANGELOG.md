## [0.6.1] - 2026-03-08
### Fixed
- `logger.debug` and `logger.trace` in config.toml now actually work
- CLI flag parsing rewritten with Go's standard `flag` package (more reliable)
- Debug/trace override logic now correctly defaults to INFO level
