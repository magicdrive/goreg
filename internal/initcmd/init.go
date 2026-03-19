package initcmd

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed default_goreg.toml
var defaultTomlContent string

func Execute() error {
	filename := "goreg.toml"

	// Check if goreg.toml already exists
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("%s already exists in current directory", filename)
	}

	// Create goreg.toml with default content
	fmt.Printf("Creating %s in current directory...\n", filename)
	if err := os.WriteFile(filename, []byte(defaultTomlContent), 0644); err != nil {
		return fmt.Errorf("failed to create %s: %w", filename, err)
	}

	fmt.Printf("%s created successfully.\n", filename)
	return nil
}
