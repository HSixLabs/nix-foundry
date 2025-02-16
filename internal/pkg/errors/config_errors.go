package errors

import "fmt"

// ConfigError represents configuration-related errors
type ConfigError struct {
	Op      string // Operation being performed
	Path    string // Path to the configuration file
	Err     error  // Original error
	Details string // Additional error details
}

func (e *ConfigError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s failed for %s: %v (%s)", e.Op, e.Path, e.Err, e.Details)
	}
	return fmt.Sprintf("%s failed: %v (%s)", e.Op, e.Err, e.Details)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}
