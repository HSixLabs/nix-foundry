package cmd

import (
	"github.com/shawnkhoffman/nix-foundry/pkg/constants"
)

// Package-level variables for test and test-summary commands
var (
	testSummaryPath string // Directory for test summaries
)

// Package-level constants for test and test-summary commands
const testSummaryDefaultPath = constants.TestSummaryDefaultDir
