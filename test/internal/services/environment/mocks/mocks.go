package mocks

import (
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/environment"
	"github.com/stretchr/testify/mock"
)

var _ environment.Service = (*Service)(nil)

type Service struct {
	mock.Mock
}

// Rollback mock
func (m *Service) Rollback(target time.Time, force bool) error {
	args := m.Called(target, force)
	return args.Error(0)
}

// Add all other required interface methods with mock implementations
func (m *Service) CheckHealth() string {
	args := m.Called()
	return args.String(0)
}

func (m *Service) ListEnvironments() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *Service) CreateEnvironment(name string, template string) error {
	args := m.Called(name, template)
	return args.Error(0)
}

func (m *Service) Switch(target string, force bool) error {
	args := m.Called(target, force)
	return args.Error(0)
}

// Add remaining methods with empty implementations for testing purposes
func (m *Service) Initialize(testMode bool) error                   { return nil }
func (m *Service) CheckPrerequisites(testMode bool) error           { return nil }
func (m *Service) SetupIsolation(testMode bool) error               { return nil }
func (m *Service) InstallBinary() error                             { return nil }
func (m *Service) RestoreEnvironment(backupPath string) error       { return nil }
func (m *Service) ValidateRestoredEnvironment(envPath string) error { return nil }
func (m *Service) GetCurrentEnvironment() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// Add other interface method mocks as needed for testing

// Add this method to the mock Service
func (m *Service) EnableFlakeFeatures() error {
	args := m.Called()
	return args.Error(0)
}

// Add this method to SetupEnvironmentSymlink
func (m *Service) SetupEnvironmentSymlink() error {
	args := m.Called()
	return args.Error(0)
}

// Add this method to InitializeNixFlake
func (m *Service) InitializeNixFlake() error {
	args := m.Called()
	return args.Error(0)
}
