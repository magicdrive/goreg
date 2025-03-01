package configfile

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// findGoregToml searches for "goreg.toml" from the current directory up to the home directory.
// If not found, it checks "~/.config/goreg/goreg.toml".
func findGoregToml() (string, error) {
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Traverse upwards to find goreg.toml
	for {
		tomlPath := filepath.Join(dir, "goreg.toml")
		if _, err := os.Stat(tomlPath); err == nil {
			return tomlPath, nil
		}

		// Stop if we reach the root directory
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	// Check ~/.config/goreg/goreg.toml
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

// loadToml loads a TOML file into a map using go-toml/v2.
func loadToml(filePath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config, nil
}

