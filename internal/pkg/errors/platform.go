package errors

// PlatformError represents an error that occurred during platform-specific operations
type PlatformError struct {
	Err error
	Op  string
}

func (e PlatformError) Error() string {
	if e.Op != "" {
		return "platform error during " + e.Op + ": " + e.Err.Error()
	}
	return "platform error: " + e.Err.Error()
}

func (e PlatformError) Unwrap() error {
	return e.Err
}

// NewPlatformError creates a new PlatformError
func NewPlatformError(err error, op string) error {
	return PlatformError{
		Err: err,
		Op:  op,
	}
}
