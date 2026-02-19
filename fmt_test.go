package fmt

import "testing"

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "Fmt integer with thousand separators (EU)",
			input: 2189009,
			want:  "2.189.009",
		},
		{
			name:  "Fmt decimal number with trailing zeros (EU)",
			input: 2189009.00,
			want:  "2.189.009",
		},
		{
			name:  "Fmt decimal number (EU)",
			input: 2189009.123,
			want:  "2.189.009,123",
		},
		{
			name:  "Fmt string number (EU)",
			input: "2189009.00",
			want:  "2.189.009",
		},
		{
			name:  "Fmt negative number (EU)",
			input: -2189009,
			want:  "-2.189.009",
		},
		{
			name:  "Fmt small number",
			input: 123,
			want:  "123",
		},
		{
			name:  "Fmt zero",
			input: 0,
			want:  "0",
		},
		{
			name:  "Non-numeric input",
			input: "hello",
			want:  "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := Convert(tt.input).Thousands().String()
			if out != tt.want {
				t.Errorf("Thousands() got = %v, want %v", out, tt.want)
			}
		})
	}

	// Anglo format tests
	t.Run("Fmt integer with thousand separators (Anglo)", func(t *testing.T) {
		out := Convert(2189009).Thousands(true).String()
		if out != "2,189,009" {
			t.Errorf("Thousands(true) got = %v, want %v", out, "2,189,009")
		}
	})
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{
			name:     "String formatting",
			format:   "Hello %s!",
			args:     []any{"World"},
			expected: "Hello World!",
		},
		{
			name:     "Integer formatting",
			format:   "Value: %d",
			args:     []any{42},
			expected: "Value: 42",
		},
		{
			name:     "Float formatting",
			format:   "Pi: %.2f",
			args:     []any{3.14159},
			expected: "Pi: 3.14",
		},
		{
			name:     "Multiple arguments",
			format:   "Hello %s, you have %d messages",
			args:     []any{"Alice", 5},
			expected: "Hello Alice, you have 5 messages",
		},
		{
			name:     "Binary formatting",
			format:   "Binary: %b",
			args:     []any{7},
			expected: "Binary: 111",
		},
		{
			name:     "Hexadecimal formatting",
			format:   "Hex: %x",
			args:     []any{255},
			expected: "Hex: ff",
		},
		{
			name:     "Octal formatting",
			format:   "Octal: %o",
			args:     []any{64},
			expected: "Octal: 100",
		},
		{
			name:     "Value formatting",
			format:   "Bool: %v",
			args:     []any{true},
			expected: "Bool: true",
		},
		{
			name:     "Percent sign",
			format:   "100%% complete",
			args:     []any{},
			expected: "100% complete",
		},
		{
			name:     "Missing argument",
			format:   "Value: %d",
			args:     []any{},
			expected: "",
		},
		{
			name:     "Boolean true formatting",
			format:   "Bool: %t",
			args:     []any{true},
			expected: "Bool: true",
		},
		{
			name:     "Quoted string formatting",
			format:   "Quoted: %q",
			args:     []any{"hello"},
			expected: "Quoted: \"hello\"",
		},
		{
			name:     "Pointer formatting",
			format:   "Pointer: %p",
			args:     []any{new(int)},
			expected: "Pointer: 0x",
		},
		{
			name:     "Unicode format",
			format:   "Unicode: %U",
			args:     []any{'A'},
			expected: "Unicode: U+0041",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := Sprintf(test.format, test.args...)
			if out != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, out)
			}
		})
	}
}

func TestFormatLocalized(t *testing.T) {
	// Register custom word for testing
	RegisterWords([]DictEntry{
		{EN: "Hello", ES: "Hola"},
	})

	// Save original language to restore later
	origLang := OutLang()
	defer OutLang(origLang)

	// Test default language (EN)
	OutLang(EN)
	if got := Sprintf("%L", "hello"); got != "Hello" {
		t.Errorf("Sprintf(%%L) EN = %q, want %q", got, "Hello")
	}

	// Test Spanish
	OutLang(ES)
	if got := Sprintf("%L", "hello"); got != "Hola" {
		t.Errorf("Sprintf(%%L) ES = %q, want %q", got, "Hola")
	}

	// Test mixed
	OutLang(EN)
	if got := Sprintf("Say %L world", "hello"); got != "Say Hello world" {
		t.Errorf("Sprintf mixed = %q, want %q", got, "Say Hello world")
	}
}

func TestErrorFormatting(t *testing.T) {
	OutLang(EN)
	err := Err("file not found").Error()
	want := "file not found"
	if err != want {
		t.Errorf("Err want %q got %q", want, err)
	}

	// Verify Err returns Conv that implements error
	var _ error = Err("some error")
}
