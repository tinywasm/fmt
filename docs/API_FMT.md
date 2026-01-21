# Fmt Package Equivalents

Replace `fmt` package functions for formatting:

| Go Standard | fmt Equivalent |
|-------------|----------------------|
| `fmt.Sprintf()` | `Sprintf(format, args...)` |
| `fmt.Sprint()` | `Convert(v).String()` |
| `fmt.Fprintf()` | `Fprintf(w, format, args...)` |
| `fmt.Sscanf()` | `Sscanf(src, format, args...)` |

## String Formatting

```go
// Printf-style formatting
result := Sprintf("Hello %s, you have %d messages", "John", 5)
// out: "Hello John, you have 5 messages"

// Multiple format specifiers
result := Sprintf("Number: %d, Float: %.2f, Bool: %v", 42, 3.14159, true)
// out: "Number: 42, Float: 3.14, Bool: true"

// Advanced formatting (hex, binary, octal)
result := Sprintf("Hex: %x, Binary: %b, Octal: %o", 255, 10, 8)
// out: "Hex: ff, Binary: 1010, Octal: 10"

// Write formatted output to io.Writer
var buf bytes.Buffer
Fprintf(&buf, "Hello %s, count: %d\n", "world", 42)

// Write to file
file, _ := os.Create("output.txt")
Fprintf(file, "Data: %v\n", someData)

// Parse formatted text from string (like fmt.Sscanf)
var pos int
var name string
n, err := Sscanf("!3F question", "!%x %s", &pos, &name)
// n = 2, pos = 63, name = "question", err = nil

// Parse complex formats
var code, unicode int
var word string
n, err := Sscanf("!3F U+003F question", "!%x U+%x %s", &code, &unicode, &word)
// n = 3, code = 63, unicode = 63, word = "question", err = nil

// Localized string formatting
// Uses the current global language or default (EN)
Sprintf("Error: %L", D.Invalid)
// out (EN): "Error: invalid"
// out (ES): "Error: inválido"
```

For more details on translation and `LocStr` usage, see [TRANSLATE.md](TRANSLATE.md).