package fmt_test

import (
	"testing"

	. "webtyp.com/fmt"
)

func TestJoinMethod(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		sep      string
		expected string
	}{
		{
			name:     "Join with default space separator",
			input:    []string{"Hello", "World"},
			sep:      "",
			expected: "Hello World",
		},
		{
			name:     "Join with custom separator",
			input:    []string{"hello", "world", "example"},
			sep:      "-",
			expected: "hello-world-example",
		},
		{
			name:     "Join with semicolon separator",
			input:    []string{"apple", "orange", "banana"},
			sep:      ";",
			expected: "apple;orange;banana",
		},
		{
			name:     "Empty slice",
			input:    []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "Single element",
			input:    []string{"test"},
			sep:      ":",
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out string
			if tt.sep == "" {
				out = Convert(tt.input).Join().String()
			} else {
				out = Convert(tt.input).Join(tt.sep).String()
			}

			if out != tt.expected {
				t.Errorf("Join test %q: expected %q, got %q",
					tt.name, tt.expected, out)
			}
		})
	}
}

func TestJoinChainMethods(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		function func([]string) string
		expected string
	}{
		{
			name:     "Join with ToUpper",
			input:    []string{"hello", "world"},
			expected: "HELLO WORLD",
			function: func(input []string) string {
				return Convert(input).Join().ToUpper().String()
			},
		},
		{
			name:     "Join with custom separator and ToLower",
			input:    []string{"HELLO", "WORLD"},
			expected: "hello-world",
			function: func(input []string) string {
				return Convert(input).Join("-").ToLower().String()
			},
		},
		{
			name:     "Join with Capitalize",
			input:    []string{"hello", "world"},
			expected: "Hello World",
			function: func(input []string) string {
				return Convert(input).Join().Capitalize().String()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tt.function(tt.input)
			if out != tt.expected {
				t.Errorf("Chain test %q: expected %q, got %q",
					tt.name, tt.expected, out)
			}
		})
	}
}
