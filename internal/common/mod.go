package common

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"

	"github.com/magicdrive/goreg/internal/model"
)

func FindGoregToml() (string, error) {
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

func LoadToml(filePath string) (*model.Config, error) {

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cfg *model.Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func LoadConfig() (*model.Config, error) {
	if filePath, err := FindGoregToml(); err != nil {
		cfg := &model.Config{}
		cfg.SetDefaults()
		return cfg, nil
	} else {
		return LoadToml(filePath)
	}
}
