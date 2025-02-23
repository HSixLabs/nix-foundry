// Package testing provides testing utilities.
package testing

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SummaryFilter defines criteria for filtering test summaries
type SummaryFilter struct {
	Since      time.Time
	Until      time.Time
	Platform   string
	HasErrors  *bool
	TestTypes  []string
	MinTests   int
	MaxTests   int
	MinRuntime time.Duration
	MaxRuntime time.Duration
}

// FilterSummaries returns test summaries that match the given filter criteria
func FilterSummaries(dir string, filter SummaryFilter) ([]*TestSummary, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var summaries []*TestSummary
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			summary, err := LoadTestSummary(filepath.Join(dir, entry.Name()))
			if err != nil {
				continue
			}
			summaries = append(summaries, summary)
		}
	}

	var filtered []*TestSummary
	for _, summary := range summaries {
		if filter.matches(summary) {
			filtered = append(filtered, summary)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	return filtered, nil
}

func (f SummaryFilter) matches(summary *TestSummary) bool {
	if !f.Since.IsZero() && summary.Timestamp.Before(f.Since) {
		return false
	}

	if !f.Until.IsZero() && summary.Timestamp.After(f.Until) {
		return false
	}

	if f.Platform != "" && !strings.EqualFold(summary.Platform, f.Platform) {
		return false
	}

	if f.HasErrors != nil {
		hasErrors := false
		for _, test := range summary.PlatformTests {
			if !test.Passed {
				hasErrors = true
				break
			}
		}
		for _, test := range summary.Integration {
			if !test.Passed {
				hasErrors = true
				break
			}
		}
		if *f.HasErrors != hasErrors {
			return false
		}
	}

	if len(f.TestTypes) > 0 {
		hasType := false
		for _, testType := range f.TestTypes {
			if strings.EqualFold(testType, "platform") && len(summary.PlatformTests) > 0 {
				hasType = true
				break
			}
			if strings.EqualFold(testType, "integration") && len(summary.Integration) > 0 {
				hasType = true
				break
			}
		}
		if !hasType {
			return false
		}
	}

	totalTests := len(summary.PlatformTests) + len(summary.Integration)
	if f.MinTests > 0 && totalTests < f.MinTests {
		return false
	}

	if f.MaxTests > 0 && totalTests > f.MaxTests {
		return false
	}

	if f.MinRuntime > 0 && summary.TotalDuration < f.MinRuntime {
		return false
	}

	if f.MaxRuntime > 0 && summary.TotalDuration > f.MaxRuntime {
		return false
	}

	return true
}

// TestStats represents statistics about test results
type TestStats struct {
	FirstRun         time.Time
	LastRun          time.Time
	TotalRuns        int
	PlatformTests    int
	IntegrationTests int
	TotalTests       int
	MinDuration      time.Duration
	MaxDuration      time.Duration
	AvgDuration      time.Duration
	AvgTestsPerRun   float64
	PassRate         float64
}

// GetTestStats returns statistics about test results
func GetTestStats(summaries []*TestSummary) TestStats {
	if len(summaries) == 0 {
		return TestStats{}
	}

	stats := TestStats{
		FirstRun: summaries[len(summaries)-1].Timestamp,
		LastRun:  summaries[0].Timestamp,
	}

	for _, summary := range summaries {
		stats.TotalRuns++
		stats.PlatformTests += len(summary.PlatformTests)
		stats.IntegrationTests += len(summary.Integration)
		stats.TotalTests = stats.PlatformTests + stats.IntegrationTests

		if stats.MinDuration == 0 || summary.TotalDuration < stats.MinDuration {
			stats.MinDuration = summary.TotalDuration
		}
		if summary.TotalDuration > stats.MaxDuration {
			stats.MaxDuration = summary.TotalDuration
		}
		stats.AvgDuration += summary.TotalDuration
	}

	stats.AvgDuration /= time.Duration(stats.TotalRuns)
	stats.AvgTestsPerRun = float64(stats.TotalTests) / float64(stats.TotalRuns)

	var passed int
	for _, summary := range summaries {
		for _, test := range summary.PlatformTests {
			if test.Passed {
				passed++
			}
		}
		for _, test := range summary.Integration {
			if test.Passed {
				passed++
			}
		}
	}

	if stats.TotalTests > 0 {
		stats.PassRate = float64(passed) / float64(stats.TotalTests)
	}

	return stats
}
