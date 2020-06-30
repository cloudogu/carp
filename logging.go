package carp

import (
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"os"
)

var log = logging.MustGetLogger("carp")

func prepareLogger(configuration Configuration) error {
	backend := logging.NewLogBackend(os.Stderr, "", 0)

	var format = logging.MustStringFormatter(configuration.LoggingFormat)
	formatter := logging.NewBackendFormatter(backend, format)

	level, err := convertLogLevel(configuration.LogLevel)
	if err != nil {
		return errors.Wrap(err, "could not prepare logger")
	}
	backendLeveled := logging.AddModuleLevel(formatter)
	backendLeveled.SetLevel(level, "")

	logging.SetBackend(backendLeveled)

	return nil
}

func convertLogLevel(logLevel string) (logging.Level, error) {
	switch logLevel {
	case "WARN":
		return logging.LogLevel("WARNING")
	default:
		return logging.LogLevel(logLevel)
	}
}
