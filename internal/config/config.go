package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
}

type Auth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Config struct {
	Server Server            `yaml:"server"`
	Auth   Auth              `yaml:"auth"`
	Vaults map[string]string `yaml:"vaults"`
}

func ResolvePath(flagPath string) (string, error) {
	if flagPath != "" {
		return flagPath, nil
	}

	if path := os.Getenv("OBSISYNC_CONFIG_PATH"); path != "" {
		return path, nil
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("get default config dir: %w", err)
	}

	return filepath.Join(configDir, "obsisync", "config.yaml"), nil
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := defaultConfig()
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, fmt.Errorf("yaml unmarshal %q: %w", path, err)
	}

	err = cfg.validate()
	if err != nil {
		return nil, fmt.Errorf("validate config %q: %w", path, err)
	}

	return cfg, nil
}

func defaultConfig() *Config {
	return &Config{Server: Server{Host: "0.0.0.0", Port: 8080}}
}

func (c *Config) validate() error {
	errs := make([]error, 0, 4)

	if c.Server.Port == 0 {
		errs = append(errs, errors.New("server.port must not be 0"))
	}
	if c.Auth.Username == "" {
		errs = append(errs, errors.New("auth.username must not be empty"))
	}
	if c.Auth.Password == "" {
		errs = append(errs, errors.New("auth.password must not be empty"))
	}
	if len(c.Vaults) == 0 {
		errs = append(errs, errors.New("vaults must have at least one value"))
	}

	return errors.Join(errs...)
}
