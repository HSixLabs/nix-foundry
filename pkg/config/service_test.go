package config

import (
	"encoding/json"
	"testing"

	"github.com/shawnkhoffman/nix-foundry/pkg/schema"
)

func TestDiffPackages(t *testing.T) {
	tests := []struct {
		name              string
		installedPackages []string
		desiredPackages   schema.Packages
		expected          schema.PackageDiff
	}{
		{
			name:              "add packages",
			installedPackages: []string{"git"},
			desiredPackages: schema.Packages{
				Core:     []string{"git", "nodejs"},
				Optional: []string{"docker"},
			},
			expected: schema.PackageDiff{
				ToInstall: []string{"nodejs", "docker"},
				ToRemove:  []string{},
			},
		},
		{
			name:              "remove packages",
			installedPackages: []string{"git", "nodejs", "docker", "python3"},
			desiredPackages: schema.Packages{
				Core: []string{"git"},
			},
			expected: schema.PackageDiff{
				ToInstall: []string{},
				ToRemove:  []string{"nodejs", "docker", "python3"},
			},
		},
		{
			name:              "mixed changes",
			installedPackages: []string{"git", "oldpackage", "python3"},
			desiredPackages: schema.Packages{
				Core:     []string{"git", "nodejs"},
				Optional: []string{"docker"},
			},
			expected: schema.PackageDiff{
				ToInstall: []string{"nodejs", "docker"},
				ToRemove:  []string{"oldpackage", "python3"},
			},
		},
		{
			name:              "no changes",
			installedPackages: []string{"git", "docker"},
			desiredPackages: schema.Packages{
				Core:     []string{"git"},
				Optional: []string{"docker"},
			},
			expected: schema.PackageDiff{
				ToInstall: []string{},
				ToRemove:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := schema.DiffPackages(tt.installedPackages, tt.desiredPackages)

			if len(diff.ToInstall) != len(tt.expected.ToInstall) {
				t.Errorf("ToInstall length mismatch: got %d, want %d", len(diff.ToInstall), len(tt.expected.ToInstall))
			}
			for _, pkg := range tt.expected.ToInstall {
				found := false
				for _, installed := range diff.ToInstall {
					if installed == pkg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to install package %s, but it was not found", pkg)
				}
			}

			if len(diff.ToRemove) != len(tt.expected.ToRemove) {
				t.Errorf("ToRemove length mismatch: got %d, want %d", len(diff.ToRemove), len(tt.expected.ToRemove))
			}
			for _, pkg := range tt.expected.ToRemove {
				found := false
				for _, removed := range diff.ToRemove {
					if removed == pkg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to remove package %s, but it was not found", pkg)
				}
			}
		})
	}
}

func TestGetInstalledPackagesParser(t *testing.T) {
	testJSONOutput := `{
		"0": {
			"name": "git-2.40.1",
			"pname": "git",
			"version": "2.40.1",
			"system": "aarch64-darwin"
		},
		"1": {
			"name": "nodejs-18.16.1",
			"pname": "nodejs",
			"version": "18.16.1",
			"system": "aarch64-darwin"
		},
		"2": {
			"name": "pre-commit-3.5.0",
			"pname": "pre-commit",
			"version": "3.5.0",
			"system": "aarch64-darwin"
		},
		"3": {
			"name": "docker-compose-2.20.0",
			"pname": "docker-compose",
			"version": "2.20.0",
			"system": "aarch64-darwin"
		}
	}`

	var packages map[string]map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(testJSONOutput), &packages); jsonErr != nil {
		t.Fatalf("Failed to parse test JSON: %v", jsonErr)
	}

	var packageNames []string
	for _, pkg := range packages {
		if pname, ok := pkg["pname"].(string); ok {
			packageNames = append(packageNames, pname)
		}
	}

	expected := []string{"git", "nodejs", "pre-commit", "docker-compose"}
	if len(packageNames) != len(expected) {
		t.Errorf("Expected %d packages, got %d", len(expected), len(packageNames))
	}

	expectedMap := make(map[string]bool)
	for _, pkg := range expected {
		expectedMap[pkg] = true
	}

	for _, pkg := range packageNames {
		if !expectedMap[pkg] {
			t.Errorf("Unexpected package found: %s", pkg)
		}
		delete(expectedMap, pkg)
	}

	for pkg := range expectedMap {
		t.Errorf("Expected package not found: %s", pkg)
	}
}
