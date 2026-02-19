package fmt

import (
	"testing"
)

func TestCapitalizeWithMultilineTranslation(t *testing.T) {
	// Register words for test
	RegisterWords([]DictEntry{
		{EN: "Shortcuts"},
		{EN: "Keyboard"},
		{EN: "Switch"},
		{EN: "Fields"},
		{EN: "Edit"},
		{EN: "Execute"},
		{EN: "Cancel"},
		{EN: "Language"},
		{EN: "Supported"},
	})

	tests := []struct {
		name        string
		appName     string
		lang        string
		expected    string
		description string
	}{
		{
			name:        "Simple multiline with Capitalize",
			appName:     "TestApp",
			lang:        "en",
			expected:    "Testapp Shortcuts Keyboard (\"en\"):\n\nTabs:\n  • Tab/Shift+Tab  - Switch Tabs\n\nFields :\n  • Left/Right     - Navigate Fields\n  • Enter          - Edit/Execute\n  • Esc            - Cancel \n\nLanguage Supported : En, Es, Zh, Hi, Ar, Pt, Fr, De, Ru",
			description: "Test that Capitalize preserves multiline structure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSimplifiedHelpContent(tt.appName, tt.lang)
			if result != tt.expected {
				t.Errorf("Test %s failed.\nExpected: %q\nGot:      %q", tt.name, tt.expected, result)
			}
		})
	}
}

func generateSimplifiedHelpContent(appName, lang string) string {
	return Translate(appName, "shortcuts", "keyboard", "(\""+lang+"\"):\n\nTabs:\n  • Tab/Shift+Tab  -", "switch", " tabs\n\n", "fields", ":\n  • Left/Right     - Navigate fields\n  • Enter          - ", "edit", "/", "execute", "\n  • Esc            - ", "cancel", " \n\n", "language", "supported", ": EN, ES, ZH, HI, AR, PT, FR, DE, RU").Capitalize().String()
}
