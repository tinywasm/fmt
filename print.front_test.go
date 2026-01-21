//go:build wasm

package fmt

import (
	"syscall/js"
	"testing"
)

// testLastOutput stores the last output for testing
var testLastOutput string

// setupTestCapture creates a JavaScript function that captures console.log calls
func setupTestCapture() func() {
	testLastOutput = ""

	// Store original and create wrapper that captures + calls original
	captureFunc := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) > 0 {
			testLastOutput = args[0].String()
		}
		return nil
	})

	js.Global().Set("__testConsoleLog", js.Global().Get("console").Get("log"))
	js.Global().Get("console").Set("log", captureFunc)

	return func() {
		js.Global().Get("console").Set("log", js.Global().Get("__testConsoleLog"))
		captureFunc.Release()
	}
}

func TestPrintln_Frontend(t *testing.T) {
	cleanup := setupTestCapture()
	defer cleanup()

	for _, tc := range TestPrintlnCases {
		t.Run(tc.name, func(t *testing.T) {
			testLastOutput = ""
			Println(tc.args...)
			if testLastOutput != tc.expected {
				t.Errorf("Println(%v) = %q, want %q", tc.args, testLastOutput, tc.expected)
			}
		})
	}
}

func TestPrintf_Frontend(t *testing.T) {
	cleanup := setupTestCapture()
	defer cleanup()

	for _, tc := range TestPrintfCases {
		t.Run(tc.name, func(t *testing.T) {
			testLastOutput = ""
			Printf(tc.format, tc.args...)
			if testLastOutput != tc.expected {
				t.Errorf("Printf(%q, %v) = %q, want %q", tc.format, tc.args, testLastOutput, tc.expected)
			}
		})
	}
}
