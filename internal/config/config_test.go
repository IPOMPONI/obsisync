package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const validDataYAML = `
server:
  host: "192.1.0.5"
  port: 5324

auth:
  username: "obsidian"
  password: "sync-secret"

vaults:
  /personal: ./data/personal
  /work: ./data/work
  /shared: ./data/shared
`

const dataWithoutServerYAML = `
auth:
  username: "obsidian"
  password: "sync-secret"

vaults:
  /personal: ./data/personal
  /work: ./data/work
  /shared: ./data/shared
`

const portZeroDataYAML = `
server:
  host: "192.1.0.5"
  port: 0

auth:
  username: "obsidian"
  password: "sync-secret"

vaults:
  /personal: ./data/personal
  /work: ./data/work
  /shared: ./data/shared
`

const emptyUsernameDataYAML = `
server:
  host: "192.1.0.5"
  port: 8080

auth:
  username:
  password: "sync-secret"

vaults:
  /personal: ./data/personal
  /work: ./data/work
  /shared: ./data/shared
`

const emptyPasswordDataYAML = `
server:
  host: "192.1.0.5"
  port: 8080

auth:
  username: "obsidian"
  password:

vaults:
  /personal: ./data/personal
  /work: ./data/work
  /shared: ./data/shared
`

const emptyVaultsDataYAML = `
server:
  host: "192.1.0.5"
  port: 8080

auth:
  username: "obsidian"
  password: "sync-secret"

vaults:
`

const badDataYAML = "server:\n\tport: 8080\n"

const wrongTypeDataYAML = `
server:
  port: "not-a-number"

auth:
  username: "obsidian"
  password: "sync-secret"

vaults:
  /personal: ./data/personal
`

const relativeVaultPathsDataYAML = `
server:
  host: "192.1.0.5"
  port: 5324

auth:
  username: "obsidian"
  password: "sync-secret"

vaults:
  /personal: ./data/personal
  /work: ~/data/work
  /shared: ../../Obsidian/data/shared
`

func writeConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "config_test.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}

func checkLoadFieldsValid(t *testing.T, cfg *Config, configPath string) {
	t.Helper()

	if cfg.Server.Host != "192.1.0.5" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "192.1.0.5")
	}

	if cfg.Server.Port != 5324 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 5324)
	}

	if cfg.Auth.Username != "obsidian" {
		t.Errorf("Auth.Username = %q, want %q", cfg.Auth.Username, "obsidian")
	}

	if cfg.Auth.Password != "sync-secret" {
		t.Errorf("Auth.Password = %q, want %q", cfg.Auth.Password, "sync-secret")
	}

	configDir := filepath.Dir(configPath)

	vaults := map[string]string{
		"/personal": filepath.Join(configDir, "data/personal"),
		"/work":     filepath.Join(configDir, "data/work"),
		"/shared":   filepath.Join(configDir, "data/shared"),
	}

	for name, want := range vaults {
		if cfg.Vaults[name] != want {
			t.Errorf("Vaults[%q] = %q, want %q", name, cfg.Vaults[name], want)
		}
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()

	configPath := writeConfig(t, validDataYAML)
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	checkLoadFieldsValid(t, cfg, configPath)
}

func checkDefaultServerFields(t *testing.T, cfg *Config) {
	t.Helper()

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %q, want %q", cfg.Server.Host, "0.0.0.0")
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
}

func TestLoadWithoutServer(t *testing.T) {
	t.Parallel()

	cfg, err := Load(writeConfig(t, dataWithoutServerYAML))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	checkDefaultServerFields(t, cfg)
}

func TestLoadValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name:     "zero server port",
			data:     portZeroDataYAML,
			expected: "server.port must not be 0",
		},
		{
			name:     "empty auth username",
			data:     emptyUsernameDataYAML,
			expected: "auth.username must not be empty",
		},
		{
			name:     "empty auth password",
			data:     emptyPasswordDataYAML,
			expected: "auth.password must not be empty",
		},
		{
			name:     "empty vaults",
			data:     emptyVaultsDataYAML,
			expected: "vaults must have at least one value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := Load(writeConfig(t, tt.data))
			if err == nil {
				t.Fatalf("Load() error = nil, want error containing %q", tt.expected)
			}

			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("Load() error = %v, want error containing %q", err, tt.expected)
			}
		})
	}
}

func TestLoadWithoutFile(t *testing.T) {
	t.Parallel()

	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Fatal("Load() error = nil, want error for missing file")
	}
}

func TestLoadBadYAML(t *testing.T) {
	t.Parallel()

	_, err := Load(writeConfig(t, badDataYAML))
	if err == nil {
		t.Fatal("Load() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "yaml") {
		t.Errorf("Load() error = %v, want yaml parse error", err)
	}
}

func TestLoadWrongType(t *testing.T) {
	t.Parallel()

	_, err := Load(writeConfig(t, wrongTypeDataYAML))
	if err == nil {
		t.Fatal("Load() error = nil, want unmarshal error")
	}
}

func TestResolveVaultPaths(t *testing.T) {
	t.Parallel()

	configPath := writeConfig(t, relativeVaultPathsDataYAML)
	cfg, _ := Load(configPath)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("skipping home dir test: %v", err)
	}

	configDir := filepath.Dir(configPath)

	tests := []struct {
		name     string
		vault    string
		expected string
	}{
		{
			name:     "resolve vault path `/personal`",
			vault:    "/personal",
			expected: filepath.Join(configDir, "data/personal"),
		},
		{
			name:     "resolve vault path `/work`",
			vault:    "/work",
			expected: filepath.Join(homeDir, "data/work"),
		},
		{
			name:     "resolve vault path `/shared`",
			vault:    "/shared",
			expected: filepath.Join(configDir, "../../Obsidian/data/shared"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := cfg.Vaults[tt.vault]
			if got != tt.expected {
				t.Errorf("Vaults[%q] = %q, want %q", tt.vault, got, tt.expected)
			}
		})
	}

}

func TestResolvePath(t *testing.T) {
	tempFilePath := filepath.Join(t.TempDir(), "config_test.yaml")

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir() error = %v", err)
	}

	defaultPath := filepath.Join(userConfigDir, "obsisync", "config.yaml")

	tests := []struct {
		name     string
		path     string
		env      string
		expected string
	}{
		{
			name:     "path from flag takes precedence",
			path:     tempFilePath,
			env:      "/env/config.yaml",
			expected: tempFilePath,
		},
		{
			name:     "path from environment variable",
			path:     "",
			env:      "/env/config.yaml",
			expected: "/env/config.yaml",
		},
		{
			name:     "default path when no flag or env",
			path:     "",
			env:      "",
			expected: defaultPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("OBSISYNC_CONFIG_PATH", tt.env)

			got, err := ResolvePath(tt.path)
			if err != nil {
				t.Fatalf("ResolvePath(%q) error = %v", tt.path, err)
			}
			if got != tt.expected {
				t.Errorf("ResolvePath(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}
