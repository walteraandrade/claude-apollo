package config

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	RepoPath   string `toml:"repo_path"`
	SeedDepth  int    `toml:"seed_depth"`
	DebounceMs int    `toml:"debounce_ms"`
}

func ApolloDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".apollo")
}

func ConfigPath() string {
	return filepath.Join(ApolloDir(), "config.toml")
}

func DBPath() string {
	return filepath.Join(ApolloDir(), "apollo.db")
}

func Defaults() Config {
	return Config{
		SeedDepth:  50,
		DebounceMs: 300,
	}
}

func Load() (Config, error) {
	cfg := Defaults()

	data, err := os.ReadFile(ConfigPath())
	if err != nil && !os.IsNotExist(err) {
		return cfg, err
	}

	if err == nil {
		if err := toml.Unmarshal(data, &cfg); err != nil {
			return cfg, err
		}
	}

	applyEnvOverrides(&cfg)
	cfg.RepoPath = ExpandHome(cfg.RepoPath)
	return cfg, nil
}

func Save(cfg Config) error {
	if err := os.MkdirAll(ApolloDir(), 0755); err != nil {
		return err
	}
	data, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigPath(), data, 0600)
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("APOLLO_REPO_PATH"); v != "" {
		cfg.RepoPath = v
	}
	if v := os.Getenv("APOLLO_SEED_DEPTH"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cfg.SeedDepth = n
		}
	}
}

func ExpandHome(path string) string {
	if len(path) == 0 || path[0] != '~' {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
