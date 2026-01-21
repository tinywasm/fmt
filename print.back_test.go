//go:build !wasm

package fmt

import (
	"bytes"
	"os"
	"testing"
)

// captureStdout captures stdout output during function execution
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestPrintln_Backend(t *testing.T) {
	RunPrintlnTest(t, captureStdout)
}

func TestPrintf_Backend(t *testing.T) {
	RunPrintfTest(t, captureStdout)
}
