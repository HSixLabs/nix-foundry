package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shawnkhoffman/nix-foundry/pkg/testing"
	"github.com/spf13/cobra"
)

var testSummaryCmd = &cobra.Command{
	Use:   "test-summary [flags] [summary-file]",
	Short: "View test summary information",
	Long: `View test summary information from previous test runs.

If no summary file is specified, displays a list of available summaries.
Use flags to control the output format, filtering options, and statistics display.

Filtering Options:
  - Time-based: Filter by age using --since and --until
  - Platform: Filter by operating system using --platform
  - Test Type: Filter by test type (platform/integration) using --type
  - Status: Show only tests with/without errors using --has-errors/--no-errors
  - Performance: Filter by test count or runtime using --min/max-tests and --min/max-runtime

Statistics:
  Use --stats to view aggregate information about matching test runs, including:
  - Total number of tests and test runs
  - Success/failure rates
  - Min/max/average durations
  - Platform and test type distribution`,
	Example: `  # List all available test summaries
  nix-foundry test-summary

  # View a specific test summary
  nix-foundry test-summary test-results/test-summary-20250222-205500.json

  # View the latest test summary
  nix-foundry test-summary --latest

  # View summaries from the last 24 hours
  nix-foundry test-summary --since 24h

  # View summaries with test statistics
  nix-foundry test-summary --stats

  # Filter by platform and show only failed tests
  nix-foundry test-summary --platform linux --has-errors

  # View integration tests that took longer than 5 minutes
  nix-foundry test-summary --type integration --min-runtime 5m

  # Complex filtering with statistics
  nix-foundry test-summary --since 7d --platform macos --min-tests 10 --stats`,
	RunE: viewTestSummary,
}

var (
	latest     bool
	since      string
	until      string
	jsonOutput bool
	platform   string
	hasErrors  bool
	noErrors   bool
	testType   string
	minTests   int
	maxTests   int
	minRuntime string
	maxRuntime string
	showStats  bool
)

func init() {
	rootCmd.AddCommand(testSummaryCmd)

	// Add flags
	testSummaryCmd.Flags().StringVar(&testSummaryPath, "dir", testSummaryDefaultPath, "Directory containing test summaries")
	testSummaryCmd.Flags().BoolVarP(&latest, "latest", "l", false, "Show only the latest test summary")
	testSummaryCmd.Flags().StringVarP(&since, "since", "s", "", "Show summaries newer than duration (e.g., 24h, 7d, 30m)")
	testSummaryCmd.Flags().StringVarP(&until, "until", "u", "", "Show summaries older than duration (e.g., 1h, 2d, 15m)")
	testSummaryCmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output in JSON format")
	testSummaryCmd.Flags().BoolVar(&showStats, "stats", false, "Show test statistics")
	testSummaryCmd.Flags().StringVarP(&platform, "platform", "p", "", "Filter by platform (macos, linux, windows)")
	testSummaryCmd.Flags().StringVarP(&testType, "type", "t", "", "Filter by test type (platform, integration)")
	testSummaryCmd.Flags().BoolVar(&hasErrors, "has-errors", false, "Show only summaries with errors")
	testSummaryCmd.Flags().BoolVar(&noErrors, "no-errors", false, "Show only summaries without errors")
	testSummaryCmd.Flags().IntVar(&minTests, "min-tests", 0, "Show only summaries with at least N tests")
	testSummaryCmd.Flags().IntVar(&maxTests, "max-tests", 0, "Show only summaries with at most N tests")
	testSummaryCmd.Flags().StringVar(&minRuntime, "min-runtime", "", "Show only summaries with runtime >= duration (e.g., 1m, 5s)")
	testSummaryCmd.Flags().StringVar(&maxRuntime, "max-runtime", "", "Show only summaries with runtime <= duration (e.g., 10m, 30s)")

	// Add flag conflicts
	testSummaryCmd.MarkFlagsMutuallyExclusive("has-errors", "no-errors")
	testSummaryCmd.MarkFlagsMutuallyExclusive("latest", "since")
	testSummaryCmd.MarkFlagsMutuallyExclusive("latest", "until")
}

func validateTimeFilters(filter *testing.SummaryFilter) error {
	if since != "" {
		sinceDuration, parseErr := time.ParseDuration(since)
		if parseErr != nil {
			return fmt.Errorf("invalid since duration: %w", parseErr)
		}
		if sinceDuration < 0 {
			return fmt.Errorf("since duration cannot be negative: %s", since)
		}
		filter.Since = time.Now().Add(-sinceDuration)
	}

	if until != "" {
		untilDuration, parseErr := time.ParseDuration(until)
		if parseErr != nil {
			return fmt.Errorf("invalid until duration: %w", parseErr)
		}
		if untilDuration < 0 {
			return fmt.Errorf("until duration cannot be negative: %s", until)
		}
		filter.Until = time.Now().Add(-untilDuration)
	}

	if filter.Since.After(filter.Until) && !filter.Since.IsZero() && !filter.Until.IsZero() {
		return fmt.Errorf("since time (%s ago) must be before until time (%s ago)", since, until)
	}

	return nil
}

func validateRuntimeFilters(filter *testing.SummaryFilter) error {
	if minRuntime != "" {
		minDuration, parseErr := time.ParseDuration(minRuntime)
		if parseErr != nil {
			return fmt.Errorf("invalid min-runtime duration: %w", parseErr)
		}
		if minDuration < 0 {
			return fmt.Errorf("min-runtime cannot be negative: %s", minRuntime)
		}
		filter.MinRuntime = minDuration
	}

	if maxRuntime != "" {
		maxDuration, parseErr := time.ParseDuration(maxRuntime)
		if parseErr != nil {
			return fmt.Errorf("invalid max-runtime duration: %w", parseErr)
		}
		if maxDuration < 0 {
			return fmt.Errorf("max-runtime cannot be negative: %s", maxRuntime)
		}
		filter.MaxRuntime = maxDuration
	}

	if filter.MinRuntime > 0 && filter.MaxRuntime > 0 && filter.MinRuntime > filter.MaxRuntime {
		return fmt.Errorf("min-runtime (%s) cannot be greater than max-runtime (%s)", minRuntime, maxRuntime)
	}

	return nil
}

func validateTestCountFilters(filter *testing.SummaryFilter) error {
	if minTests < 0 {
		return fmt.Errorf("min-tests cannot be negative: %d", minTests)
	}
	if maxTests < 0 {
		return fmt.Errorf("max-tests cannot be negative: %d", maxTests)
	}
	if maxTests > 0 && minTests > maxTests {
		return fmt.Errorf("min-tests (%d) cannot be greater than max-tests (%d)", minTests, maxTests)
	}

	filter.MinTests = minTests
	filter.MaxTests = maxTests
	return nil
}

func validatePlatformAndType(filter *testing.SummaryFilter) error {
	if platform != "" {
		platform = strings.ToLower(platform)
		switch platform {
		case "macos", "darwin", "linux", "windows":
			if platform == "darwin" {
				platform = "macos"
			}
			filter.Platform = platform
		default:
			return fmt.Errorf("invalid platform: %q (must be 'macos', 'linux', or 'windows')", platform)
		}
	}

	if testType != "" {
		testType = strings.ToLower(testType)
		if testType != "platform" && testType != "integration" {
			return fmt.Errorf("invalid test type: %q (must be 'platform' or 'integration')", testType)
		}
		filter.TestTypes = []string{testType}
	}

	return nil
}

func buildFilter() (*testing.SummaryFilter, error) {
	filter := &testing.SummaryFilter{}

	if err := validateTimeFilters(filter); err != nil {
		return nil, err
	}

	if err := validateRuntimeFilters(filter); err != nil {
		return nil, err
	}

	if err := validateTestCountFilters(filter); err != nil {
		return nil, err
	}

	if err := validatePlatformAndType(filter); err != nil {
		return nil, err
	}

	if hasErrors {
		t := true
		filter.HasErrors = &t
	} else if noErrors {
		f := false
		filter.HasErrors = &f
	}

	return filter, nil
}

func viewTestSummary(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return showSingleSummary(args[0])
	}

	if mkdirErr := os.MkdirAll(testSummaryPath, 0755); mkdirErr != nil {
		return fmt.Errorf("failed to create summary directory: %w", mkdirErr)
	}

	filter, err := buildFilter()
	if err != nil {
		return err
	}

	summaries, err := testing.FilterSummaries(testSummaryPath, *filter)
	if err != nil {
		return err
	}

	if len(summaries) == 0 {
		return fmt.Errorf("no matching test summaries found")
	}

	if latest {
		return showSingleSummary(summaries[0])
	}

	fmt.Printf("Found %d matching test summaries:\n\n", len(summaries))
	for _, summary := range summaries {
		fmt.Printf("📊 %s\n", summary.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Platform: %s (WSL: %v)\n", summary.Platform, summary.IsWSL)
		fmt.Printf("   Tests: %d passed in %s\n", summary.TotalTests, summary.TotalDuration.Round(time.Millisecond))
		if len(summary.PlatformTests) > 0 {
			fmt.Printf("   Platform Tests: %d\n", len(summary.PlatformTests))
		}
		if len(summary.Integration) > 0 {
			fmt.Printf("   Integration Tests: %d\n", len(summary.Integration))
		}
		fmt.Println()
	}

	if showStats {
		fmt.Println(testing.GetTestStats(summaries))
	}

	return nil
}

func showSingleSummary(summary interface{}) error {
	var s *testing.TestSummary
	var filename string

	switch v := summary.(type) {
	case string:
		loadedSummary, loadErr := testing.LoadTestSummary(v)
		if loadErr != nil {
			return fmt.Errorf("failed to load test summary: %w", loadErr)
		}
		s = loadedSummary
		filename = filepath.Base(v)
	case *testing.TestSummary:
		s = v
		filename = s.Timestamp.Format("2006-01-02-150405")
	default:
		return fmt.Errorf("invalid summary type: %T", summary)
	}

	if jsonOutput {
		fmt.Printf("%s\n", summary)
		return nil
	}

	fmt.Printf("Test Summary: %s\n", filename)
	fmt.Printf("====================\n\n")
	fmt.Printf("Platform: %s (WSL: %v)\n", s.Platform, s.IsWSL)
	fmt.Printf("Timestamp: %s\n", s.Timestamp.Format(time.RFC3339))
	fmt.Printf("Total Duration: %s\n", s.TotalDuration.Round(time.Millisecond))
	fmt.Printf("Total Tests: %d\n\n", s.TotalTests)

	if len(s.PlatformTests) > 0 {
		fmt.Println("Platform Tests:")
		for _, test := range s.PlatformTests {
			fmt.Printf("  %s: %s (%s)\n", test.Name, test.Status(), test.Duration().Round(time.Millisecond))
			if test.Error != nil {
				fmt.Printf("    Error: %s\n", test.ErrorMessage())
			}
		}
		fmt.Println()
	}

	if len(s.Integration) > 0 {
		fmt.Println("Integration Tests:")
		for _, test := range s.Integration {
			fmt.Printf("  %s: %s (%s)\n", test.Name, test.Status(), test.Duration().Round(time.Millisecond))
			if test.Error != nil {
				fmt.Printf("    Error: %s\n", test.ErrorMessage())
			}
		}
		fmt.Println()
	}

	return nil
}
