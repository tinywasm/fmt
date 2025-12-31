//go:build wasm

package fmt

import "syscall/js"

// GetPathBase returns the domain root path using syscall/js.
func GetPathBase() string {
	if global := js.Global(); global.Truthy() {
		if loc := global.Get("location"); loc.Truthy() {
			if origin := loc.Get("origin"); origin.Truthy() {
				return origin.String()
			}
		}
	}
	return "/"
}
