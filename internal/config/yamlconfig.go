package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// YamlConfig represents the YAML config file structure.
// Field names match CLI flag names.
type YamlConfig struct {
	Interface string `yaml:"interface"`
	Protocol  string `yaml:"protocol"`
	Port      int    `yaml:"port"`
	IP        string `yaml:"ip"`
	Direction string `yaml:"direction"`
	Process   string `yaml:"process"`
	PID       int    `yaml:"pid"`
	Stateful  bool   `yaml:"stateful"`
	Verbosity int    `yaml:"verbosity"`
	Output    string `yaml:"output"`
	Debug     bool   `yaml:"debug"`
	LogFile   string `yaml:"log-file"`
	Stats     bool   `yaml:"stats"`
	Graceful  bool   `yaml:"graceful"`
}

// DefaultPath returns the default config file path.
// Handles sudo by checking SUDO_USER environment variable.
func DefaultPath() string {
	// When running with sudo, use the original user's home directory
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		return filepath.Join("/home", sudoUser, ".config", "portlens", "config.yaml")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "portlens", "config.yaml")
}

// Load reads and parses a YAML config file.
// Returns an empty config (not error) if file doesn't exist.
func Load(path string) (*YamlConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &YamlConfig{}, nil
		}
		return nil, err
	}

	var cfg YamlConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
