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

type BackupError struct {
	Path string
	Err  error
	Msg  string
}

func (e *BackupError) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Err)
}

func NewBackupError(path string, err error, msg string) error {
	return &BackupError{
		Path: path,
		Err:  err,
		Msg:  msg,
	}
}

func NewLoadError(path string, err error, details string) error {
	return &ConfigError{
		Op:      "Load",
		Path:    path,
		Err:     err,
		Details: details,
	}
}

// ValidationError represents a user-facing validation error
type ValidationError struct {
	Field       string
	Err         error
	UserMessage string
}

func (e *ValidationError) Error() string {
	return e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error with user message
func NewValidationError(field string, err error, userMsg string) *ValidationError {
	return &ValidationError{
		Field:       field,
		Err:         err,
		UserMessage: userMsg,
	}
}

// Add the missing NotFoundError type and constructor
type NotFoundError struct {
	Err     error
	Details string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %v (%s)", e.Err, e.Details)
}

func (e *NotFoundError) Unwrap() error {
	return e.Err
}

func NewNotFoundError(err error, details string) error {
	return &NotFoundError{
		Err:     err,
		Details: details,
	}
}

// Add ConflictError type
type ConflictError struct {
	Path        string
	Message     string
	Resolution  string
	OriginalErr error
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict at %s: %v", e.Path, e.OriginalErr)
}

func (e *ConflictError) Unwrap() error {
	return e.OriginalErr
}

func NewConflictError(path string, err error, message string, resolution string) error {
	return &ConflictError{
		Path:        path,
		OriginalErr: err,
		Message:     message,
		Resolution:  resolution,
	}
}
