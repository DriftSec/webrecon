package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

var Cfg Config // gllobal config access for core

type Config struct {
	General struct {
		Debug   bool   `yaml:"debug"`
		Errors  bool   `yaml:"errors"`
		DataDir string `yaml:"data_dir"`
	} `yaml:"general"`
	Recon struct {
		TargetID Runners `yaml:"target_identification"`
		Flyover  Runners `yaml:"flyover"`
	} `yaml:"recon"`
}

// InitConfig returns a new decoded Config struct
func (c *Config) Init(configPath string) error {
	err := validateConfigPath(configPath)
	if err != nil {
		return err
	}

	file, err := os.Open(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)

	if err := d.Decode(&c); err != nil {
		return err
	}
	Cfg = *c
	return nil
}

// validateConfigPath just makes sure, that the path provided is a file, and can be read
func validateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
