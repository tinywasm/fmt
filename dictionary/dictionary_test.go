package dictionary_test

import (
	"testing"

	fmt "github.com/tinywasm/fmt"
	_ "github.com/tinywasm/fmt/dictionary"
)

// TestBuiltinDictLoaded verifies EN and ES are loaded by default.
// Other languages fall back to EN when not registered.
func TestBuiltinDictLoaded(t *testing.T) {
	cases := []struct {
		l    any
		key  string
		want string
	}{
		{fmt.EN, "empty", "Empty"},
		{fmt.ES, "empty", "Vacío"},
		// Other languages fall back to EN (expected behavior with EN/ES-only dict)
		{fmt.FR, "empty", "Empty"},
		{fmt.DE, "empty", "Empty"},
		{fmt.ZH, "empty", "Empty"},
		{fmt.PT, "empty", "Empty"},
		{fmt.RU, "empty", "Empty"},
	}

	for _, tc := range cases {
		fmt.OutLang(tc.l)
		got := fmt.Translate(tc.key).String()
		if got != tc.want {
			t.Errorf("lang %v key %q: want %q got %q", tc.l, tc.key, tc.want, got)
		}
	}
}

func TestDictionaryReSorts(t *testing.T) {
	// Register a word that should come before "all"
	fmt.RegisterWords([]fmt.DictEntry{
		{EN: "aaa"},
	})

	got := fmt.Translate("aaa").String()
	if got != "aaa" {
		t.Errorf("expected 'aaa', got %q", got)
	}
}
