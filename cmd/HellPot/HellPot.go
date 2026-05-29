package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"

	"github.com/bdk38/HellPot/heffalump"
	"github.com/bdk38/HellPot/internal/config"
	"github.com/bdk38/HellPot/internal/extra"
	"github.com/bdk38/HellPot/internal/http"
)

var (
	log     zerolog.Logger
	version string // set by linker
)

func init() {
	if version != "" {
		config.Version = version[1:]
	}
	config.Init()
	if config.BannerOnly {
		extra.Banner()
		os.Exit(0)
	}

	switch config.DockerLogging {
	case true:
		config.CurrentLogFile = "/dev/stdout"
		config.CurrentAccessLogFile = "/dev/stdout"
		config.NoColor = true
		log = config.StartLogger(false, os.Stdout)
		config.StartAccessLogger(false, os.Stdout)
	default:
		log = config.StartLogger(true)
		config.StartAccessLogger(true)
	}

	extra.Banner()

	log.Info().Str("caller", "config").Str("file", config.Filename).Msg(config.Filename)
	log.Info().Str("caller", "logger").Msg(config.CurrentLogFile)
	log.Info().Str("caller", "access_logger").Msg(config.CurrentAccessLogFile)
	log.Debug().Str("caller", "logger").Msg("debug enabled")
	log.Trace().Str("caller", "logger").Msg("trace enabled")

	// Initialise config-dependent components (chunk pool, rate limiter) now
	// that config is loaded and the logger is running. This cannot happen
	// during heffalump's package init because config values are zero at that
	// point — Go initialises imported packages before the main package.
	heffalump.InitFromConfig()

}

func main() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Fatal().Err(http.Serve()).Msg("HTTP error")
	}()

	<-stopChan // wait for SIGINT
	log.Warn().Msg("Shutting down server...")

}
