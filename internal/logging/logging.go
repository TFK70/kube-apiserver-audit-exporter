package logging

import (
	"fmt"
	"io"

	"github.com/sirupsen/logrus"
)

var (
	logger      *logrus.Logger
	IsNullified = false
)

func SetupLogger() *logrus.Logger {
	if logger != nil {
		return logger
	}

	logger = logrus.New()

	return logger
}

func NullifyLogger() error {
	if logger == nil {
		return fmt.Errorf("No logger to nullify")
	}

	logger.SetOutput(io.Discard)
	IsNullified = true

	return nil
}

func GetLogger() (*logrus.Logger, error) {
	if logger == nil {
		return nil, fmt.Errorf("Logger not found")
	}

	return logger, nil
}

func GetNamedLogger(name string) (*logrus.Entry, error) {
	if name == "" {
		return nil, fmt.Errorf("Name for logger was not provided")
	}

	if logger == nil {
		return nil, fmt.Errorf("Logger not found")
	}

	namedLogger := logger.WithField("logger", name)

	return namedLogger, nil
}
