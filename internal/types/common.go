package types

type CommonConfig struct {
    Version     string            `yaml:"version"`
    Environment string            `yaml:"environment"`
    Settings    map[string]string `yaml:"settings"`
}
