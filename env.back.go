//go:build !wasm

package fmt

import (
	"os"
)

// Println prints arguments to stdout followed by newline (like fmt.Println)
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
	os.Stdout.WriteString(c.String() + "\n")
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
