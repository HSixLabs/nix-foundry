package progress

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// Spinner represents a loading spinner with success/failure states
type Spinner struct {
	spinner *spinner.Spinner
	message string
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = message + " "
	return &Spinner{
		spinner: s,
		message: message,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.spinner.Start()
}

// Stop stops the spinner animation
func (s *Spinner) Stop() {
	s.spinner.Stop()
}

// Success stops the spinner and shows a success message
func (s *Spinner) Success(message string) {
	s.spinner.Stop()
	fmt.Printf("✅ %s\n", message)
}

// Fail stops the spinner and shows a failure message
func (s *Spinner) Fail(message string) {
	s.spinner.Stop()
	fmt.Printf("❌ %s\n", message)
}

// Update changes the spinner message
func (s *Spinner) Update(message string) {
	s.message = message
	s.spinner.Prefix = message + " "
}

// WithProgress wraps a function with a progress spinner
func WithProgress(message string, fn func() error) error {
	s := NewSpinner(message)
	s.Start()

	if err := fn(); err != nil {
		s.Fail("Operation failed")
		return err
	}

	s.Success("Operation completed")
	return nil
}
