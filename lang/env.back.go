//go:build !wasm

package lang

import (
	"os"
	"github.com/tinywasm/fmt"
)

// getSystemLang detects system language from environment variables
func getSystemLang() lang {
	return langParser(
		os.Getenv("LANG"),
		os.Getenv("LANGUAGE"),
		os.Getenv("LC_ALL"),
		os.Getenv("LC_MESSAGES"),
	)
}

// Println prints arguments to stdout followed by newline (like fmt.Println)
func Println(args ...any) {
	os.Stdout.WriteString(SmartArgs(fmt.GetConv(), fmt.BuffOut, " ", false, false, args...).String() + "\n")
}

// Printf prints formatted output to stdout (like fmt.Printf)
func Printf(format string, args ...any) {
	os.Stdout.WriteString(fmt.Sprintf(format, args...))
}
