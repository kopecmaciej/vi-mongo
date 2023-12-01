package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Mongui MonguiConfig `yaml:"mongui"`
}

type MonguiConfig struct {
	Log   LogConfig   `yaml:"log"`
	Mongo MongoConfig `yaml:"mongo"`
	Debug bool        `yaml:"debug"`
}

type MongoConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type LogConfig struct {
	Path        string `yaml:"path"`
	Level       string `yaml:"level"`
	PrettyPrint bool   `yaml:"prettyPrint"`
}

func LoadAppConfig() (*MonguiConfig, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config.Mongui, nil
}
