package logging

import (
	"go.uber.org/zap"
)

var defaultLogger *zap.SugaredLogger

// Logger wraps zap.SugaredLogger to provide application-specific logging
type Logger struct {
	*zap.SugaredLogger
}

// InitLogger initializes the global logger with the specified log level
func InitLogger(debug bool) error {
	var config zap.Config

	if debug {
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	}

	logger, err := config.Build()
	if err != nil {
		return err
	}

	defaultLogger = logger.Sugar()
	return nil
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		// Create a basic logger if not initialized
		logger, _ := zap.NewProduction()
		defaultLogger = logger.Sugar()
	}
	return &Logger{defaultLogger}
}

// WithError adds error context to the logger
func (l *Logger) WithError(err error) *Logger {
	return &Logger{l.SugaredLogger.With("error", err)}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{l.SugaredLogger.With(key, value)}
}

// Add component logger support to logging package
func NewComponentLogger(component string) *Logger {
	return GetLogger().WithField("component", component)
}

// Add WithComponent method to Logger
func (l *Logger) WithComponent(component string) *Logger {
	return l.WithField("component", component)
}
