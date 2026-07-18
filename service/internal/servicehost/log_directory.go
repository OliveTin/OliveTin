package servicehost

import (
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
