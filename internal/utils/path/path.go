package path

import (
	"os"
	"path/filepath"
)

func ProjectRoot() string {
	wd, _ := os.Getwd()
	return filepath.Join(wd, "..")
}
