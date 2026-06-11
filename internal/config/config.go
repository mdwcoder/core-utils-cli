package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	AppName            = "cu"
	AppVersion         = "0.1.0"
	DefaultRegistryURL = "https://core-utils.dev/api/registry"
)

var (
	BaseDir       string
	BinDir        string
	ToolsDir      string
	ConfigFile    string
	CacheFile     string
	InstalledFile string
)

func init() {
	if d := os.Getenv("CORE_UTILS_HOME"); d != "" {
		BaseDir = d
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		BaseDir = filepath.Join(home, ".core-utils")
	}
	BinDir = filepath.Join(BaseDir, "bin")
	ToolsDir = filepath.Join(BaseDir, "tools")
	ConfigFile = filepath.Join(BaseDir, "config.json")
	CacheFile = filepath.Join(BaseDir, "registry-cache.json")
	InstalledFile = filepath.Join(BaseDir, "installed.json")
}

// Config holds local CLI settings.
type Config struct {
	RegistryURL string `json:"registry_url,omitempty"`
}

func EnsureDirs() error {
	for _, d := range []string{BaseDir, BinDir, ToolsDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func SaveConfig(c *Config) error {
	if err := os.MkdirAll(BaseDir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile, data, 0644)
}

func GetRegistryURL() string {
	if env := os.Getenv("CORE_UTILS_REGISTRY_URL"); env != "" {
		return env
	}
	cfg, err := LoadConfig()
	if err == nil && cfg.RegistryURL != "" {
		return cfg.RegistryURL
	}
	return DefaultRegistryURL
}
