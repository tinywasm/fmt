package fmt_test

import (
	"testing"

	. "webtyp.com/fmt"
)

func TestDebugReplaceEmpty(t *testing.T) {
	// Test case: Convert("").Replace("", "123") should be ""

	result := Convert("").Replace("", "123").String()
	t.Logf("Convert(\"\").Replace(\"\", \"123\").String(): %q", result)

	if result != "" {
		t.Errorf("Expected empty string, got %q", result)
	}
}

func TestStringOperations(t *testing.T) {
	t.Run("Replace", func(t *testing.T) {
		tests := []struct {
			input    string
			old      any
			newStr   any
			n        int
			expected string
		}{
			{"Este es un ejemplo de texto de prueba.", "ejemplo", "cambio", -1, "Este es un cambio de texto de prueba."},
			{"Hola mundo!", "mundo", "Gophers", -1, "Hola Gophers!"},
			{"abc abc abc", "abc", "123", -1, "123 123 123"},
			{"abc", "xyz", "123", -1, "abc"},
			{"", "", "123", -1, ""},
			{"abcdabcdabcd", "cd", "12", -1, "ab12ab12ab12"},
			{"palabra, punto,", ",", ".", -1, "palabra. punto."},
			// Pruebas con tipos diferentes de any
			{"Test 123 value", 123, 456, -1, "Test 456 value"},
			{"Boolean true in Conv", true, false, -1, "Boolean false in Conv"},
			{"Pi is 3.14159", 3.14159, 3.142, -1, "Pi is 3.142"},
			// Pruebas con límite de reemplazos
			{"abc abc abc", "abc", "123", 1, "123 abc abc"},
			{"abc abc abc", "abc", "123", 2, "123 123 abc"},
			{"abc abc abc", "abc", "123", 0, "abc abc abc"},
		}

		for _, test := range tests {
			var out string
			if test.n >= 0 {
				out = Convert(test.input).Replace(test.old, test.newStr, test.n).String()
			} else {
				out = Convert(test.input).Replace(test.old, test.newStr).String()
			}

			if out != test.expected {
				t.Errorf("Para input '%s', old '%v', new '%v', n '%d', esperado '%s', pero obtenido '%s'",
					test.input, test.old, test.newStr, test.n, test.expected, out)
			}
		}
	})

	t.Run("TrimSuffix", func(t *testing.T) {
		tests := []struct {
			input, suffix, expected string
		}{
			{"hello.txt", ".txt", "hello"},
			{"example", "123", "example"},
			{"file.txt.txt", ".txt", "file.txt"},
			{"", "", ""},
			{"abc", "xyz", "abc"},
			{"mi_directorio\\cmd", "\\cmd", "mi_directorio"},
		}

		for _, test := range tests {
			out := Convert(test.input).TrimSuffix(test.suffix).String()
			if out != test.expected {
				t.Errorf("Para input '%s', suffix '%s', esperado '%s', pero obtenido '%s'", test.input, test.suffix, test.expected, out)
			}
		}
	})

	t.Run("TrimPrefix", func(t *testing.T) {
		tests := []struct {
			input, prefix, expected string
		}{
			{"prefix-hello", "prefix-", "hello"},
			{"example", "123", "example"},
			{"txt.file", "txt.", "file"},
			{"", "", ""},
			{"abc", "xyz", "abc"},
		}

		for _, test := range tests {
			out := Convert(test.input).TrimPrefix(test.prefix).String()
			if out != test.expected {
				t.Errorf("Para input '%s', prefix '%s', esperado '%s', pero obtenido '%s'", test.input, test.prefix, test.expected, out)
			}
		}
	})

	t.Run("TrimSpace", func(t *testing.T) {
		tests := []struct {
			input, expected string
		}{
			{"  hello world  ", "hello world"},
			{"abc123", "abc123"},
			{"  trim me  ", "trim me"},
			{"", ""},
			{"  ", ""},
			{"    mucho espacio\n\n\t\tcon salto\n\n\t\tde linea     \n\t\t\t\t\t\t\n\t\t\t\t", "mucho espacio\n\n\t\tcon salto\n\n\t\tde linea"},
			{`    mucho espacio
		
		con salto

		de linea     
		              
		
		`, `mucho espacio
		
		con salto

		de linea`},
		}

		for _, test := range tests {
			out := Convert(test.input).TrimSpace().String()
			if out != test.expected {
				t.Errorf("Para input '%s', esperado '%s', pero obtenido '%s'", test.input, test.expected, out)
			}
		}
	})

	t.Run("ChainableMethods", func(t *testing.T) {
		tests := []struct {
			name     string
			input    string
			expected string
			chain    func(input string) string
		}{
			{
				name:     "Replace and TrimSpace",
				input:    "  hello world  ",
				expected: "hello universe",
				chain: func(input string) string {
					return Convert(input).TrimSpace().Replace("world", "universe").String()
				},
			},
			{
				name:     "Replace, TrimSpace, and TrimSuffix",
				input:    "  filename.txt  ",
				expected: "file",
				chain: func(input string) string {
					return Convert(input).TrimSpace().Replace("name", "").TrimSuffix(".txt").String()
				},
			},
			{
				name:     "Multiple Replaces",
				input:    "replace multiple words in this Conv",
				expected: "change many terms in this content",
				chain: func(input string) string {
					return Convert(input).
						Replace("replace", "change").
						Replace("multiple", "many").
						Replace("words", "terms").
						Replace("Conv", "content").
						String()
				},
			},
			{
				name:     "TrimPrefix and TrimSuffix",
				input:    "prefix-content.suffix",
				expected: "content",
				chain: func(input string) string {
					return Convert(input).TrimPrefix("prefix-").TrimSuffix(".suffix").String()
				},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				out := test.chain(test.input)
				if out != test.expected {
					t.Errorf("Chain test '%s': expected '%s', got '%s'", test.name, test.expected, out)
				}
			})
		}
	})
}
