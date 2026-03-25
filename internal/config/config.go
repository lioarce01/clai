package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	API     APIConfig     `toml:"api"`
	Model   ModelConfig   `toml:"model"`
	UI      UIConfig      `toml:"ui"`
	Storage StorageConfig `toml:"storage"`
}

type APIConfig struct {
	APIKey  string `toml:"api_key"`
	BaseURL string `toml:"base_url"`
}

type ModelConfig struct {
	Name         string  `toml:"name"`
	Temperature  float64 `toml:"temperature"`
	MaxTokens    int     `toml:"max_tokens"`
	TopP         float64 `toml:"top_p"`
	SystemPrompt string  `toml:"system_prompt"`
}

type UIConfig struct {
	Theme        string `toml:"theme"`
	MouseEnabled bool   `toml:"mouse_enabled"`
}

type StorageConfig struct {
	DataDir string `toml:"data_dir"`
}

func defaults() Config {
	return Config{
		API: APIConfig{
			BaseURL: "https://api.openai.com/v1",
		},
		Model: ModelConfig{
			Name:         "gpt-4o",
			Temperature:  0.7,
			MaxTokens:    4096,
			TopP:         1.0,
			SystemPrompt: "You are a helpful assistant.",
		},
		UI: UIConfig{
			Theme:        "dark",
			MouseEnabled: true,
		},
	}
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "clai", "config.toml"), nil
}

func dataDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "clai"), nil
}

// Load reads config from disk, applying env var overrides and defaults.
func Load() (*Config, error) {
	cfg := defaults()

	path, err := configPath()
	if err != nil {
		return nil, err
	}

	// Load from file if it exists
	if _, err := os.Stat(path); err == nil {
		if _, err := toml.DecodeFile(path, &cfg); err != nil {
			return nil, err
		}
	}

	// Apply env var overrides
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.API.APIKey = v
	}
	if v := os.Getenv("OPENAI_BASE_URL"); v != "" {
		cfg.API.BaseURL = v
	}

	// Set data dir if not configured
	if cfg.Storage.DataDir == "" {
		dir, err := dataDir()
		if err != nil {
			return nil, err
		}
		cfg.Storage.DataDir = dir
	}

	return &cfg, nil
}

// Save persists the config to disk.
func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(cfg)
}

// Validate checks that required fields are set.
func Validate(cfg *Config) error {
	if cfg.API.APIKey == "" {
		return errors.New("API key is required. Set OPENAI_API_KEY env var or configure it in settings (Ctrl+O)")
	}
	if cfg.API.BaseURL == "" {
		return errors.New("API base URL is required")
	}
	return nil
}
