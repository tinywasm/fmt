package lang

import (
	"testing"
	"github.com/tinywasm/fmt"
)

func TestHtmlWithTranslation(t *testing.T) {
	// Register custom words for testing
	RegisterWords([]DictEntry{
		{EN: "Hello", ES: "Hola"},
		{EN: "User", ES: "Usuario"},
		{EN: "Hello User", ES: "Hola Usuario"},
		{EN: "Format", ES: "Formato"},
	})

	t.Run("Concatenation with translated word", func(t *testing.T) {
		OutLang(EN)
		got := fmt.Html("<div>", "hello", "</div>").String()
		want := "<div>Hello</div>"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Format string with %L", func(t *testing.T) {
		OutLang(EN)
		got := fmt.Html("<span>%L</span>", "user").String()
		want := "<span>User</span>"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Language ES explicit (via OutLang)", func(t *testing.T) {
		OutLang(ES)
		got := fmt.Html("<div>", "hello", "</div>").String()
		want := "<div>Hola</div>"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Language ES with format (via OutLang)", func(t *testing.T) {
		OutLang(ES)
		got := fmt.Html("<span>%L</span>", "user").String()
		want := "<span>Usuario</span>"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Multiline component with format", func(t *testing.T) {
		OutLang(EN)
		got := fmt.Html(`<div class='container'>
	<h1>%L</h1>
	<p>%v</p>
</div>`, "hello user", 42).String()
		want := `<div class='container'>
	<h1>Hello User</h1>
	<p>42</p>
</div>`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("Multiline component with format ES (via OutLang)", func(t *testing.T) {
		OutLang(ES)
		got := fmt.Html(`<div class='container'>
	<h1>%L</h1>
	<p>%v</p>
</div>`, "hello user", 42).String()
		want := `<div class='container'>
	<h1>Hola Usuario</h1>
	<p>42</p>
</div>`
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}
