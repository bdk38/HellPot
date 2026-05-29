package config

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var (
	// CurrentLogFile is used for accessing the location of the currently used system log file across packages.
	CurrentLogFile string
	logFile        io.Writer
	logDir         string
	logger         zerolog.Logger
	loggerReady    bool // set to true by StartLogger; guards GetLogger against pre-init calls

	// CurrentAccessLogFile is the location of the currently used access log file.
	CurrentAccessLogFile string
	accessLogFile        io.Writer
	accessLogDir         string
	accessLogger         zerolog.Logger
	accessLoggerReady    bool // set to true by StartAccessLogger
)

func prepLogDir() {
	logDir = snek.String("logger.directory")
	if logDir == "" {
		logDir = filepath.Join(home, ".local", "share", Title, "logs")
	}
	_ = os.MkdirAll(logDir, 0750)
}

// prepAccessLogDir resolves the access log directory from config.
// Falls back to the system log directory if not explicitly configured.
func prepAccessLogDir() {
	accessLogDir = AccessLogDirectory
	if accessLogDir == "" {
		// Default to the system log directory so a minimal config Just Works.
		prepLogDir()
		accessLogDir = logDir
	}
	_ = os.MkdirAll(accessLogDir, 0750)
}

// buildLogFileName constructs a log filename from a prefix, optionally
// appending a datestamp when use_date_filename is true.
func buildLogFileName(prefix string) string {
	name := prefix
	if snek.Bool("logger.use_date_filename") {
		tn := strings.ReplaceAll(time.Now().Format(time.RFC822), " ", "_")
		tn = strings.ReplaceAll(tn, ":", "-")
		name = name + "_" + tn
	}
	return name
}

// StartLogger instantiates the system logger (debug, info, warn, error, fatal).
// While this does return a logger, it should not be used for additional retrievals
// of the logger. Use GetLogger().
func StartLogger(pretty bool, targets ...io.Writer) zerolog.Logger {
	prefix := snek.String("logger.log_file_prefix")
	if prefix == "" {
		prefix = "hellpot"
	}

	logFileName := buildLogFileName(prefix)

	var err error

	switch {
	case len(targets) > 0:
		logFile = io.MultiWriter(targets...)
	default:
		prepLogDir()
		CurrentLogFile = path.Join(logDir, logFileName+".log")
		//nolint:lll
		logFile, err = os.OpenFile(CurrentLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666) // #nosec G304 G302 -- we are not using user input to create the file
		if err != nil {
			println("cannot create log file: " + err.Error())
			os.Exit(1)
		}
	}

	var logWriter = logFile

	if pretty {
		logWriter = zerolog.MultiLevelWriter(zerolog.ConsoleWriter{TimeFormat: ConsoleTimeFormat, NoColor: NoColor, Out: os.Stdout}, logFile)
	}

	logger = zerolog.New(logWriter).With().Timestamp().Logger()
	loggerReady = true
	return logger
}

// StartAccessLogger instantiates the access logger for client connection events.
// The access logger omits the level field — callers use .Log() instead of
// .Info()/.Debug()/etc. Console output suppresses the level column as well.
func StartAccessLogger(pretty bool, targets ...io.Writer) zerolog.Logger {
	prefix := AccessLogPrefix
	if prefix == "" {
		prefix = "access"
	}

	accessLogFileName := buildLogFileName(prefix)

	var err error

	switch {
	case len(targets) > 0:
		accessLogFile = io.MultiWriter(targets...)
	default:
		prepAccessLogDir()
		CurrentAccessLogFile = path.Join(accessLogDir, accessLogFileName+".log")
		//nolint:lll
		accessLogFile, err = os.OpenFile(CurrentAccessLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666) // #nosec G304 G302
		if err != nil {
			println("cannot create access log file: " + err.Error())
			os.Exit(1)
		}
	}

	var logWriter = accessLogFile

	if pretty {
		consoleWriter := zerolog.ConsoleWriter{
			TimeFormat: ConsoleTimeFormat,
			NoColor:    NoColor,
			Out:        os.Stdout,
			// Suppress the level column — access log events use .Log() (NoLevel)
			// and the level field adds no value for connection records.
			FormatLevel: func(_ interface{}) string { return "" },
		}
		logWriter = zerolog.MultiLevelWriter(consoleWriter, accessLogFile)
	}

	accessLogger = zerolog.New(logWriter).With().Timestamp().Logger()
	accessLoggerReady = true
	return accessLogger
}

// GetLogger retrieves the global system logger.
// If called before StartLogger, it returns a stderr fallback logger and logs a
// warning rather than silently returning a zero-value logger that may drop messages.
func GetLogger() *zerolog.Logger {
	if !loggerReady {
		fb := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
		return &fb
	}
	return &logger
}

// GetAccessLogger retrieves the global access logger for client connection events.
// If called before StartAccessLogger, it returns a stderr fallback logger.
func GetAccessLogger() *zerolog.Logger {
	if !accessLoggerReady {
		fb := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
		return &fb
	}
	return &accessLogger
}
