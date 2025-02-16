package core

import (
	"log"
	"os/exec"
	"strings"
)

func GetModulePath() string {
	cmd := exec.Command("go", "list", "-m")
	out, err := cmd.Output()
	if err != nil {
		log.Fatal("Error: Failed to get module path. Ensure this is a Go module project (with go.mod).")
	}
	_modulePath := strings.TrimSpace(string(out))
	return _modulePath
}
