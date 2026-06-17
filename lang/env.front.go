//go:build wasm

package lang

import (
	"syscall/js"
	"github.com/tinywasm/fmt"
)

// getSystemLang detects browser language from navigator.language
func getSystemLang() lang {
	navigator := js.Global().Get("navigator")
	if navigator.IsUndefined() {
		return EN
	}

	language := navigator.Get("language")
	if language.IsUndefined() {
		return EN
	}

	return langParser(language.String())
}

// Println prints arguments to console.log (like fmt.Println)
func Println(args ...any) {
	js.Global().Get("console").Call("log", SmartArgs(fmt.GetConv(), fmt.BuffOut, " ", false, false, args...).String())
}

// Printf prints formatted output to console.log (like fmt.Printf)
func Printf(format string, args ...any) {
	js.Global().Get("console").Call("log", fmt.Sprintf(format, args...))
}
