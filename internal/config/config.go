package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port     int               `yaml:"port"`
	ApiPort  int               `yaml:"api_port"`
	Backends []string          `yaml:"backends"`
	Rules    map[string]string `yaml:"rules"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
