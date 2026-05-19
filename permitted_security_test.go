package fmt

import (
	"testing"
)

func TestPermitted_NoHTML_BlocksInjection(t *testing.T) {
	p := Permitted{Letters: true}.NoHTML()

	tests := []struct {
		input   string
		wantErr bool
	}{
		{"hello", false},
		{"hel<lo", true},
		{"hel>lo", true},
		{"hel&lo", true},
		{"hel\"lo", true},
		{"hel'lo", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if err := p.Validate("f", tt.input); (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestPermitted_NoHTML_AllowsNormal(t *testing.T) {
	p := Permitted{Letters: true, Numbers: true}.NoHTML()
	if err := p.Validate("f", "hello123"); err != nil {
		t.Errorf("expected no error for normal text, got %v", err)
	}
}

func TestWidget_StandardPrinciples_RejectsHTML(t *testing.T) {
	// Standard widgets in tinywasm/form/input (Text, Textarea, Email)
	// use Permitted whitelists that don't include <, >, &.
	// We simulate this behavior here.

	textWidgetPermitted := Permitted{Letters: true, Numbers: true, Spaces: true}

	tests := []struct {
		name  string
		input string
	}{
		{"script", "<script>"},
		{"bold", "<b>bold</b>"},
		{"entity", "fish & chips"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := textWidgetPermitted.Validate("f", tt.input); err == nil {
				t.Errorf("expected error for %q due to whitelist, but got nil", tt.input)
			}
		})
	}
}
