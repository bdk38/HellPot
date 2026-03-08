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

	if customconfig {
		associateExportedVariables()
		return
	}

	setDefaults()

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
	for _, p := range []string{"./config.toml", filepath.Join(pwd, "config.toml")} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
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
		"logger.directory":           &logDir,
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
		"performance.restrict_concurrency": &RestrictConcurrency,
		"http.use_unix_socket":             &UseUnixSocket,
		"logger.debug":                     &Debug,
		"logger.trace":                     &Trace,
		"logger.nocolor":                   &NoColor,
		"logger.docker_logging":            &DockerLogging,
		"http.router.makerobots":           &MakeRobots,
		"http.router.catchall":             &CatchAll,
	}
	// integer options and their exported variables
	intOpt := map[string]*int{
		"performance.max_workers": &MaxWorkers,
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
	// Load environment variables (HELLPOT_*)
	_ = snek.Load(env.Provider(".", env.Opt{
		Prefix: "HELLPOT_",
		TransformFunc: func(s, v string) (string, any) {
			s = strings.TrimPrefix(s, "HELLPOT_")
			s = strings.ToLower(s)
			s = strings.ReplaceAll(s, "__", " ")
			s = strings.ReplaceAll(s, "_", ".")
			s = strings.ReplaceAll(s, " ", "_")
			return s, v
		},
	}), nil)
	processOpts()

	if noColorForce {
		NoColor = true
	}

	if UseUnixSocket {
		UnixSocketPath = snek.String("http.unix_socket_path")
		parsedPermissions, err := strconv.ParseUint(snek.String("http.unix_socket_permissions"), 8, 32)
		if err == nil {
			UnixSocketPermissions = uint32(parsedPermissions)
		}
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
}
