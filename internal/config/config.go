package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	gofsrs "github.com/open-spaced-repetition/go-fsrs/v4"
)

type Config struct {
	DBPath string     `toml:"db_path"`
	FSRS   FSRSConfig `toml:"fsrs"`
}

type FSRSConfig struct {
	RequestRetention float64   `toml:"request_retention"`
	MaximumInterval  float64   `toml:"maximum_interval"`
	EnableFuzz       bool      `toml:"enable_fuzz"`
	EnableShortTerm  bool      `toml:"enable_short_term"`
	Weights          []float64 `toml:"weights"`
}

func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	params := gofsrs.DefaultParam()
	return Config{
		DBPath: filepath.Join(home, ".local", "share", "recurva", "recurva.db"),
		FSRS: FSRSConfig{
			RequestRetention: params.RequestRetention,
			MaximumInterval:  float64(params.MaximumInterval),
			EnableFuzz:       params.EnableFuzz,
			EnableShortTerm:  params.EnableShortTerm,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "recurva")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.toml")
}

func (c *Config) ToFSRSParams() gofsrs.Parameters {
	params := gofsrs.DefaultParam()
	if c.FSRS.RequestRetention > 0 {
		params.RequestRetention = c.FSRS.RequestRetention
	}
	if c.FSRS.MaximumInterval > 0 {
		params.MaximumInterval = c.FSRS.MaximumInterval
	}
	params.EnableFuzz = c.FSRS.EnableFuzz
	params.EnableShortTerm = c.FSRS.EnableShortTerm
	return params
}
