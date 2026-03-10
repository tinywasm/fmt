package fmt

import "testing"

func TestJSONEscape(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Empty", "", ""},
		{"Plain", "hello world", "hello world"},
		{"Quotes", `hello "world"`, `hello \"world\"`},
		{"Backslash", `a\b`, `a\\b`},
		{"Newlines", "line1\nline2\rline3\tline4", `line1\nline2\rline3\tline4`},
		{"ControlChars", "\x00\x1f", `\u0000\u001f`},
		{"Unicode", "José", "José"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Convert()
			JSONEscape(tt.input, b)
			out := b.String()
			if out != tt.expected {
				t.Errorf("JSONEscape(%q) = %q, want %q", tt.input, out, tt.expected)
			}
		})
	}
}
