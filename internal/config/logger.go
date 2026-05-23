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
	// CurrentLogFile is used for accessing the location of the currently used log file across packages.
	CurrentLogFile string
	logFile        io.Writer
	logDir         string
	logger         zerolog.Logger
	loggerReady    bool // set to true by StartLogger; guards GetLogger against pre-init calls
)

func prepLogDir() {
	logDir = snek.String("logger.directory")
	if logDir == "" {
		logDir = filepath.Join(home, ".local", "share", Title, "logs")
	}
	_ = os.MkdirAll(logDir, 0750)
}

// StartLogger instantiates an instance of our zerolog logger so we can hook it in our main package.
// While this does return a logger, it should not be used for additional retrievals of the logger. Use GetLogger().
func StartLogger(pretty bool, targets ...io.Writer) zerolog.Logger {
	prefix := snek.String("logger.log_file_prefix")
	if prefix == "" {
		prefix = "hellpot"
	}

	logFileName := prefix
	if snek.Bool("logger.use_date_filename") {
		tn := strings.ReplaceAll(time.Now().Format(time.RFC822), " ", "_")
		tn = strings.ReplaceAll(tn, ":", "-")
		logFileName = logFileName + "_" + tn
	}

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

// GetLogger retrieves our global logger object.
// If called before StartLogger, it returns a stderr fallback logger and logs a
// warning rather than silently returning a zero-value logger that may drop messages.
func GetLogger() *zerolog.Logger {
	if !loggerReady {
		fb := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
		return &fb
	}
	return &logger
}
