// Package testing provides testing utilities.
package testing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/shawnkhoffman/nix-foundry/pkg/constants"
)

// TestSummary represents a complete test run summary
type TestSummary struct {
	Platform      string
	Timestamp     time.Time
	TotalDuration time.Duration
	PlatformTests []TestResult
	Integration   []TestResult
	SystemInfo    map[string]string
	IsWSL         bool
	TotalTests    int
}

// SaveTestSummary saves the test results to a JSON file
func SaveTestSummary(summary TestSummary, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	filename := fmt.Sprintf("%s%s%s", constants.TestSummaryFilePrefix, time.Now().Format("20060102-150405"), constants.TestSummaryFileExt)
	outputPath := filepath.Join(outputDir, filename)

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	if err := os.WriteFile(outputPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write summary file: %w", err)
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for TestSummary
func (s TestSummary) MarshalJSON() ([]byte, error) {
	type Alias TestSummary
	return json.Marshal(&struct {
		Alias
		TotalDuration string `json:"totalDuration"`
		Timestamp     string `json:"timestamp"`
	}{
		Alias:         Alias(s),
		TotalDuration: s.TotalDuration.String(),
		Timestamp:     s.Timestamp.Format(time.RFC3339),
	})
}

// String returns a JSON string representation of the TestSummary
func (s TestSummary) String() string {
	data, _ := json.MarshalIndent(s, "", "  ")
	return string(data)
}

// LoadTestSummary loads a test summary from a JSON file
func LoadTestSummary(path string) (*TestSummary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read summary file: %w", err)
	}

	var summary TestSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return nil, fmt.Errorf("failed to unmarshal summary: %w", err)
	}

	return &summary, nil
}

// UnmarshalJSON implements custom JSON unmarshaling for TestSummary
func (s *TestSummary) UnmarshalJSON(data []byte) error {
	type Alias TestSummary
	aux := &struct {
		*Alias
		TotalDuration string `json:"totalDuration"`
		Timestamp     string `json:"timestamp"`
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.TotalDuration != "" {
		duration, err := time.ParseDuration(aux.TotalDuration)
		if err != nil {
			return fmt.Errorf("failed to parse duration: %w", err)
		}
		s.TotalDuration = duration
	}

	if aux.Timestamp != "" {
		timestamp, err := time.Parse(time.RFC3339, aux.Timestamp)
		if err != nil {
			return fmt.Errorf("failed to parse timestamp: %w", err)
		}
		s.Timestamp = timestamp
	}

	return nil
}

// MarshalJSON implements custom JSON marshaling for TestResult
func (r TestResult) MarshalJSON() ([]byte, error) {
	type Alias TestResult
	return json.Marshal(&struct {
		Alias
		Duration string `json:"duration"`
	}{
		Alias:    Alias(r),
		Duration: r.Duration().String(),
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for TestResult
func (r *TestResult) UnmarshalJSON(data []byte) error {
	type Alias TestResult
	aux := &struct {
		*Alias
		Duration string `json:"duration"`
	}{
		Alias: (*Alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Duration != "" {
		duration, err := time.ParseDuration(aux.Duration)
		if err != nil {
			return fmt.Errorf("failed to parse duration: %w", err)
		}
		r.EndTime = r.StartTime.Add(duration)
	}

	return nil
}
