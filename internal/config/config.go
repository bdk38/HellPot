package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/rs/zerolog"

	"github.com/knadh/koanf/parsers/toml/v2"
	koanf "github.com/knadh/koanf/v2" // ← clean alias (no more confusing "viper")
)

// generic vars
var (
	noColorForce = false
	customconfig = false
	home         string
	snek         = koanf.New(".") // keep snek. No step on snek.
)

func init() {
	home, _ = os.UserHomeDir()
	if home == "" {
		home = os.Getenv("HOME")
	}
	if home == "" {
		println("WARNING: could not determine home directory")
	}
}

// exported generic vars
var (
	// Trace is the value of our trace (extra verbose) on/off toggle as per the current configuration.
	Trace bool
	// Debug is the value of our debug (verbose) on/off toggle as per the current configuration.
	Debug bool
	// Filename returns the current location of our toml config file.
	Filename string
)

// Init will initialize our toml configuration engine and define our default configuration values.
func Init() {
	argParse()

	// GenConfig is set by --genconfig. gen() writes the embedded default_config.toml
	// to disk and exits. It must be checked before we attempt to find or load a config
	// file, since the intent is to produce one rather than consume one.
	if GenConfig {
		gen("")
	}

	if customconfig {
		associateExportedVariables()
		return
	}

	chosen := findConfigPath()
	if chosen == "" {
		printConfigErrorAndExit()
		return
	}

	// Load the config file (only once)
	if err := snek.Load(file.Provider(chosen), toml.Parser()); err != nil {
		fmt.Printf("failed to load config file %s: %v\n", chosen, err)
		os.Exit(1)
	}

	Filename = chosen
	associateExportedVariables()
}

// findConfigPath returns the first existing config.toml in our standard search order.
// Returns "" if none is found.
func findConfigPath() string {
	// 1. System-wide config (Linux/macOS only)
	switch runtime.GOOS {
	case "windows":
		// Windows support can be added here later if needed
	default:
		etcPath := filepath.Join("/etc/", Title, "config.toml")
		if _, err := os.Stat(etcPath); err == nil {
			return etcPath
		}
	}

	// 2. User config directory (~/.config/HellPot/config.toml)
	uconf, _ := os.UserConfigDir()
	if uconf == "" && home != "" {
		uconf = filepath.Join(home, ".config")
	}
	if uconf != "" {
		userPath := filepath.Join(uconf, Title, "config.toml")
		_ = os.MkdirAll(filepath.Join(uconf, Title), 0750)
		if _, err := os.Stat(userPath); err == nil {
			return userPath
		}
	}

	// 3. Current working directory
	pwd, _ := os.Getwd()
	if _, err := os.Stat(filepath.Join(pwd, "config.toml")); err == nil {
		return filepath.Join(pwd, "config.toml")
	}

	return ""
}

func printConfigErrorAndExit() {
	println("ERROR: No configuration file found.")
	println("")
	println("Generate a default config with:")
	println("  ./HellPot --genconfig")
	println("")
	println("Or specify a config file:")
	println("  ./HellPot --config /path/to/config.toml")
	println("")
	println("See README.md for detailed setup instructions.")
	os.Exit(1)
}

func loadCustomConfig(path string) {
	Filename, _ = filepath.Abs(path)

	if err := snek.Load(file.Provider(Filename), toml.Parser()); err != nil {
		fmt.Println("failed to load specified config file: ", err.Error())
		os.Exit(1)
	}

	customconfig = true
}

func processOpts() {
	// string options and their exported variables
	stringOpt := map[string]*string{
		"http.bind_addr":             &HTTPBind,
		"http.bind_port":             &HTTPPort,
		"http.real_ip_header":        &HeaderName,
		"logger.console_time_format": &ConsoleTimeFormat,
		"deception.server_name":      &FakeServerName,
	}
	// string slice options and their exported variables
	strSliceOpt := map[string]*[]string{
		"http.router.paths":            &Paths,
		"http.uagent_string_blacklist": &UseragentBlacklistMatchers,
	}
	// bool options and their exported variables
	boolOpt := map[string]*bool{
		"http.use_unix_socket":  &UseUnixSocket,
		"logger.debug":         &Debug,
		"logger.trace":         &Trace,
		"logger.nocolor":       &NoColor,
		"logger.docker_logging": &DockerLogging,
		"http.router.makerobots": &MakeRobots,
		"http.router.catchall":   &CatchAll,
	}
	// integer options and their exported variables
	intOpt := map[string]*int{
		"performance.max_workers":              &MaxWorkers,
		"performance.baseline_rate_kbps":       &BaselineRateKbps,
		"performance.max_total_kbps":           &MaxTotalKbps,
		"performance.chunks.pool_size_mb":      &ChunkPoolSizeMB,
		"performance.chunks.chunk_size_kb":     &ChunkSizeKB,
		"performance.chunks.refill_rate_kbps":  &ChunkRefillRateKbps,
	}

	for key, opt := range stringOpt {
		*opt = snek.String(key)
	}
	for key, opt := range strSliceOpt {
		*opt = snek.Strings(key)
	}
	for key, opt := range boolOpt {
		*opt = snek.Bool(key)
	}
	for key, opt := range intOpt {
		*opt = snek.Int(key)
	}
}

func associateExportedVariables() {
	// Load environment variables with prefix HELLPOT_.
	// Key mapping convention: a single underscore becomes a section separator (.)
	// and a double underscore becomes a literal underscore within a key name.
	// Examples:
	//   HELLPOT_HTTP_BIND_PORT        → http.bind_port
	//   HELLPOT_HTTP_BIND__ADDR       → http.bind_addr   (__ preserves _ in bind_addr)
	//   HELLPOT_HTTP_REAL__IP__HEADER → http.real_ip_header
	_ = snek.Load(env.Provider(".", env.Opt{
		Prefix: "HELLPOT_",
		TransformFunc: func(s, v string) (string, any) {
			s = strings.TrimPrefix(s, "HELLPOT_")
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, "__", " ") // protect literal underscores
			s = strings.ReplaceAll(s, "_", ".")  // single _ → section separator
			s = strings.ReplaceAll(s, " ", "_")  // restore protected underscores
			return s, v
		},
	}), nil)
	processOpts()

	if noColorForce {
		NoColor = true
	}

	// Always load the unix socket settings regardless of UseUnixSocket so that
	// a config-file change from false→true is honoured on restart without also
	// requiring an env-var re-export, and so that validation code can inspect
	// the path even before deciding whether sockets are in use.
	UnixSocketPath = snek.String("http.unix_socket_path")
	if perm, err := strconv.ParseUint(snek.String("http.unix_socket_permissions"), 8, 32); err == nil {
		UnixSocketPermissions = uint32(perm)
	}

	// === CLI flag overrides (flags always win over config + env) ===
	// We do this last so --debug / --trace take precedence.
	if forceTrace || Trace {
		Trace = true
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else if forceDebug || Debug {
		Debug = true
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// neither enabled → force INFO level (this was missing)
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	validatePerformanceConfig()
}

// validatePerformanceConfig checks the [performance] and [performance.chunks]
// settings for invalid or contradictory values. Every problem is corrected to a
// safe default and logged as a warning to stderr (the logger is not yet
// initialised when this runs). The process never exits — a misconfigured tarpit
// that falls back to sane defaults is far better than one that refuses to start.
func validatePerformanceConfig() {
	const (
		defaultChunkSizeSmallKB  = 64
		defaultChunkSizeMediumKB = 128
		defaultChunkSizeLargeKB  = 256
		minTotalKbps             = 64
		maxChunkSizeKB           = 1024
	)

	warn := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, "[WARN] performance config: "+format+"\n", args...)
	}
	info := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, "[INFO] performance config: "+format+"\n", args...)
	}

	// ── max_total_kbps ───────────────────────────────────────────────────────
	if MaxTotalKbps > 0 && MaxTotalKbps < minTotalKbps {
		warn("max_total_kbps (%d) is implausibly low (min %d) — reverting to 2048",
			MaxTotalKbps, minTotalKbps)
		MaxTotalKbps = 2048
	}

	// ── baseline_rate_kbps vs max_total_kbps ─────────────────────────────────
	if MaxTotalKbps > 0 && BaselineRateKbps > MaxTotalKbps {
		warn("baseline_rate_kbps (%d) exceeds max_total_kbps (%d) — clamping baseline to max_total",
			BaselineRateKbps, MaxTotalKbps)
		BaselineRateKbps = MaxTotalKbps
	}

	// max_workers × baseline informational — total cap will govern
	if MaxTotalKbps > 0 && BaselineRateKbps > 0 && MaxWorkers > 0 {
		if projected := MaxWorkers * BaselineRateKbps; projected > MaxTotalKbps {
			info("max_workers (%d) × baseline_rate_kbps (%d) = %d KB/s exceeds max_total_kbps (%d) — max_total_kbps will be the binding constraint",
				MaxWorkers, BaselineRateKbps, projected, MaxTotalKbps)
		}
	}

	// ── chunk pool ───────────────────────────────────────────────────────────
	if ChunkPoolSizeMB <= 0 {
		// Pool disabled — skip chunk validation entirely.
		return
	}

	// Derive chunk_size_kb if not explicitly set (0 = not set)
	if ChunkSizeKB <= 0 {
		switch {
		case ChunkPoolSizeMB <= 32:
			ChunkSizeKB = defaultChunkSizeSmallKB
		case ChunkPoolSizeMB <= 128:
			ChunkSizeKB = defaultChunkSizeMediumKB
		default:
			ChunkSizeKB = defaultChunkSizeLargeKB
		}
	}

	// chunk larger than the whole pool
	if ChunkSizeKB*1024 > ChunkPoolSizeMB*1024*1024 {
		warn("chunk_size_kb (%d KB) exceeds pool_size_mb (%d MB) — reverting chunk_size_kb to %d",
			ChunkSizeKB, ChunkPoolSizeMB, defaultChunkSizeSmallKB)
		ChunkSizeKB = defaultChunkSizeSmallKB
	}

	// unreasonably large chunk
	if ChunkSizeKB > maxChunkSizeKB {
		warn("chunk_size_kb (%d) exceeds maximum (%d KB) — reverting to %d",
			ChunkSizeKB, maxChunkSizeKB, defaultChunkSizeSmallKB)
		ChunkSizeKB = defaultChunkSizeSmallKB
	}

	// Derive refill_rate_kbps if not set
	if ChunkRefillRateKbps <= 0 {
		if MaxTotalKbps > 0 {
			ChunkRefillRateKbps = MaxTotalKbps / 10
		}
		if ChunkRefillRateKbps < 128 {
			ChunkRefillRateKbps = 128
		}
		if ChunkRefillRateKbps > 4096 {
			ChunkRefillRateKbps = 4096
		}
	}

	// refill faster than serve rate — wasteful but not harmful
	if MaxTotalKbps > 0 && ChunkRefillRateKbps > MaxTotalKbps {
		info("refill_rate_kbps (%d) exceeds max_total_kbps (%d) — refill is faster than serve rate (wasteful but not harmful)",
			ChunkRefillRateKbps, MaxTotalKbps)
	}

	// unlimited baseline with pool enabled — informational
	if BaselineRateKbps == 0 {
		info("baseline_rate_kbps is 0 (unlimited) with chunk pool enabled — connections will serve at full speed; consider setting a baseline rate to protect host resources")
	}
}
