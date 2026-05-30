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
	// [http]
	HTTP.BindAddr = snek.String("http.bind_addr")
	HTTP.BindPort = snek.String("http.bind_port")
	HTTP.RealIPHeader = snek.String("http.real_ip_header")
	HTTP.UseUnixSocket = snek.Bool("http.use_unix_socket")
	HTTP.UABlacklist = snek.Strings("http.uagent_string_blacklist")

	// [http.router]
	HTTP.Router.CatchAll = snek.Bool("http.router.catchall")
	HTTP.Router.MakeRobots = snek.Bool("http.router.makerobots")
	HTTP.Router.Paths = snek.Strings("http.router.paths")

	// [performance]
	Perf.MaxWorkers = snek.Int("performance.max_workers")
	Perf.BaselineRateKbps = snek.Int("performance.baseline_rate_kbps")
	Perf.MaxTotalKbps = snek.Int("performance.max_total_kbps")

	// [performance.chunks]
	Perf.Chunks.PoolSizeMB = snek.Int("performance.chunks.pool_size_mb")
	Perf.Chunks.ChunkSizeKB = snek.Int("performance.chunks.chunk_size_kb")
	Perf.Chunks.RefillRateKbps = snek.Int("performance.chunks.refill_rate_kbps")

	// [logger]
	Logger.Debug = snek.Bool("logger.debug")
	Logger.Trace = snek.Bool("logger.trace")
	Logger.NoColor = snek.Bool("logger.nocolor")
	Logger.DockerLogging = snek.Bool("logger.docker_logging")
	Logger.UseDateFilename = snek.Bool("logger.use_date_filename")
	Logger.Directory = snek.String("logger.directory")
	Logger.LogFilePrefix = snek.String("logger.log_file_prefix")
	Logger.AccessDirectory = snek.String("logger.access_directory")
	Logger.AccessPrefix = snek.String("logger.access_prefix")
	Logger.ConsoleTimeFormat = snek.String("logger.console_time_format")

	// [deception]
	Deception.ServerName = snek.String("deception.server_name")
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
		Logger.NoColor = true
	}

	// Always load the unix socket settings regardless of UseUnixSocket so that
	// a config-file change from false→true is honoured on restart without also
	// requiring an env-var re-export, and so that validation code can inspect
	// the path even before deciding whether sockets are in use.
	HTTP.UnixSocketPath = snek.String("http.unix_socket_path")
	if perm, err := strconv.ParseUint(snek.String("http.unix_socket_permissions"), 8, 32); err == nil {
		HTTP.UnixSocketPermissions = uint32(perm)
	}

	// === CLI flag overrides (flags always win over config + env) ===
	// We do this last so --debug / --trace take precedence.
	if forceTrace || Logger.Trace {
		Logger.Trace = true
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	} else if forceDebug || Logger.Debug {
		Logger.Debug = true
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
	if Perf.MaxTotalKbps > 0 && Perf.MaxTotalKbps < minTotalKbps {
		warn("max_total_kbps (%d) is implausibly low (min %d) — reverting to 2048",
			Perf.MaxTotalKbps, minTotalKbps)
		Perf.MaxTotalKbps = 2048
	}

	// ── baseline_rate_kbps vs max_total_kbps ─────────────────────────────────
	if Perf.MaxTotalKbps > 0 && Perf.BaselineRateKbps > Perf.MaxTotalKbps {
		warn("baseline_rate_kbps (%d) exceeds max_total_kbps (%d) — clamping baseline to max_total",
			Perf.BaselineRateKbps, Perf.MaxTotalKbps)
		Perf.BaselineRateKbps = Perf.MaxTotalKbps
	}

	// max_workers × baseline informational — total cap will govern
	if Perf.MaxTotalKbps > 0 && Perf.BaselineRateKbps > 0 && Perf.MaxWorkers > 0 {
		if projected := Perf.MaxWorkers * Perf.BaselineRateKbps; projected > Perf.MaxTotalKbps {
			info("max_workers (%d) × baseline_rate_kbps (%d) = %d KB/s exceeds max_total_kbps (%d) — max_total_kbps will be the binding constraint",
				Perf.MaxWorkers, Perf.BaselineRateKbps, projected, Perf.MaxTotalKbps)
		}
	}

	// ── chunk pool ───────────────────────────────────────────────────────────
	if Perf.Chunks.PoolSizeMB <= 0 {
		// Pool disabled — skip chunk validation entirely.
		return
	}

	// Derive chunk_size_kb if not explicitly set (0 = not set)
	if Perf.Chunks.ChunkSizeKB <= 0 {
		switch {
		case Perf.Chunks.PoolSizeMB <= 32:
			Perf.Chunks.ChunkSizeKB = defaultChunkSizeSmallKB
		case Perf.Chunks.PoolSizeMB <= 128:
			Perf.Chunks.ChunkSizeKB = defaultChunkSizeMediumKB
		default:
			Perf.Chunks.ChunkSizeKB = defaultChunkSizeLargeKB
		}
	}

	// chunk larger than the whole pool
	if Perf.Chunks.ChunkSizeKB*1024 > Perf.Chunks.PoolSizeMB*1024*1024 {
		warn("chunk_size_kb (%d KB) exceeds pool_size_mb (%d MB) — reverting chunk_size_kb to %d",
			Perf.Chunks.ChunkSizeKB, Perf.Chunks.PoolSizeMB, defaultChunkSizeSmallKB)
		Perf.Chunks.ChunkSizeKB = defaultChunkSizeSmallKB
	}

	// unreasonably large chunk
	if Perf.Chunks.ChunkSizeKB > maxChunkSizeKB {
		warn("chunk_size_kb (%d) exceeds maximum (%d KB) — reverting to %d",
			Perf.Chunks.ChunkSizeKB, maxChunkSizeKB, defaultChunkSizeSmallKB)
		Perf.Chunks.ChunkSizeKB = defaultChunkSizeSmallKB
	}

	// Derive refill_rate_kbps if not set
	if Perf.Chunks.RefillRateKbps <= 0 {
		if Perf.MaxTotalKbps > 0 {
			Perf.Chunks.RefillRateKbps = Perf.MaxTotalKbps / 10
		}
		if Perf.Chunks.RefillRateKbps < 128 {
			Perf.Chunks.RefillRateKbps = 128
		}
		if Perf.Chunks.RefillRateKbps > 4096 {
			Perf.Chunks.RefillRateKbps = 4096
		}
	}

	// refill faster than serve rate — wasteful but not harmful
	if Perf.MaxTotalKbps > 0 && Perf.Chunks.RefillRateKbps > Perf.MaxTotalKbps {
		info("refill_rate_kbps (%d) exceeds max_total_kbps (%d) — refill is faster than serve rate (wasteful but not harmful)",
			Perf.Chunks.RefillRateKbps, Perf.MaxTotalKbps)
	}

	// unlimited baseline with pool enabled — informational
	if Perf.BaselineRateKbps == 0 {
		info("baseline_rate_kbps is 0 (unlimited) with chunk pool enabled — connections will serve at full speed; consider setting a baseline rate to protect host resources")
	}
}
