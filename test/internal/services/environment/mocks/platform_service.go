package mocks

import (
	"github.com/shawnkhoffman/nix-foundry/internal/services/platform"
	"github.com/stretchr/testify/mock"
)

// MockPlatformService implements platform.Service for testing
type MockPlatformService struct {
	mock.Mock
}

func (m *MockPlatformService) Initialize(testMode bool) error {
	args := m.Called(testMode)
	return args.Error(0)
}

func (m *MockPlatformService) SetupIsolation(testMode bool) error {
	args := m.Called(testMode)
	return args.Error(0)
}

func (m *MockPlatformService) EnableFlakeFeatures() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPlatformService) GetCurrentEnvironment() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockPlatformService) SetupEnvironmentSymlink() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPlatformService) CheckPrerequisites(testMode bool) error {
	args := m.Called(testMode)
	return args.Error(0)
}

func (m *MockPlatformService) InstallHomeManager() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPlatformService) RestoreFromBackup(backupPath string, targetPath string) error {
	args := m.Called(backupPath, targetPath)
	return args.Error(0)
}

func (m *MockPlatformService) SetupPlatform(testMode bool) error {
	args := m.Called(testMode)
	return args.Error(0)
}

func (m *MockPlatformService) ValidateBackup(backupPath string) error {
	args := m.Called(backupPath)
	return args.Error(0)
}

func (m *MockPlatformService) GetPlatformType() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockPlatformService) IsHomeManagerInstalled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPlatformService) InstallNix() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPlatformService) IsNixInstalled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockPlatformService) Validate() error {
	args := m.Called()
	return args.Error(0)
}

var _ platform.Service = (*MockPlatformService)(nil)
