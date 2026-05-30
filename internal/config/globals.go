package config

// Title is the name of the application used throughout the configuration process.
const Title = "HellPot"

var Version = "0.7.0"

var (
	// BannerOnly when toggled causes HellPot to only print the banner and version then exit.
	BannerOnly = false
	// GenConfig when toggled causes HellPot to write its default config to the cwd and then exit.
	GenConfig = false
)

// HTTPConfig holds all settings from the [http] TOML section.
type HTTPConfig struct {
	// BindAddr is the address that HellPot listens on.
	BindAddr string
	// BindPort is the port that HellPot listens on.
	BindPort string
	// RealIPHeader is the HTTP Header containing the original client IP
	// in traditional reverse proxy deployments.
	RealIPHeader string

	// UseUnixSocket determines if we will listen for HTTP connections on a unix socket.
	UseUnixSocket bool
	// UnixSocketPath is the path of the socket HellPot listens on
	// if UseUnixSocket is set to true.
	UnixSocketPath        string
	UnixSocketPermissions uint32

	// UABlacklist contains useragent substrings checked with strings.Contains()
	// that prevent HellPot from firing off.
	UABlacklist []string

	Router RouterConfig
}

// RouterConfig holds settings from the [http.router] TOML section.
type RouterConfig struct {
	// CatchAll when true will cause HellPot to respond to all paths.
	// Note that this will override MakeRobots.
	CatchAll bool
	// MakeRobots when false will not respond to requests for robots.txt.
	MakeRobots bool
	// Paths are the paths that HellPot will present for "robots.txt"
	// and respond to. Other paths will serve a 404.
	Paths []string
}

// PerformanceConfig holds all settings from the [performance] TOML section.
type PerformanceConfig struct {
	// MaxWorkers is the maximum number of concurrent connections HellPot will handle.
	// Set to 0 to use the fasthttp default (262144). WARNING: setting this to 0 on a
	// low-resource server can exhaust memory and OOM the process — each trapped connection
	// holds a 256KB buffer for the duration of the stream. Size this to your available RAM.
	MaxWorkers int

	// BaselineRateKbps is the per-connection write rate cap in KB/s.
	// This is the primary CPU protection knob on constrained hardware — lower values
	// mean less Markov generation pressure per connection. 0 = unlimited.
	BaselineRateKbps int

	// MaxTotalKbps is the hard ceiling on total outbound bandwidth across all connections
	// combined, in KB/s. This is the primary bandwidth protection knob. 0 = unlimited.
	// WARNING: setting this to 0 on a metered or low-bandwidth host can exhaust your
	// outbound allowance — at full speed HellPot can push 10+ MB/s per connection.
	MaxTotalKbps int

	Chunks ChunkConfig
}

// ChunkConfig holds settings from the [performance.chunks] TOML section.
type ChunkConfig struct {
	// PoolSizeMB is the total RAM budget for the pre-generated Markov chunk pool,
	// in MB. When set, HellPot pre-generates this much Markov text at startup and serves
	// connections from memory (memcpy) instead of generating on the fly. This dramatically
	// reduces per-connection CPU cost — recommended for ARM and other constrained hardware.
	// Set to 0 to disable the pool and use the original on-the-fly generation behavior.
	// 16MB is comfortable on a router; 64–128MB suits a server.
	PoolSizeMB int

	// ChunkSizeKB is the size of each pre-generated chunk in KB.
	// Derived automatically from PoolSizeMB if not set.
	ChunkSizeKB int

	// RefillRateKbps is the rate at which the background goroutine regenerates
	// consumed chunks, in KB/s. Derived as 10% of MaxTotalKbps (floor 128, ceil 4096)
	// if not set. Lower values use less CPU for background regeneration.
	RefillRateKbps int
}

// LoggerConfig holds all settings from the [logger] TOML section.
type LoggerConfig struct {
	// Debug is the value of the debug (verbose) on/off toggle.
	Debug bool
	// Trace is the value of the trace (extra verbose) on/off toggle.
	Trace bool
	// NoColor when true will disable the banner and any colored console output.
	NoColor bool
	// DockerLogging when true will disable the banner and any colored console output,
	// as well as disable the log file. Assumes NoColor == true.
	DockerLogging bool
	// UseDateFilename when true appends a datestamp to log file names.
	UseDateFilename bool

	// Directory is the directory for the system log (startup, config, errors).
	// Leave empty to use the platform default.
	Directory string
	// LogFilePrefix is the prefix for the system log file name.
	// A datestamp is appended when UseDateFilename is true.
	LogFilePrefix string
	// AccessDirectory is the directory for the access log (client connection events).
	// Defaults to the same directory as the system log if empty.
	AccessDirectory string
	// AccessPrefix is the filename prefix for the access log.
	AccessPrefix string
	// ConsoleTimeFormat sets the time format for the console.
	// The string is passed to time.Format() down the line.
	ConsoleTimeFormat string
}

// DeceptionConfig holds all settings from the [deception] TOML section.
type DeceptionConfig struct {
	// ServerName is the configured value for the "Server: " response header.
	ServerName string
}

var (
	HTTP      HTTPConfig
	Perf      PerformanceConfig
	Logger    LoggerConfig
	Deception DeceptionConfig
)
