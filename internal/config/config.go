package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// FileConfig represents the YAML config file structure.
// Field names match CLI flag names.
type FileConfig struct {
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
}
