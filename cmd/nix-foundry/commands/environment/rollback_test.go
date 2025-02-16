package environment

import (
	"bytes"
	"testing"
	"time"

	"github.com/shawnkhoffman/nix-foundry/internal/services/environment/mocks"
)

func TestRollbackCommand(t *testing.T) {
	// Setup mock service
	mockSvc := &mocks.Service{}
	mockSvc.On("Rollback", time.Now().Add(-time.Hour), true).Return(nil)
	mockSvc.On("Rollback", time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC), false).Return(nil)

	testCases := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"relative time", []string{"1h", "--force"}, false},
		{"exact timestamp", []string{"20230101-120000"}, false},
		{"invalid format", []string{"invalid-time"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := NewRollbackCommand(mockSvc)
			cmd.SetArgs(tc.args)

			var buf bytes.Buffer
			cmd.SetOut(&buf)

			err := cmd.Execute()
			if (err != nil) != tc.expectError {
				t.Errorf("Unexpected error state: %v", err)
			}
		})
	}
}
