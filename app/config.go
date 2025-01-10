package app

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
    Server struct {
        Port int `yaml:"port"`
    } `yaml:"server"`
}

func LoadConfig() (*Config, error) {
    path := "config/config.yml" // Path is now hardcoded

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
