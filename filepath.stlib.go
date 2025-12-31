//go:build !wasm

package fmt

import (
	"os"
)

// GetPathBase returns the base path using os.Getwd().
func GetPathBase() string {
	if wd, err := os.Getwd(); err == nil {
		cleaned, _ := pathClean(wd)
		return cleaned
	}
	return ""
}
