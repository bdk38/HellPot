# PR: Write Granularity Fix, Dual Log System, and Pool Init Corrections

## Summary

This PR addresses four interrelated issues discovered during log system development and testing: a write granularity mismatch that caused delayed data delivery and ghost connections, splits logging into separate system and access logs, fixes a package init ordering bug that silently prevented the chunk pool from ever being created, and fixes an infinite loop in pool chunk generation.

## Changes

### 1. WriteHell Write Granularity (`heffalump/heffalump.go`)

**Problem:** When rate limiting was active, `WriteHell` read an entire chunk (e.g., 128KB) and calculated a single sleep for the full amount before writing. With `baseline_rate_kbps = 2` and 128KB chunks, this produced 64-second sleeps followed by 128KB bursts — the client saw no data for over a minute, and disconnects went undetected for up to 256 seconds because no I/O occurred during sleep.

**Fix:** Introduced `writeSliced()`, which drains each chunk in 4KB increments with proportional micro-sleeps. Added `bw.Flush()` after the `<html><body>` prefix so the client receives data immediately. When rate limiting is disabled, full-buffer writes are preserved for maximum throughput.

**Result:** With `baseline_rate_kbps = 2`, time to first data drops from ~64s to immediate, write frequency increases from every 64s to every ~2s, and disconnect detection drops from up to 256s to 2-4s.

### 2. Dual Log System (`internal/config/logger.go`, `internal/config/globals.go`, `internal/config/config.go`, `internal/config/default_config.toml`, `internal/http/router.go`, `internal/http/robots.go`, `cmd/HellPot/HellPot.go`, `docker_config.toml`)

**Problem:** A single logger handled both HellPot operational events and client connection records. These have different audiences and different needs — operational logs benefit from severity levels; connection records need completeness without level noise.

**Fix:** Split into two loggers:

- **System log** — startup, config, errors, shutdown. Uses standard zerolog levels. Configured via existing `directory` and `log_file_prefix` keys.
- **Access log** — client connection events (NEW, FINISH, IGNORED_UA, DENIED_METHOD, PROXY_ABUSE_ATTEMPT, SERVE_ROBOTS). No level field — all events use `.Log()`. Configured via new `access_directory` and `access_prefix` keys under `[logger]`.

Security-relevant events (DENIED_METHOD, PROXY_ABUSE_ATTEMPT, SERVE_ROBOTS_ERROR) are dual-logged: the access log gets the connection record without level, the system log gets the operational warning/error at the appropriate severity.

New config keys:
```toml
[logger]
access_directory = ""       # defaults to system log directory if empty
access_prefix = "access"    # produces access.log or access_<datestamp>.log
```

### 3. Pool Init Ordering (`heffalump/heffalump.go`, `cmd/HellPot/HellPot.go`)

**Problem:** `heffalump.init()` called `NewHeffalump()` which checked `config.ChunkPoolSizeMB` to decide whether to create the chunk pool. But Go initializes imported packages before the main package, so `config.Init()` (which loads the TOML and sets `ChunkPoolSizeMB`) hadn't run yet. The value was always 0 at that point — the pool was silently never created regardless of config. Same issue affected the global rate limiter (`MaxTotalKbps`). The package-level `var log = config.GetLogger()` also captured the pre-init stderr fallback instead of the real logger.

**Fix:** Split `NewHeffalump()` into two phases:
- `NewHeffalump()` — called during package init. Creates only the base struct (Markov map, sync.Pools). Reads no config values.
- `InitFromConfig()` — new exported function called explicitly from `main.init()` after `config.Init()` and `StartLogger()`. Creates the chunk pool and global rate limiter using actual config values. Re-fetches the logger.

### 4. Pool Generate Infinite Loop (`heffalump/pool.go`)

**Problem:** `generate()` filled each chunk by calling `mr.Read(buf[total:])` in a loop, breaking on `err != nil || total >= ChunkSize`. When the buffer was nearly full (e.g., 2 bytes remaining), `Read()` couldn't fit any word + space separator, returned `n = 0`, but `generate()` didn't check for that — it looped back and called `Read()` again with the same result, forever. The Markov chain entry point `("", "")` has exactly one successor (the first word of the corpus), and the `w1/w2` state never advanced because the copy was never reached. Every chunk hit this infinite busy loop on its last few bytes, pinning a CPU core at 100%.

This bug was latent since PR #64 — unreachable until the init ordering fix above caused the pool to actually be created for the first time.

**Fix:** Added `n == 0` to the break condition. A chunk a few bytes short of perfectly full is functionally identical for bot consumption.

## Files Changed

| File | Change |
|------|--------|
| `heffalump/heffalump.go` | `writeSliced()` helper, HTML prefix flush, `InitFromConfig()`, `NewHeffalump()` stripped of config reads |
| `heffalump/pool.go` | `n == 0` break in `generate()` |
| `internal/config/logger.go` | `StartAccessLogger()`, `GetAccessLogger()`, `buildLogFileName()` shared helper |
| `internal/config/globals.go` | `AccessLogDirectory`, `AccessLogPrefix` |
| `internal/config/config.go` | `access_directory` and `access_prefix` in `processOpts()` |
| `internal/config/default_config.toml` | New access log config keys with comments |
| `internal/http/router.go` | `alog` access logger, connection events use `.Log()`, security events dual-logged |
| `internal/http/robots.go` | Uses access logger, SERVE_ROBOTS_ERROR dual-logged |
| `cmd/HellPot/HellPot.go` | Starts access logger, calls `heffalump.InitFromConfig()` |
| `docker_config.toml` | Added `access_directory` and `access_prefix` |

## Testing

- `baseline_rate_kbps = 2` with `pool_size_mb = 128`: data streams immediately in smooth 4KB chunks, curl disconnect detected within 2 seconds
- POST request produces DENIED_METHOD in both system log (with `"level":"warn"`) and access log (without level field)
- GET connection produces NEW/FINISH in access log only — system log stays quiet during normal trapped-bot operation
- Pool startup completes in seconds (previously infinite loop at 100% CPU)
- Pool refill goroutine visible as slight periodic CPU activity (~1 chunk/sec at default 128 KB/s refill rate)
- Start/stop cycling confirms both logs pick up correctly
