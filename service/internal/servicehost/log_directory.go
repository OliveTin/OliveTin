package servicehost

import (
	"os"
	"path/filepath"
)

func resolveLogDirectory(dir string, baseDir string) string {
	if dir == "" {
		return ""
	}

	if filepath.IsAbs(dir) {
		return dir
	}

	if baseDir == "" {
		return dir
	}

	return filepath.Join(baseDir, dir)
}

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
