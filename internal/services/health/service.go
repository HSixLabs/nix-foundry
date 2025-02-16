package health

// Add status constants at top of file
type CheckStatus int

const (
	StatusOK CheckStatus = iota
	StatusWarning
	StatusError
)

type Checker interface {
	RunChecks() []CheckResult
	AuditConfigs() []CheckResult
}

// Add concrete implementations
// type SystemChecker struct{}

// func NewSystemChecker() *SystemChecker {
//     return &SystemChecker{}
// }
