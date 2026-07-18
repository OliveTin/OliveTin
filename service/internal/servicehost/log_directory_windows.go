//go:build windows

package servicehost

import (
	"os"
	"path/filepath"
)

func executableDirectory() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(ex), nil
}

func configuredServiceLogDirectory(dir string) (string, error) {
	if dir == "" {
		return "", nil
	}

	exeDir, err := executableDirectory()
	if err != nil {
		return "", err
	}

	return resolveLogDirectory(dir, exeDir), nil
}
