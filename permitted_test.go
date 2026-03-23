package fmt

import (
	"strings"
	"testing"
)

func TestPermittedValidate(t *testing.T) {
	tests := []struct {
		name    string
		p       Permitted
		text    string
		wantErr bool
	}{
		// Length (with Letters: true to pass character check)
		{"Min length pass", Permitted{Minimum: 3, Letters: true}, "abc", false},
		{"Min length fail", Permitted{Minimum: 3, Letters: true}, "ab", true},
		{"Max length pass", Permitted{Maximum: 5, Letters: true}, "abcde", false},
		{"Max length fail", Permitted{Maximum: 5, Letters: true}, "abcdef", true},
		{"Unicode count", Permitted{Minimum: 2, Maximum: 2, Letters: true}, "ññ", false},

		// Character classes
		{"Letters only pass", Permitted{Letters: true}, "abcABCñÑ", false},
		{"Letters only fail", Permitted{Letters: true}, "abc1", true},
		{"Numbers only pass", Permitted{Numbers: true}, "123", false},
		{"Numbers only fail", Permitted{Numbers: true}, "123a", true},
		{"Tilde pass", Permitted{Tilde: true}, "áéíóúÁÉÍÓÚ", false},
		{"Tilde fail", Permitted{Tilde: true}, "abc", true},
		{"Spaces pass", Permitted{Spaces: true}, "   ", false},
		{"Spaces fail with non-space", Permitted{Spaces: true}, " a", true},

		// Whitespace more cases
		{"Breakline pass", Permitted{BreakLine: true}, "\n\n", false},
		{"Breakline fail", Permitted{Letters: true}, "\n", true},
		{"Tab pass", Permitted{Tab: true}, "\t", false},
		{"Tab fail", Permitted{Letters: true}, "\t", true},

		// Extra and Forbidden
		{"Extra pass", Permitted{Extra: []rune{'@', '.'}, Letters: true}, "alice@example.com", false},
		{"Extra fail", Permitted{Extra: []rune{'@'}, Letters: true}, "alice.com", true},
		{"Not allowed fail", Permitted{Letters: true, NotAllowed: []string{"bad"}}, "this is bad", true},
		{"Not allowed pass", Permitted{Letters: true, NotAllowed: []string{"bad"}}, "good", false},

		// StartWith
		{"StartWith pass", Permitted{Letters: true, StartWith: &Permitted{Letters: true}}, "Alice", false},
		{"StartWith fail", Permitted{Letters: true, StartWith: &Permitted{Numbers: true}}, "Alice", true},
		{"StartWith empty string", Permitted{Letters: true, StartWith: &Permitted{Letters: true}}, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Validate("field", tt.text); (err != nil) != tt.wantErr {
				t.Errorf("Permitted.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPermittedErrorMessages(t *testing.T) {
	tests := []struct {
		name     string
		p        Permitted
		text     string
		contains string
	}{
		{"Space error", Permitted{Letters: true}, "a b", "space not allowed"},
		{"Tab error", Permitted{Letters: true}, "a\tb", "tab not allowed"},
		{"Newline error", Permitted{Letters: true}, "a\nb", "new line not allowed"},
		{"Char error", Permitted{Letters: true}, "a1b", "character not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.p.Validate("field", tt.text)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			msg := err.Error()
			if !strings.Contains(msg, tt.contains) {
				t.Errorf("expected error message to contain %q, got %q", tt.contains, msg)
			}
		})
	}
}

func TestPermittedRecursiveStartWith(t *testing.T) {
	// StartWith should only check the first rune
	p := Permitted{
		Letters: true,
		StartWith: &Permitted{
			Letters: true,
		},
	}
	if err := p.Validate("name", "Alice"); err != nil {
		t.Errorf("expected success, got %v", err)
	}

	p.StartWith.Numbers = true
	p.StartWith.Letters = false
	if err := p.Validate("name", "Alice"); err == nil {
		t.Error("expected failure because 'A' is not a number")
	}
}
