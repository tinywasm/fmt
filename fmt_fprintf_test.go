package fmt

import (
	"bytes"
	"strings"
	"testing"
)

func TestFprintf(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "simple string",
			format:   "Hello %s",
			args:     []any{"world"},
			expected: "Hello world",
		},
		{
			name:     "integer formatting",
			format:   "Number: %d",
			args:     []any{42},
			expected: "Number: 42",
		},
		{
			name:     "multiple args",
			format:   "Hello %s, you have %d messages",
			args:     []any{"John", 5},
			expected: "Hello John, you have 5 messages",
		},
		{
			name:     "float formatting",
			format:   "Value: %.2f",
			args:     []any{3.14159},
			expected: "Value: 3.14",
		},
		{
			name:     "boolean formatting",
			format:   "Active: %t",
			args:     []any{true},
			expected: "Active: true",
		},
		{
			name:     "hex formatting",
			format:   "Hex: %x",
			args:     []any{255},
			expected: "Hex: ff",
		},
		{
			name:     "no args",
			format:   "Simple text",
			args:     []any{},
			expected: "Simple text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := Fprintf(&buf, tt.format, tt.args...)

			if err != nil {
				t.Errorf("Fprintf() error = %v", err)
				return
			}

			result := buf.String()
			if result != tt.expected {
				t.Errorf("Fprintf() = %q, want %q", result, tt.expected)
			}

			if n != len(tt.expected) {
				t.Errorf("Fprintf() returned n = %d, want %d", n, len(tt.expected))
			}
		})
	}
}

func TestFprintf_Errors(t *testing.T) {
	tests := []struct {
		name   string
		format string
		args   []any
	}{
		{
			name:   "missing argument",
			format: "Hello %s %d",
			args:   []any{"world"}, // Missing second argument
		},
		{
			name:   "invalid format",
			format: "Hello %z", // %z is not supported
			args:   []any{"world"},
		},
		{
			name:   "wrong type for %d",
			format: "Number: %d",
			args:   []any{"not a number"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			n, err := Fprintf(&buf, tt.format, tt.args...)

			if err == nil {
				t.Errorf("Fprintf() expected error but got none, result: %q", buf.String())
			}

			if n != 0 {
				t.Errorf("Fprintf() on error returned n = %d, want 0", n)
			}
		})
	}
}

func TestFprintf_WriterError(t *testing.T) {
	// Test with a writer that always returns an error
	errorWriter := &errorOnlyWriter{}

	n, err := Fprintf(errorWriter, "Hello %s", "world")

	if err == nil {
		t.Error("Fprintf() expected write error but got none")
	}

	if n != 0 {
		t.Errorf("Fprintf() on write error returned n = %d, want 0", n)
	}

	// Error should be from the writer, not from formatting
	if !strings.Contains(err.Error(), "write error") {
		t.Errorf("Fprintf() error = %v, want write error", err)
	}
}

// Helper writer that always returns an error
type errorOnlyWriter struct{}

func (w *errorOnlyWriter) Write(p []byte) (n int, err error) {
	return 0, &writeError{"write error"}
}

type writeError struct {
	msg string
}

func (e *writeError) Error() string {
	return e.msg
}

func BenchmarkFprintf(b *testing.B) {
	var buf bytes.Buffer

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		Fprintf(&buf, "Hello %s, number %d, float %.2f", "world", 42, 3.14159)
	}
}

func BenchmarkFprintf_vs_Fmt(b *testing.B) {
	var buf bytes.Buffer
	format := "Hello %s, number %d, float %.2f"
	args := []any{"world", 42, 3.14159}

	b.Run("Fprintf", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			Fprintf(&buf, format, args...)
		}
	})

	b.Run("Fmt+Write", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			result := Sprintf(format, args...)
			buf.WriteString(result)
		}
	})
}
