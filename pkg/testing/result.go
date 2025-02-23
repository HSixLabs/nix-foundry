package testing

import "time"

// TestResult represents the result of a single test
type TestResult struct {
	Name      string
	Passed    bool
	Error     error
	StartTime time.Time
	EndTime   time.Time
}

// Duration returns the duration of the test
func (r *TestResult) Duration() time.Duration {
	return r.EndTime.Sub(r.StartTime)
}

// Status returns a string representation of the test status
func (r *TestResult) Status() string {
	if r.Passed {
		return "✅ Passed"
	}
	return "❌ Failed"
}

// ErrorMessage returns the error message if the test failed, or an empty string if it passed
func (r *TestResult) ErrorMessage() string {
	if r.Error != nil {
		return r.Error.Error()
	}
	return ""
}
