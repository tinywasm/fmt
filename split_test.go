package fmt_test

import (
	"testing"

	. "webtyp.com/fmt"
)

func TestStringSplit(t *testing.T) {
	// Test cases with explicit separator
	testCasesWithSeparator := []struct {
		data      string
		separator string
		expected  []string
	}{
		// Original test cases
		{"texto1,texto2", ",", []string{"texto1", "texto2"}},
		{"apple,banana,cherry", ",", []string{"apple", "banana", "cherry"}},
		{"one.two.three.four", ".", []string{"one", "two", "three", "four"}},
		{"hello world", " ", []string{"hello", "world"}},
		{"hello. world", ".", []string{"hello", " world"}},
		{"h.", ".", []string{"h."}},

		// Edge condition test cases
		{"", ",", []string{""}},                                   // Empty string
		{"abc", "", []string{"a", "b", "c"}},                      // Empty separator
		{"ab", ",", []string{"ab"}},                               // String shorter than 3 chars
		{"a", "a", []string{"a"}},                                 // Single char string equals separator
		{"aaa", "a", []string{"", "", "", ""}},                    // All chars are separators
		{"abc,def,", ",", []string{"abc", "def", ""}},             // Trailing separator
		{",abc,def", ",", []string{"", "abc", "def"}},             // Leading separator
		{"abc,,def", ",", []string{"abc", "", "def"}},             // Adjacent separators
		{"abc", "abc", []string{"", ""}},                          // Separator equals data
		{"abcdefghi", "def", []string{"abc", "ghi"}},              // Separator in the middle
		{"abc:::def:::ghi", ":::", []string{"abc", "def", "ghi"}}, // Multi-char separator
	}

	// Test cases for whitespace splitting (no explicit separator)
	testCasesWhitespace := []struct {
		data     string
		expected []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"  hello  world  ", []string{"hello", "world"}},
		{"hello\tworld", []string{"hello", "world"}},
		{"hello\nworld", []string{"hello", "world"}},
		{"hello\rworld", []string{"hello", "world"}},
		{"hello world  test", []string{"hello", "world", "test"}},
		{"", []string{}},                               // Empty string
		{"hello", []string{"hello"}},                   // Single word
		{"   ", []string{}},                            // Only whitespace
		{"hello\n\tworld", []string{"hello", "world"}}, // Mixed whitespace
	}

	// Test with explicit separator
	for _, tc := range testCasesWithSeparator {
		out := Convert(tc.data).Split(tc.separator)

		if !areStringSlicesEqual(out, tc.expected) {
			t.Errorf("Split(%q, %q) = %v; expected %v", tc.data, tc.separator, out, tc.expected)
		}
	}

	// Test with default whitespace separator
	for _, tc := range testCasesWhitespace {
		out := Convert(tc.data).Split()

		if !areStringSlicesEqual(out, tc.expected) {
			t.Errorf("Split(%q) = %v; expected %v", tc.data, out, tc.expected)
		}
	}
}

func areStringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
