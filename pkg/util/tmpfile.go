package util

import (
	"path/filepath"
	"os"
	"runtime"
)

//fix issue with docker needing mounted files to exist within home dir
func UnikTmpDir() string {
	if runtime.GOOS == "darwin" {
		return filepath.Join(os.Getenv("HOME"), ".unik")
	}
	return ""
}
