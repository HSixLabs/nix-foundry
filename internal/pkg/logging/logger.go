package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger *zap.SugaredLogger

// Logger wraps zap.SugaredLogger to provide application-specific logging
type Logger struct {
	*zap.SugaredLogger
}

// InitLogger initializes the global logger with the specified log level
func InitLogger(level string) error {
	config := zap.NewProductionConfig()

	// Parse log level
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return err
	}
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	// Enable development mode for debug level
	if level == "debug" {
		config = zap.NewDevelopmentConfig()
	}

	// Configure output paths
	config.OutputPaths = []string{"stdout", "nix-foundry.log"}

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
