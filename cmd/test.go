// Package cmd provides the command-line interface for Nix Foundry.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
	"github.com/shawnkhoffman/nix-foundry/pkg/testing"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	testFilter string
	testConfig string
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests",
	Long: `Run tests.
This command runs integration tests to verify Nix Foundry functionality.`,
	RunE: runTest,
}

func init() {
	rootCmd.AddCommand(testCmd)
	testCmd.Flags().StringVarP(&testFilter, "filter", "f", "", "Filter tests by name")
	testCmd.Flags().StringVarP(&testConfig, "config", "c", "", "Path to test configuration file")
}

// TestingT implements testing.T interface.
type TestingT struct{}

func (t *TestingT) Helper() {}

func (t *TestingT) Fatalf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}

func (t *TestingT) Errorf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func runTest(cmd *cobra.Command, args []string) error {
	config := testing.GetTestConfig()

	testConfig := testing.TestConfig{
		Files: map[string]string{
			"config.yaml": `version: v1
kind: NixConfig
metadata:
  name: test
  description: Test configuration
settings:
  shell: /bin/bash
  logLevel: debug
  autoUpdate: true
  updateInterval: 1h
nix:
  manager: nix-env
  packages:
    core:
      - git
      - curl
    optional:
      - ripgrep
      - jq
  scripts:
    - name: test
      description: Test script
      commands: |
        echo 'test'
`,
		},
	}

	t := &TestingT{}
	tmpDir, cleanup := testing.SetupTest(t, testConfig)
	defer cleanup()

	fmt.Printf("Running test in %s\n", tmpDir)

	configPath := filepath.Join(tmpDir, "config.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var actual schema.Config
	if err := yaml.Unmarshal(content, &actual); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	if err := schema.ValidateConfig(&actual); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	testing.CompareConfigs(t, config, &actual)

	fmt.Println("✨ Tests completed successfully")
	return nil
}
