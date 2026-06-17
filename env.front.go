//go:build wasm

package fmt

import (
	"syscall/js"
)

// Println prints arguments to console.log (like fmt.Println)
func Println(args ...any) {
	c := GetConv()
	for i, arg := range args {
		if i > 0 {
			c.WrString(BuffOut, " ")
		}
		switch v := arg.(type) {
		case string:
			c.WrString(BuffOut, tr(v))
		default:
			c.AnyToBuff(BuffWork, v)
			if c.hasContent(BuffWork) {
				c.WrString(BuffOut, c.GetString(BuffWork))
				c.ResetBuffer(BuffWork)
			}
		}
	}
	js.Global().Get("console").Call("log", c.String())
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
