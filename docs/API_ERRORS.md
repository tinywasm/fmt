# Errors Package Equivalents

Replace `errors` package functions for error handling with multilingual support:

| Go Standard | fmt Equivalent |
|-------------|----------------------|
| `errors.New()` | `Err(message)` |
| `fmt.Errorf()` | `Errf(format, args...)` |
| `fmt.Errorf("%w: %w", cause, sentinel)` | `ErrType(cause, sentinel)` |

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

Multilingual error support is **opt-in**. To enable translations, you must import `github.com/tinywasm/fmt/lang`.

```go
import "github.com/tinywasm/fmt/lang"

// Translation requires the lang package to be imported
// Note: Err no longer accepts language constants like ES as an argument.
// Use lang.OutLang(lang.ES) to change the global language.
err := Err("format", "invalid")
// → "formato inválido" (if lang.OutLang(lang.ES) was called)
```

For more details on multilingual support, see the [Translation Guide](TRANSLATE.md).

## Error Wrapping

`ErrType` allows wrapping an error with a sentinel for identification while preserving the original error's message and identity.

```go
var ErrNotFound = fmt.Err("not found")

func FindUser(id string) error {
    err := db.Query(...) // returns database error
    return fmt.ErrType(err, ErrNotFound)
}

// Consuming the error
err := FindUser("123")
if errors.Is(err, ErrNotFound) {
    // Identity preserved
}
fmt.Print(err) // "db connection timeout: not found"
```
