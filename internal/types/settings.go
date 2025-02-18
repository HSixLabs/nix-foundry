package types

type Settings struct {
    AutoUpdate     bool   `yaml:"autoUpdate"`
    UpdateInterval string `yaml:"updateInterval"`
    LogLevel       string `yaml:"logLevel"`
}
