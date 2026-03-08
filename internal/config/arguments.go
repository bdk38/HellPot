package config

import (
	"flag"
)

// forceDebug and forceTrace are set by CLI flags and later override config values.
var (
	forceDebug = false
	forceTrace = false
)

// argParse now uses Go's standard flag package (replaces the old manual os.Args loop).
// This resolves the previous TODO.
func argParse() {
	var configFile string

	flag.BoolVar(&forceDebug, "debug", false, "enable debug logging")
	flag.BoolVar(&forceDebug, "v", false, "alias for --debug")

	flag.BoolVar(&forceTrace, "trace", false, "enable trace logging")
	flag.BoolVar(&forceTrace, "vv", false, "alias for --trace")

	flag.BoolVar(&noColorForce, "nocolor", false, "disable colored output")
	flag.BoolVar(&BannerOnly, "banner", false, "show banner only and exit")
	flag.BoolVar(&GenConfig, "genconfig", false, "generate default config and exit")

	flag.StringVar(&configFile, "config", "", "path to custom config file")
	flag.StringVar(&configFile, "c", "", "path to custom config file (shorthand)")

	// Safe wrapper for help function
	flag.Usage = func() {
		CLI.printUsage()
	}

	flag.Parse()

	// If a custom config was specified, load it immediately
	if configFile != "" {
		loadCustomConfig(configFile)
	}
}
