package types

// CommonConfig contains configuration fields shared between packages
type CommonConfig struct {
	Version     string
	Environment string
	Settings    Settings
}

type Settings struct {
	LogLevel       string
	AutoUpdate     bool
	UpdateInterval string
}
