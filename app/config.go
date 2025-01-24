package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port  int    `yaml:"port"`
		Token string `yaml:"token"`
	} `yaml:"server"`
	Memory struct {
		MaxData int `yaml:"max-data"`
		MaxSize int `yaml:"max-size"`
	} `yaml:"memory"`
}

var AppConfig *Config

func init() {
	var err error
	AppConfig, err = LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}
}

func LoadConfig() (*Config, error) {
	path := "config/config.yml"

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
