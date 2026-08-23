package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Host string `yaml:"host" env-required:"true"`
	Port uint16 `yaml:"port" env-required:"true"`
}

type Auth struct {
	Username string `yaml:"username" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
}

type Config struct {
	Server Server            `yaml:"server"`
	Auth   Auth              `yaml:"auth"`
	Vaults map[string]string `yaml:"vaults" env-required:"true"`
}

func CheckPath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}

	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get default config dir: %w", err)
	}

	return filepath.Join(configDir, "obsisync", "config.yaml"), nil
}

func Load(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		return nil, fmt.Errorf("yaml unmarshal: %w", err)
	}

	return &cfg, nil
}
