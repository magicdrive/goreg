package core

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func GetModulePath() (string, error) {
	modulePath, err := getModulePathFromGoMod()
	if err == nil {
		return modulePath, nil
	}

	modulePath, err = GetModulePathFromGoList()
	if err == nil {
		return modulePath, nil
	}

	return "", errors.New("failed to get module path: Ensure this is a Go module project (with go.mod)")
}

func getModulePathFromGoMod() (string, error) {
	goModPath, err := findGoModFile()
	if err != nil {
		return "", err
	}

	return extractModulePath(goModPath)
}

func findGoModFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return goModPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

func extractModulePath(goModPath string) (string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}

	return "", os.ErrNotExist
}

var GetModulePathFromGoList = func() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

