package config

// Title is the name of the application used throughout the configuration process.
const Title = "HellPot"

var Version = "0.7.0"

var (
	// BannerOnly when toggled causes HellPot to only print the banner and version then exit.
	BannerOnly = false
	// GenConfig when toggled causes HellPot to write its default config to the cwd and then exit.
	GenConfig = false
	// NoColor when true will disable the banner and any colored console output.
	NoColor bool
	// DockerLogging when true will disable the banner and any colored console output, as well as disable the log file.
	// Assumes NoColor == true.
	DockerLogging bool
	// MakeRobots when false will not respond to requests for robots.txt.
	MakeRobots bool
	// CatchAll when true will cause HellPot to respond to all paths.
	// Note that this will override MakeRobots.
	CatchAll bool
	// ConsoleTimeFormat sets the time format for the console. The string is passed to time.Format() down the line.
	ConsoleTimeFormat string
)

// "http"
var (
	// HTTPBind is defined via our toml configuration file. It is the address that HellPot listens on.
	HTTPBind string
	// HTTPPort is defined via our toml configuration file. It is the port that HellPot listens on.
	HTTPPort string
	// HeaderName is defined via our toml configuration file. It is the HTTP Header containing the original IP of the client,
	// in traditional reverse Proxy deployments.
	HeaderName string

	// Paths are defined via our toml configuration file. These are the paths that HellPot will present for "robots.txt"
	//       These are also the paths that HellPot will respond for. Other paths will throw a warning and will serve a 404.
	Paths []string

	// UseUnixSocket determines if we will listen for HTTP connections on a unix socket.
	UseUnixSocket bool

	// UnixSocketPath is defined via our toml configuration file. It is the path of the socket HellPot listens on
	// if UseUnixSocket, also defined via our toml configuration file, is set to true.
	UnixSocketPath        = ""
	UnixSocketPermissions uint32

	// UseragentBlacklistMatchers contains useragent matches checked for with strings.Contains() that
	// prevent HellPot from firing off.
	// See: https://github.com/bdk38/HellPot/issues/23
	UseragentBlacklistMatchers []string
)

// "performance"
var (
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

	// ChunkPoolSizeMB is the total RAM budget for the pre-generated Markov chunk pool,
	// in MB. When set, HellPot pre-generates this much Markov text at startup and serves
	// connections from memory (memcpy) instead of generating on the fly. This dramatically
	// reduces per-connection CPU cost — recommended for ARM and other constrained hardware.
	// Set to 0 to disable the pool and use the original on-the-fly generation behavior.
	// 16MB is comfortable on a router; 64–128MB suits a server.
	ChunkPoolSizeMB int

	// ChunkSizeKB is the size of each pre-generated chunk in KB.
	// Derived automatically from ChunkPoolSizeMB if not set:
	//   ≤32MB pool  → 64KB chunks
	//   ≤128MB pool → 128KB chunks
	//   >128MB pool → 256KB chunks
	ChunkSizeKB int

	// ChunkRefillRateKbps is the rate at which the background goroutine regenerates
	// consumed chunks, in KB/s. Derived as 10% of MaxTotalKbps (floor 128, ceil 4096)
	// if not set. Lower values use less CPU for background regeneration.
	ChunkRefillRateKbps int
)

// "logger" — access log
var (
	// AccessLogDirectory is the directory for the access log (client connection events).
	// Defaults to the same directory as the system log if empty.
	AccessLogDirectory string

	// AccessLogPrefix is the filename prefix for the access log.
	// A datestamp is appended when use_date_filename is true.
	AccessLogPrefix string
)

// "deception"
var (
	// FakeServerName is our configured value for the "Server: " response header when serving HTTP clients
	FakeServerName string
)
