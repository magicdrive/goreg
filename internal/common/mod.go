package common

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/magicdrive/goreg/internal/model"
)

func FindGoregToml() (string, error) {

	if notUseflg := os.Getenv("GOREG_NOT_USE_CONFIGFILE"); notUseflg != "" {
		//lint:ignore ST1005 existing message, kept verbatim
		return "", errors.New("specified env GOREG_NOT_USE_CONFIGFILE.")
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		tomlPath := filepath.Join(dir, "goreg.toml")
		if _, err := os.Stat(tomlPath); err == nil {
			return tomlPath, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(homeDir, ".config", "goreg", "goreg.toml")
	if _, err := os.Stat(configPath); err == nil {
		return configPath, nil
	}

	return "", errors.New("goreg.toml not found")
}

// LoadToml reads the goreg.toml at filePath. Keys it does not set take
// their default values (an empty file is the default configuration).
func LoadToml(filePath string) (*model.Config, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	cfg := &model.Config{}
	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", filePath, err)
	}
	cfg.SetDefaults()

	return cfg, nil
}

// LoadConfig returns the configuration of the goreg.toml FindGoregToml
// finds, or the default configuration when there is none. A goreg.toml that
// exists but cannot be read or is not valid TOML is an error: it is never
// silently replaced by the defaults.
func LoadConfig() (*model.Config, error) {
	filePath, err := FindGoregToml()
	if err != nil {
		cfg := &model.Config{}
		cfg.SetDefaults()
		return cfg, nil
	}
	cfg, err := LoadToml(filePath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}
