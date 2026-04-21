package fmt

import "testing"

// TestSprintfQ_EscapesSpecialChars verifica que %q en Sprintf escapa correctamente
// los caracteres especiales dentro del string.
//
// BUG CONOCIDO: la implementación actual en fmt_template.go hace:
//
//	return "\"" + v + "\""
//
// lo que produce JSON inválido cuando el string contiene comillas.
// El fix correcto es reutilizar Convert(v).Quote() que ya existe en quote.go.
func TestSprintfQ_EscapesSpecialChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain string",
			input:    "hello",
			expected: `"hello"`,
		},
		{
			name:     "double quotes must be escaped",
			input:    `{"key":"value"}`,
			expected: `"{\"key\":\"value\"}"`,
		},
		{
			name:     "backslash must be escaped",
			input:    `path\file`,
			expected: `"path\\file"`,
		},
		{
			name:     "newline must be escaped",
			input:    "line1\nline2",
			expected: `"line1\nline2"`,
		},
		{
			name:     "tab must be escaped",
			input:    "a\tb",
			expected: `"a\tb"`,
		},
		{
			// Caso real que causó el footer vacío en la TUI:
			// daemon.go usaba fmt.Sprintf(`...,"result":%q`, stateJSON)
			// donde stateJSON es un array JSON con comillas internas.
			name:  "json array payload (real MCP state bug)",
			input: `[{"tab_title":"BUILD","handler_name":"WasmClient","handler_type":1}]`,
			expected: `"[{\"tab_title\":\"BUILD\",\"handler_name\":\"WasmClient\",\"handler_type\":1}]"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sprintf("%q", tt.input)
			if got != tt.expected {
				t.Errorf("Sprintf(%%q, %q)\n got:  %s\n want: %s", tt.input, got, tt.expected)
			}
		})
	}
}
