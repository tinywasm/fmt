package fmt

import "testing"

func TestT_LanguageDetection(t *testing.T) {
	// With new API, EN IS the lookup key (case-insensitive)
	// "format" matches {EN: "Format", ES: "Formato"}
	RegisterWords([]DictEntry{
		{EN: "Format", ES: "Formato", FR: "Format"},
		{EN: "Invalid", ES: "Inválido", FR: "Invalide"},
	})

	t.Run("lang constant ES", func(t *testing.T) {
		got := Translate(ES, "format").String()
		if got != "Formato" {
			t.Errorf("expected 'Formato', got '%s'", got)
		}
	})

	t.Run("lang string ES", func(t *testing.T) {
		got := Translate("es", "format").String()
		if got != "Formato" {
			t.Errorf("expected 'Formato', got '%s'", got)
		}
	})

	t.Run("lang constant FR", func(t *testing.T) {
		got := Translate(FR, "format").String()
		if got != "Format" {
			t.Errorf("expected 'Format', got '%s'", got)
		}
	})

	t.Run("lang string FR", func(t *testing.T) {
		got := Translate("FR", "format").String()
		if got != "Format" {
			t.Errorf("expected 'Format', got '%s'", got)
		}
	})

	t.Run("default lang EN", func(t *testing.T) {
		OutLang(EN)
		got := Translate("format").String()
		if got != "Format" {
			t.Errorf("expected 'Format', got '%s'", got)
		}
	})

	// Test phrase composition
	t.Run("phrase ES", func(t *testing.T) {
		got := Translate("ES", "format", "invalid").String()
		if got != "Formato Inválido" {
			t.Errorf("expected 'Formato Inválido', got '%s'", got)
		}
	})

	t.Run("phrase EN", func(t *testing.T) {
		OutLang(EN)
		got := Translate("format", "invalid").String()
		if got != "Format Invalid" {
			t.Errorf("expected 'Format Invalid', got '%s'", got)
		}
	})
}

func TestTranslationFormatting(t *testing.T) {
	RegisterWords([]DictEntry{
		{EN: "Fields", ES: "Campos"},
		{EN: "Cancel", ES: "Cancelar"},
	})

	t.Run("no leading space, custom format", func(t *testing.T) {
		OutLang(EN)
		got := Translate("fields", ":", "cancel", ")").String()
		want := "Fields: Cancel)"
		if got != want {
			t.Errorf("expected '%s', got '%s'", want, got)
		}
	})

	t.Run("no space before colon, phrase with punctuation", func(t *testing.T) {
		got := Translate("format", ":", "invalid").String()
		want := "Format: Invalid"
		if got != want {
			t.Errorf("expected '%s', got '%s'", want, got)
		}
	})

	t.Run("newline with translated field alignment", func(t *testing.T) {
		got := Translate("Tabs:\n", "fields", ":").String()
		want := "Tabs:\nFields:"
		if got != want {
			t.Errorf("expected '%s', got '%s'", want, got)
		}
	})
}

func BenchmarkTranslate(b *testing.B) {
	b.ReportAllocs()
	b.Run("Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := Translate("format")
			c.putConv()
		}
	})
	b.Run("WithLang", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := Translate(ES, "format")
			c.putConv()
		}
	})
	b.Run("Complex", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			c := Translate(ES, "format", ":", "invalid")
			c.putConv()
		}
	})
}
