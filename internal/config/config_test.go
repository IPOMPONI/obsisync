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
		t.Fatalf("write config: %v", err)
	}
	return path
}

func checkLoadFieldsValid(t *testing.T, cfg *Config, configPath string) {
	t.Helper()

	t.Run("Server host", func(t *testing.T) {
		if cfg.Server.Host != "192.1.0.5" {
			t.Errorf("got: %q, want: %q", cfg.Server.Host, "192.1.0.5")
		}
	})

	t.Run("Server port", func(t *testing.T) {
		if cfg.Server.Port != 5324 {
			t.Errorf("got: %d, want: %d", cfg.Server.Port, 5324)
		}
	})

	t.Run("Auth username", func(t *testing.T) {
		if cfg.Auth.Username != "obsidian" {
			t.Errorf("got: %q, want: %q", cfg.Auth.Username, "obsidian")
		}
	})

	t.Run("Auth password", func(t *testing.T) {
		if cfg.Auth.Password != "sync-secret" {
			t.Errorf("got: %q, want: %q", cfg.Auth.Password, "sync-secret")
		}
	})

	configDir := filepath.Dir(configPath)
	t.Run("Vault /personal", func(t *testing.T) {
		want := filepath.Join(configDir, "data/personal")
		if cfg.Vaults["/personal"] != want {
			t.Errorf("got: %q, want: %q", cfg.Vaults["/personal"], want)
		}
	})

	t.Run("Vault /work", func(t *testing.T) {
		want := filepath.Join(configDir, "data/work")
		if cfg.Vaults["/work"] != want {
			t.Errorf("got: %q, want: %q", cfg.Vaults["/work"], want)
		}
	})

	t.Run("Vault /shared", func(t *testing.T) {
		want := filepath.Join(configDir, "data/shared")
		if cfg.Vaults["/shared"] != want {
			t.Errorf("got: %q, want: %q", cfg.Vaults["/shared"], want)
		}
	})
}

func TestLoadValid(t *testing.T) {
	t.Parallel()

	configPath := writeConfig(t, validDataYAML)
	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	checkLoadFieldsValid(t, cfg, configPath)
}

func checkDefaultServerFields(t *testing.T, cfg *Config) {
	t.Helper()

	t.Run("Server host", func(t *testing.T) {
		if cfg.Server.Host != "0.0.0.0" {
			t.Errorf("got: %q, want: %q", cfg.Server.Host, "0.0.0.0")
		}
	})

	t.Run("Server port", func(t *testing.T) {
		if cfg.Server.Port != 8080 {
			t.Errorf("got: %d, want: %d", cfg.Server.Port, 8080)
		}
	})
}

func TestLoadWithoutServer(t *testing.T) {
	t.Parallel()

	cfg, err := Load(writeConfig(t, dataWithoutServerYAML))
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	checkDefaultServerFields(t, cfg)
}

func TestLoadValidateErrs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     string
		expected string
	}{
		{
			name:     "Zero server port",
			data:     portZeroDataYAML,
			expected: "server.port must not be 0",
		},
		{
			name:     "Empty auth username",
			data:     emptyUsernameDataYAML,
			expected: "auth.username must not be empty",
		},
		{
			name:     "Empty auth password",
			data:     emptyPasswordDataYAML,
			expected: "auth.password must not be empty",
		},
		{
			name:     "Empty vaults",
			data:     emptyVaultsDataYAML,
			expected: "vaults must have at least one value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(writeConfig(t, tt.data))
			if err == nil {
				t.Fatalf("want error containing %q, got nil", tt.expected)
			}
			if !strings.Contains(err.Error(), tt.expected) {
				t.Errorf("err = %v, want contains %q", err, tt.expected)
			}
		})
	}
}

func TestLoadWithoutFile(t *testing.T) {
	t.Parallel()

	_, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err == nil {
		t.Fatal("want error for missing file, got nil")
	}
}

func TestLoadBadYAML(t *testing.T) {
	t.Parallel()

	_, err := Load(writeConfig(t, badDataYAML))
	if err == nil {
		t.Fatal("want parse error, got nil")
	}
	if !strings.Contains(err.Error(), "yaml") {
		t.Errorf("err = %v, want yaml parse error", err)
	}
}

func TestLoadWrongType(t *testing.T) {
	t.Parallel()

	_, err := Load(writeConfig(t, wrongTypeDataYAML))
	if err == nil {
		t.Fatal("want unmarshal error, got nil")
	}
}

func TestResolveVaultPaths(t *testing.T) {
	t.Parallel()

	configPath := writeConfig(t, relativeVaultPathsDataYAML)
	cfg, _ := Load(configPath)

	configDir := filepath.Dir(configPath)
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("skipping home dir test: %v", err)
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{
			name:     "Resolve vault path `/personal`",
			got:      cfg.Vaults["/personal"],
			expected: filepath.Join(filepath.Dir(configPath), "data/personal"),
		},
		{
			name:     "Resolve vault path `/work`",
			got:      cfg.Vaults["/work"],
			expected: filepath.Join(homeDir, "data/work"),
		},
		{
			name:     "Resolve vault path `/shared`",
			got:      cfg.Vaults["/shared"],
			expected: filepath.Join(configDir, "../../Obsidian/data/shared"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got: %q, want: %q", tt.got, tt.expected)
			}
		})
	}

}

func TestResolvePath(t *testing.T) {
	tempFilePath := filepath.Join(t.TempDir(), "config_test.yaml")

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("get user config dir: %v", err)
	}
	defaultPath := filepath.Join(userConfigDir, "obsisync", "config.yaml")

	tests := []struct {
		name     string
		path     string
		env      string
		expected string
	}{
		{
			name:     "Path in flag",
			path:     tempFilePath,
			env:      "/env/config.yaml",
			expected: tempFilePath,
		},
		{
			name:     "Path in env",
			path:     "",
			env:      "/env/config.yaml",
			expected: "/env/config.yaml",
		},
		{
			name:     "Default path",
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
				t.Fatalf("ResolvePath: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got: %v, want: %v", got, tt.expected)
			}
		})
	}
}
