package main

import (
	"os"
	"path/filepath"

	. "github.com/tinywasm/fmt"
)

// BinaryInfo represents information about a compiled binary file
type BinaryInfo struct {
	Name     string
	Size     int64
	SizeStr  string
	Type     string // "native" or "wasm"
	Library  string // "standard" or "tinystring"
	OptLevel string // "default", "ultra", "speed", "debug"
}

// OptimizationConfig represents a TinyGo optimization configuration
type OptimizationConfig struct {
	Name        string
	Flags       string
	Description string
	Suffix      string
}

// FormatSize converts bytes to human-readable format (moved from existing code)
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// FindBinaries searches for binary files in the specified directory
func FindBinaries(dir string, patterns []string) ([]BinaryInfo, error) {
	var binaries []BinaryInfo

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		filename := info.Name()
		for _, pattern := range patterns {
			if Contains(filename, pattern) {
				binary := BinaryInfo{
					Name:     filename,
					Size:     info.Size(),
					SizeStr:  FormatSize(info.Size()),
					OptLevel: extractOptLevel(filename),
				}

				// Determine type and library from filename/path
				if Contains(filename, ".wasm") {
					binary.Type = "wasm"
				} else {
					binary.Type = "native"
				}

				if Contains(path, "standard") {
					binary.Library = "standard"
				} else if Contains(path, "tinystring") {
					binary.Library = "tinystring"
				}

				binaries = append(binaries, binary)
				break
			}
		}

		return nil
	})

	return binaries, err
}

// extractOptLevel extracts optimization level from filename
func extractOptLevel(filename string) string {
	if Contains(filename, "-ultra") {
		return "ultra"
	} else if Contains(filename, "-speed") {
		return "speed"
	} else if Contains(filename, "-debug") {
		return "debug"
	}
	return "default"
}

// LogStep prints a formatted step message
func LogStep(message string) {
	os.Stdout.Write([]byte(Sprintf("📋 %s\n", message)))
}

// LogSuccess prints a formatted success message
func LogSuccess(message string) {
	os.Stdout.Write([]byte(Sprintf("✅ %s\n", message)))
}

// LogError prints a formatted error message
func LogError(message string) {
	os.Stdout.Write([]byte(Sprintf("❌ %s\n", message)))
}

// LogInfo prints a formatted info message
func LogInfo(message string) {
	os.Stdout.Write([]byte(Sprintf("ℹ️  %s\n", message)))
}
