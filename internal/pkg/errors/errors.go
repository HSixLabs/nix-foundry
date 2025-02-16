package errors

import (
	"fmt"
	"strings"
)

// Operation represents a specific operation being performed
type Operation string

// ErrorCode represents specific error conditions
type ErrorCode string

const (
	// Configuration errors
	ErrConfigInvalid    ErrorCode = "CONFIG_INVALID"
	ErrConfigNotFound   ErrorCode = "CONFIG_NOT_FOUND"
	ErrConfigValidation ErrorCode = "CONFIG_VALIDATION"

	// Environment errors
	ErrEnvSwitch     ErrorCode = "ENV_SWITCH"
	ErrEnvPermission ErrorCode = "ENV_PERMISSION"
	ErrEnvIsolation  ErrorCode = "ENV_ISOLATION"

	// System errors
	ErrSystemRequirement ErrorCode = "SYSTEM_REQUIREMENT"
	ErrPlatformSupport   ErrorCode = "PLATFORM_SUPPORT"

	// Backup errors
	ErrBackupCreate     ErrorCode = "BACKUP_CREATE"
	ErrBackupPermission ErrorCode = "BACKUP_PERMISSION"
)

// FoundryError represents a structured error in the application
type FoundryError struct {
	Op      Operation
	Code    ErrorCode
	Path    string
	Err     error
	Context map[string]interface{}
}

func (e *FoundryError) Error() string {
	var b strings.Builder

	if e.Op != "" {
		fmt.Fprintf(&b, "[%s] ", e.Op)
	}

	if e.Code != "" {
		fmt.Fprintf(&b, "%s: ", e.Code)
	}

	if e.Err != nil {
		b.WriteString(e.Err.Error())
	}

	if e.Path != "" {
		fmt.Fprintf(&b, " (path: %s)", e.Path)
	}

	if len(e.Context) > 0 {
		b.WriteString("\nContext:")
		for k, v := range e.Context {
			fmt.Fprintf(&b, "\n  %s: %v", k, v)
		}
	}

	return b.String()
}

// E creates a new FoundryError
func E(op Operation, code ErrorCode, err error) *FoundryError {
	return &FoundryError{
		Op:   op,
		Code: code,
		Err:  err,
	}
}

// WithPath adds a path to the error
func (e *FoundryError) WithPath(path string) *FoundryError {
	e.Path = path
	return e
}

// WithContext adds context to the error
func (e *FoundryError) WithContext(key string, value interface{}) *FoundryError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Implement error unwrapping
func (e *FoundryError) Unwrap() error {
	return e.Err
}

// Add proper error implementations
type OperationError struct {
	Operation string
	Err       error
	Context   map[string]interface{}
}

func (e *OperationError) Error() string {
	return fmt.Sprintf("%s operation failed: %v", e.Operation, e.Err)
}

func NewOperationError(op string, err error, ctx map[string]interface{}) error {
	return &OperationError{Operation: op, Err: err, Context: ctx}
}

type AlreadyExistsError struct {
	Path string
}

func (e *AlreadyExistsError) Error() string {
	return fmt.Sprintf("resource already exists: %s", e.Path)
}

func NewAlreadyExistsError(path string) error {
	return &AlreadyExistsError{Path: path}
}
