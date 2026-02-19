# HTML Generation & Escaping

fmt provides utilities for generating and escaping HTML content safely and efficiently.

## HTML Generation

The `Html` function creates HTML strings with support for concatenation, formatting, and localization.

### Features

- **Concatenation**: Joins arguments without spaces (unlike `Translate` or `Err`).
- **Formatting**: Supports `printf`-style formatting if the first argument is a format string.
- **Localization**: Supports `LocStr` translation and explicit language selection.

### Usage

```go
// 1. Simple Concatenation
Html("<div>", "content", "</div>").String()
// -> "<div>content</div>"

// 2. Formatting (printf-style)
Html("<div class='%s'>", "my-class").String()
// -> "<div class='my-class'>"

// 3. Localization (using dictionary keys)
// Note: requires import _ "github.com/tinywasm/fmt/dictionary"
Html("<span>", "user", "</span>").String()
// -> "<span>User</span>" (EN)
// -> "<span>Usuario</span>" (ES)

// 4. Explicit Language Selection
// Pass language as first argument (lang constant)
Html(ES, "<div>", "hello", "</div>").String()
// -> "<div>Hola</div>"

// 5. Zero-allocation (pooled Conv)
c := Html("<span>", "format", "</span>")
defer c.PutConv() // Manual cleanup if not calling .String()
html := c.String() // Auto-releases to pool

// 6. Multiline Component (using format specifiers)
Html(`<div class='container'>
	<h1>%L</h1>
	<p>%v</p>
</div>`, "hello", 42).String()
// -> "<div class='container'>
//     	<h1>Hello</h1>
//     	<p>42</p>
//     </div>"

// 7. Multiline Component with explicit language
Html(ES, `<div class='container'>
	<h1>%L</h1>
	<p>%v</p>
</div>`, "hello", 42).String()
// -> "<div class='container'>
//     	<h1>Hola</h1>
//     	<p>42</p>
//     </div>"
```

## HTML Escaping

fmt provides two convenience helpers to escape text for HTML:

- `Convert(...).EscapeAttr()` — escape a value for safe inclusion inside an HTML attribute value.
- `Convert(...).EscapeHTML()` — escape a value for safe inclusion inside HTML content.

Both functions perform simple string replacements and will escape the characters: `&`, `<`, `>`, `"`, and `'`.
Note that existing HTML entities will be escaped again (for example `&amp;` -> `&amp;amp;`). This library follows a simple replace-based escaping strategy — if you need entity-aware unescaping/escaping, consider using a full HTML parser.

### Examples

```go
Convert(`Tom & Jerry's "House" <tag>`).EscapeAttr()
// -> `Tom &amp; Jerry&#39;s &quot;House&quot; &lt;tag&gt;`

Convert(`<div>1 & 2</div>`).EscapeHTML()
// -> `&lt;div&gt;1 &amp; 2&lt;/div&gt;`
```