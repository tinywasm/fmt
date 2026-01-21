package fmt

import "testing"

// TestPrintlnCases contains test cases for Println/Printf
var TestPrintlnCases = []struct {
	name     string
	args     []any
	expected string
}{
	{"single string", []any{"hello"}, "hello"},
	{"multiple strings", []any{"hello", "world"}, "hello world"},
	{"mixed types", []any{"count:", 42}, "count: 42"},
	{"boolean", []any{true, false}, "true false"},
	{"float", []any{"pi:", 3.14}, "pi: 3.14"},
}

var TestPrintfCases = []struct {
	name     string
	format   string
	args     []any
	expected string
}{
	{"string", "Hello %s", []any{"world"}, "Hello world"},
	{"integer", "Count: %d", []any{42}, "Count: 42"},
	{"float", "Pi: %.2f", []any{3.14159}, "Pi: 3.14"},
	{"mixed", "%s has %d items", []any{"User", 5}, "User has 5 items"},
}

// RunPrintlnTest is called by platform-specific tests
func RunPrintlnTest(t *testing.T, captureFunc func(func()) string) {
	for _, tc := range TestPrintlnCases {
		t.Run(tc.name, func(t *testing.T) {
			output := captureFunc(func() {
				Println(tc.args...)
			})
			// Remove trailing newline for comparison
			output = Convert(output).TrimSpace().String()
			if output != tc.expected {
				t.Errorf("Println(%v) = %q, want %q", tc.args, output, tc.expected)
			}
		})
	}
}

// RunPrintfTest is called by platform-specific tests
func RunPrintfTest(t *testing.T, captureFunc func(func()) string) {
	for _, tc := range TestPrintfCases {
		t.Run(tc.name, func(t *testing.T) {
			output := captureFunc(func() {
				Printf(tc.format, tc.args...)
			})
			output = Convert(output).TrimSpace().String()
			if output != tc.expected {
				t.Errorf("Printf(%q, %v) = %q, want %q", tc.format, tc.args, output, tc.expected)
			}
		})
	}
}
