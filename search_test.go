package fmt_test

import (
	"testing"

	. "github.com/tinywasm/fmt"
)

func TestCount(t *testing.T) {
	var testCases = map[string]struct {
		Conv     string
		search   string
		expected int
	}{
		"Caso1": {
			Conv:     "Hola, mundo!",
			search:   "mundo",
			expected: 1,
		},
		"Caso2": {
			Conv:     "Hola, mundo!",
			search:   "golang",
			expected: 0,
		},
		"Caso3": {
			Conv:     "Hola, mundo!",
			search:   "",
			expected: 0,
		},
		"Caso4": {
			Conv:     "Hola",
			search:   "Hola, mundo!",
			expected: 0,
		},
		"Caso5": {
			Conv:     "abracadabra",
			search:   "abra",
			expected: 2,
		},
		"Caso6": {
			Conv:     "abracadabra",
			search:   "bra",
			expected: 2,
		},
		"Caso7": {
			Conv:     "abra,cadabra",
			search:   ",",
			expected: 1,
		},
		"Caso8": {
			Conv:     "(abraLol,*?¡¡",
			search:   "Lol",
			expected: 1,
		},
		"Caso9 ": {
			Conv:     "(abraLol,*?¡¡",
			search:   "LoL",
			expected: 0,
		},
		"Caso10 ": {
			Conv:     "(¡ab¡raLol,*?¡¡",
			search:   "¡",
			expected: 4,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := Count(tc.Conv, tc.search)
			if out != tc.expected {
				t.Errorf("Error: Se esperaba %v, pero se obtuvo %v. Texto: %s, Búsqueda: %s", tc.expected, out, tc.Conv, tc.search)
			}
		})
	}
}

func TestContains(t *testing.T) {
	var testCases = map[string]struct {
		Conv     string
		search   string
		expected bool
	}{
		"Encontrado": {
			Conv:     "Hola, mundo!",
			search:   "mundo",
			expected: true,
		},
		"No encontrado": {
			Conv:     "Hola, mundo!",
			search:   "golang",
			expected: false,
		},
		"Búsqueda vacía": {
			Conv:     "Hola, mundo!",
			search:   "",
			expected: false,
		},
		"Texto más corto que búsqueda": {
			Conv:     "Hola",
			search:   "Hola, mundo!",
			expected: false,
		},
		"Múltiples ocurrencias": {
			Conv:     "abracadabra",
			search:   "abra",
			expected: true,
		},
		"Sensible a mayúsculas": {
			Conv:     "(abraLol,*?¡¡",
			search:   "LoL",
			expected: false,
		},
		"Búsqueda de caracteres especiales": {
			Conv:     "(¡ab¡raLol,*?¡¡",
			search:   "¡",
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			out := Contains(tc.Conv, tc.search)
			if out != tc.expected {
				t.Errorf("Error: Se esperaba %v, pero se obtuvo %v. Texto: %s, Búsqueda: %s", tc.expected, out, tc.Conv, tc.search)
			}
		})
	}
}

func TestIndex(t *testing.T) {
	var testCases = map[string]struct {
		s        string
		substr   string
		expected int
	}{
		"Encontrado al inicio": {
			s:        "Hola mundo",
			substr:   "Hola",
			expected: 0,
		},
		"Encontrado en el medio": {
			s:        "Hola mundo",
			substr:   "mundo",
			expected: 5,
		},
		"Encontrado al final": {
			s:        "Hola mundo",
			substr:   "ndo",
			expected: 7,
		},
		"No encontrado": {
			s:        "Hola mundo",
			substr:   "golang",
			expected: -1,
		},
		"Cadena vacía en texto": {
			s:        "Hola mundo",
			substr:   "",
			expected: 0,
		},
		"Cadena vacía en texto vacío": {
			s:        "",
			substr:   "",
			expected: 0,
		},
		"Buscar en cadena vacía": {
			s:        "",
			substr:   "algo",
			expected: -1,
		},
		"Un solo carácter encontrado": {
			s:        "abcdef",
			substr:   "c",
			expected: 2,
		},
		"Un solo carácter no encontrado": {
			s:        "abcdef",
			substr:   "z",
			expected: -1,
		},
		"Carácter nulo encontrado": {
			s:        "abc\x00def",
			substr:   "\x00",
			expected: 3,
		},
		"Carácter nulo no encontrado": {
			s:        "abcdef",
			substr:   "\x00",
			expected: -1,
		},
		"Carácter nulo al inicio": {
			s:        "\x00abcdef",
			substr:   "\x00",
			expected: 0,
		},
		"Carácter nulo al final": {
			s:        "abcdef\x00",
			substr:   "\x00",
			expected: 6,
		},
		"Múltiples caracteres nulos": {
			s:        "abc\x00def\x00ghi",
			substr:   "\x00",
			expected: 3, // Debe encontrar el primero
		},
		"Cadena con bytes de control": {
			s:        "text\x01\x02\x03more",
			substr:   "\x02",
			expected: 5,
		},
		"Subcadena más larga que cadena": {
			s:        "corto",
			substr:   "cadena muy larga",
			expected: -1,
		},
		"Cadena idéntica": {
			s:        "identical",
			substr:   "identical",
			expected: 0,
		},
		"Sensible a mayúsculas": {
			s:        "Hola Mundo",
			substr:   "mundo",
			expected: -1,
		},
		"Caracteres especiales Unicode": {
			s:        "¡Hola! ñoño",
			substr:   "ñoño",
			expected: 8, // ¡ ocupa 2 bytes en UTF-8
		},
		"Primera ocurrencia de múltiples": {
			s:        "abracadabra",
			substr:   "abra",
			expected: 0,
		},
		"Solapamiento parcial": {
			s:        "aaabaaab",
			substr:   "aaab",
			expected: 0,
		},
		"Patrón repetitivo": {
			s:        "abcabcabc",
			substr:   "abc",
			expected: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := Index(tc.s, tc.substr)
			if result != tc.expected {
				t.Errorf("Index(%q, %q) = %d; expected %d",
					tc.s, tc.substr, result, tc.expected)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	cases := map[string]struct {
		content  string
		terms    []string
		expected bool
	}{
		"un término presente":         {"Hello World", []string{"hello"}, true},
		"AND ambos presentes":         {"Hello World", []string{"hello", "world"}, true},
		"AND un término ausente":      {"Hello World", []string{"hello", "xyz"}, false},
		"sin términos":                {"Hello World", []string{}, false},
		"término vacío":               {"Hello World", []string{""}, false},
		"content vacío":               {"", []string{"hello"}, false},
		"mayúsculas normalizadas":     {"GOLANG", []string{"golang"}, true},
		"término vacío entre válidos": {"Hello World", []string{"hello", ""}, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := Matches(tc.content, tc.terms...)
			if got != tc.expected {
				t.Errorf("Matches(%q, %v) = %v; want %v", tc.content, tc.terms, got, tc.expected)
			}
		})
	}
}

func TestMatchesAny(t *testing.T) {
	cases := map[string]struct {
		content  string
		terms    []string
		expected bool
	}{
		"OR uno presente": {"Hello World", []string{"hello", "xyz"}, true},
		"OR ninguno":      {"Hello World", []string{"foo", "bar"}, false},
		"sin términos":    {"Hello World", []string{}, false},
		"término vacío":   {"Hello World", []string{""}, false},
		"content vacío":   {"", []string{"hello"}, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := MatchesAny(tc.content, tc.terms...)
			if got != tc.expected {
				t.Errorf("MatchesAny(%q, %v) = %v; want %v", tc.content, tc.terms, got, tc.expected)
			}
		})
	}
}
