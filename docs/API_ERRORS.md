# Errors Package Equivalents

Replace `errors` package functions for error handling with multilingual support:

| Go Standard | fmt Equivalent |
|-------------|----------------------|
| `errors.New()` | `Err(message)` |
| `fmt.Errorf()` | `Errf(format, args...)` |

## Error Creation

```go
// Multiple error messages and types
err := Err("invalid format", "expected number", 404)
// out: "invalid format expected number 404"

// Formatted errors (like fmt.Errorf)
err := Errf("invalid value: %s at position %d", "abc", 5)
// out: "invalid value: abc at position 5"
```

## Multilingual Error Messages

For multilingual error messages using dictionary terms that can be translated into multiple languages, see the [Translation Guide](TRANSLATE.md).

```go
// Using dictionary terms for translatable errors
// Note: requires import _ "github.com/tinywasm/fmt/dictionary"
err := Err("format", "invalid")
// → "format invalid" (in English) or translated based on global language setting

// Force specific language
err := Err(ES, "format", "invalid")
// → "formato inválido"
```