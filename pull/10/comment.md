This PR represents an important architectural shift in how we handle default configuration generation.

**The Problem We Were Fighting:**
We spent PRs #8 and #9 trying to make go-toml/v2's marshaler produce proper TOML sections from nested Go maps. Despite multiple approaches (type conversions, different map structures), the marshaler consistently produced quoted dotted keys like `'http.bind_addr' = '...'` instead of proper `[http]` sections.

**The Mindset Shift:**
Instead of continuing to fight the marshaler, we asked: 'Why are we generating TOML from Go code at all?' The answer led us to a better architectural pattern:

**Static Configuration as Source of Truth:**
- The TOML file IS the configuration, not a representation of Go data structures
- We embed it at compile time using `//go:embed`, making it part of the binary
- Config generation becomes trivial: just write the embedded bytes
- Future config changes are made in TOML (the native format), not Go code

**Benefits:**
1. **Correctness by construction** - Hand-crafted TOML is always valid
2. **Version control** - Config changes are visible in git diffs of actual TOML
3. **Maintainability** - Edit config.toml, rebuild, done. No marshaling code to debug
4. **Simplicity** - Removes ~15 lines of complex marshaling logic
5. **Separation of concerns** - Config format lives in config files, not Go structs

This is a pattern worth remembering: sometimes the right solution is to stop fighting a tool and rethink the approach entirely. The embedded static file pattern is simpler and more maintainable.
