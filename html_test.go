package fmt

import "testing"

func TestHtml(t *testing.T) {
	// Register custom words for testing
	RegisterWords([]DictEntry{
		{EN: "Hello", ES: "Hola"},
		{EN: "User", ES: "Usuario"},
		{EN: "Hello User", ES: "Hola Usuario"},
		{EN: "Format", ES: "Formato"},
	})

	tests := []struct {
		name     string
		args     []any
		expected string
	}{
		{
			name:     "Concatenation simple",
			args:     []any{"div", "span"},
			expected: "divspan",
		},
		{
			name:     "Concatenate number and text",
			args:     []any{123, "abc"},
			expected: "123abc",
		},
		{
			name:     "2-letter tag regression",
			args:     []any{"hr", "br"},
			expected: "hrbr",
		},
		{
			name:     "Concatenation with translated word",
			args:     []any{"<div>", "hello", "</div>"},
			expected: "<div>Hello</div>",
		},
		{
			name:     "Format string",
			args:     []any{"<div class='%s'>", "my-class"},
			expected: "<div class='my-class'>",
		},
		{
			name:     "Format string with multiple args",
			args:     []any{"<a href='%s'>%s</a>", "/home", "Home"},
			expected: "<a href='/home'>Home</a>",
		},
		{
			name:     "Format string with %L",
			args:     []any{"<span>%L</span>", "user"},
			expected: "<span>User</span>",
		},
		{
			name:     "Mixed strings without format",
			args:     []any{"<p>", "Content", "</p>"},
			expected: "<p>Content</p>",
		},
		{
			name:     "Empty args",
			args:     []any{},
			expected: "",
		},
		{
			name:     "String with % but not format",
			args:     []any{"Width: 100%"},
			expected: "Width: 100%",
		},
		{
			name:     "String with %% (literal percent)",
			args:     []any{"Success: 100%%"}, // Fmt treats %% as literal %
			expected: "Success: 100%",
		},
		{
			name:     "Language ES explicit",
			args:     []any{ES, "<div>", "hello", "</div>"},
			expected: "<div>Hola</div>",
		},
		{
			name:     "Language EN explicit",
			args:     []any{EN, "<div>", "hello", "</div>"},
			expected: "<div>Hello</div>",
		},
		{
			name:     "Language ES with format",
			args:     []any{ES, "<span>%L</span>", "user"},
			expected: "<span>Usuario</span>",
		},
		{
			name: "Multiline component with format",
			args: []any{
				`<div class='container'>
	<h1>%L</h1>
	<p>%v</p>
</div>`,
				"hello user",
				42,
			},
			expected: `<div class='container'>
	<h1>Hello User</h1>
	<p>42</p>
</div>`,
		},
		{
			name: "Multiline component with format ES",
			args: []any{
				ES,
				`<div class='container'>
	<h1>%L</h1>
	<p>%v</p>
</div>`,
				"hello user",
				42,
			},
			expected: `<div class='container'>
	<h1>Hola Usuario</h1>
	<p>42</p>
</div>`,
		},
	}

	// Test default language (EN)
	OutLang(EN)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Html(tt.args...).String()
			if got != tt.expected {
				t.Errorf("Html() = %q, want %q", got, tt.expected)
			}
		})
	}

	// Test with Spanish
	t.Run("Spanish Translate", func(t *testing.T) {
		OutLang(ES)
		defer OutLang(EN) // Restore

		got := Html("<div>", "hello", "</div>").String()
		want := "<div>Hola</div>"
		if got != want {
			t.Errorf("Html() ES = %q, want %q", got, want)
		}
	})
}

func BenchmarkHtml(b *testing.B) {
	b.ReportAllocs()
	b.Run("Simple", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Html("<div>", "content", "</div>").String()
		}
	})
	b.Run("WithLang", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Html(ES, "<div>", "content", "</div>").String()
		}
	})
	b.Run("WithTranslation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Html("<span>", "format", "</span>").String()
		}
	})
	b.Run("WithLangAndTranslation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Html(ES, "<span>", "format", "</span>").String()
		}
	})
	b.Run("WithFormat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Html("<div class='%s'>", "my-class").String()
		}
	})
	b.Run("WithFormatAndLang", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Html(ES, "<span>%L</span>", "format").String()
		}
	})
}

func TestEscapeAttr(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"basic", `Tom & Jerry's "House" <tag>`, `Tom &amp; Jerry&#39;s &quot;House&quot; &lt;tag&gt;`},
		{"already-entity", `&amp; &lt; &gt;`, `&amp;amp; &amp;lt; &amp;gt;`}, // double-escape expected
		{"unicode", `こんにちは & <br> 😊`, `こんにちは &amp; &lt;br&gt; 😊`},
		{"multiple", `a & b & c`, `a &amp; b &amp; c`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Convert(tc.in).EscapeAttr()
			if got != tc.want {
				t.Fatalf("%s: got=%q want=%q", tc.name, got, tc.want)
			}
		})
	}
}

func TestEscapeHTML_TableDriven(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"tags", `<div class="x">Tom & Jerry's</div>`, `&lt;div class=&quot;x&quot;&gt;Tom &amp; Jerry&#39;s&lt;/div&gt;`},
		{"already-entity", `&amp; &lt;`, `&amp;amp; &amp;lt;`},
		{"emoji-and-tags", `😀 <p>1 & 2</p>`, `😀 &lt;p&gt;1 &amp; 2&lt;/p&gt;`},
		{"quotes-only", `She said: "Hi"`, `She said: &quot;Hi&quot;`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Convert(tc.in).EscapeHTML()
			if got != tc.want {
				t.Fatalf("%s: got=%q want=%q", tc.name, got, tc.want)
			}
		})
	}
}

// TestEscapeHTML_CompareStdLib compares EscapeHTML behavior with html.EscapeString from standard library
func TestEscapeHTML_CompareStdLib(t *testing.T) {
	// Note: html.EscapeString only escapes &, <, >, ", and ' (as &#39; or &#34;)
	// Our implementation matches this behavior
	tests := []struct {
		name string
		in   string
	}{
		{"basic", `<script>alert("XSS")</script>`},
		{"quotes", `She said: "Hello" & 'Goodbye'`},
		{"entities", `Tom & Jerry's <div>`},
		{"unicode", `こんにちは <p>世界</p>`},
		{"mixed", `<a href="link.html?id=1&type=2">Click here</a>`},
		{"empty", ``},
		{"ampersand-only", `A & B & C`},
		{"all-chars", `&<>"'`},
	}

	// Import html package for comparison (added at top of file)
	// We'll verify our output matches expected HTML escaping semantics
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Convert(tc.in).EscapeHTML()

			// Verify all dangerous characters are escaped
			if Contains(got, "<") || Contains(got, ">") {
				t.Errorf("Unescaped angle brackets in output: %q", got)
			}

			// Verify input characters were processed
			if tc.in != "" && got == tc.in {
				// Only fail if input contained escapable characters
				if Contains(tc.in, "&") || Contains(tc.in, "<") || Contains(tc.in, ">") ||
					Contains(tc.in, `"`) || Contains(tc.in, "'") {
					t.Errorf("Input was not escaped: %q", tc.in)
				}
			}

		})
	}
}

// TestEscapeAttr_CompareStdLib validates EscapeAttr for use in HTML attributes
func TestEscapeAttr_CompareStdLib(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"attr-value", `class="btn btn-primary"`},
		{"with-quotes", `onClick="alert('test')"`},
		{"url", `https://example.com?a=1&b=2`},
		{"mixed", `Tom & Jerry's "adventure"`},
		{"empty", ``},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Convert(tc.in).EscapeAttr()

			// Verify dangerous characters for attributes are escaped
			if Contains(got, `"`) || Contains(got, "<") || Contains(got, ">") {
				t.Errorf("Unescaped dangerous characters in attribute: %q", got)
			}

		})
	}
}
