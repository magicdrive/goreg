package core

import (
	"log"
	"os/exec"
	"strings"
)

var ModulePath string

func GetModulePath() string {
	if ModulePath != "" {
		return ModulePath
	}
	cmd := exec.Command("go", "list", "-m")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal("Error: Failed to get module path. Ensure this is a Go module project (with go.mod).")
	}
	_modulePath := strings.TrimSpace(string(out))
	return _modulePath
}
