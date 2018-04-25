package utils

import (
	"os"
)

// FileExists check if file exists
func FileExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}

// Mkdir create directory
func Mkdir(file string) {
	if err := os.MkdirAll(file, 0775); err != nil {
		panic(err)
	}
}

// Remove file/directory
func Remove(path string) {
	if err := os.RemoveAll(path); err != nil {
		panic(err)
	}
}
