package fmt

import "testing"

func TestCapitalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple Conv",
			input:    "hello world",
			expected: "Hello World",
		},
		{
			name:     "Already capitalized",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Mixed case",
			input:    "hELLo wORLd",
			expected: "Hello World",
		},
		{
			name:     "Extra spaces",
			input:    "  hello   world  ",
			expected: "  Hello   World  ",
		},
		{
			name:     "With numbers",
			input:    "hello 123 world",
			expected: "Hello 123 World",
		},
		{
			name:     "With special characters",
			input:    "héllö wörld",
			expected: "Héllö Wörld",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Single word",
			input:    "hello",
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := Convert(tt.input).Capitalize().String()
			if out != tt.expected {
				t.Errorf("Capitalize() = %q, want %q", out, tt.expected)
			}
		})
	}
}

func TestCapitalizeChaining(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		chain    func(*Conv) *Conv
	}{
		{
			name:     "With Tilde",
			input:    "hólá múndo",
			expected: "Hola Mundo",
			chain: func(Conv *Conv) *Conv {
				return Conv.Tilde().Capitalize()
			},
		},
		{
			name:     "After ToLower",
			input:    "HELLO WORLD",
			expected: "Hello World",
			chain: func(Conv *Conv) *Conv {
				return Conv.ToLower().Capitalize()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := tt.chain(Convert(tt.input)).String()
			if out != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, out, tt.expected)
			}
		})
	}
}

func TestHasUpperPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"Empty string", "", false},
		{"ASCII uppercase", "Hello", true},
		{"ASCII lowercase", "hello", false},
		{"Accented uppercase Á", "Ángel", true},
		{"Accented uppercase É", "Éxito", true},
		{"Accented lowercase á", "ángel", false},
		{"Number first", "123abc", false},
		{"Space first", " Hello", false},
		{"Single uppercase", "A", true},
		{"Single lowercase", "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasUpperPrefix(tt.input)
			if got != tt.expected {
				t.Errorf("HasUpperPrefix(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
