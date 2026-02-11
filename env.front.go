//go:build wasm

package fmt

import (
	"syscall/js"
)

// getSystemLang detects browser language from navigator.language
func (c *Conv) getSystemLang() lang {
	// Get browser language
	navigator := js.Global().Get("navigator")
	if navigator.IsUndefined() {
		return EN
	}

	language := navigator.Get("language")
	if language.IsUndefined() {
		return EN
	}

	// Use the centralized parser.
	return c.langParser(language.String())
}

// Println prints arguments to console.log (like fmt.Println)
func Println(args ...any) {
	js.Global().Get("console").Call("log", GetConv().SmartArgs(BuffOut, " ", false, false, args...).String())
}

// Printf prints formatted output to console.log (like fmt.Printf)
func Printf(format string, args ...any) {
	js.Global().Get("console").Call("log", Sprintf(format, args...))
}

// isWasm reports whether the current binary is compiled for WASM.
// Used for conditional testing.
func isWasm() bool {
	return true
}
