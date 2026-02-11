//go:build !wasm

package fmt

import (
	"os"
)

// getSystemLang detects system language from environment variables
func (c *Conv) getSystemLang() lang {
	// Use the centralized parser with common environment variables.
	return c.langParser(
		os.Getenv("LANG"),
		os.Getenv("LANGUAGE"),
		os.Getenv("LC_ALL"),
		os.Getenv("LC_MESSAGES"),
	)
}

// Println prints arguments to stdout followed by newline (like fmt.Println)
func Println(args ...any) {
	os.Stdout.WriteString(GetConv().SmartArgs(BuffOut, " ", false, false, args...).String() + "\n")
}

// Printf prints formatted output to stdout (like fmt.Printf)
func Printf(format string, args ...any) {
	os.Stdout.WriteString(Sprintf(format, args...))
}

// isWasm reports whether the current binary is compiled for WASM.
// Used for conditional testing.
func isWasm() bool {
	return false
}
