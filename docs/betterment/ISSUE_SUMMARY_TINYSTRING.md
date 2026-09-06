# fmt - Library Context for LLM Maintenance

## What is fmt?

Lightweight Go library for string manipulation with fluid API, specifically designed for small devices and web applications using TinyGo compiler.

## Core Problem

**Excessive binary sizes in WebAssembly when using Go standard library**, specifically:
- Large binaries slow web app loading
- Standard library (`fmt`, `strings`, `strconv`) adds significant overhead  
- Memory constraints on small devices and edge computing
- Universal need for string manipulation in all projects

## Primary Goal

Enable Go WebAssembly adoption by reducing binary size while providing essential string operations through manual implementations that avoid standard library imports.

## Key Features

- Fluid chainable API
- Zero standard library dependencies
- TinyGo compatible
- Universal type conversion (string, int, float, bool)
- Manual implementations replace: `strconv.ParseFloat`, `strconv.FormatFloat`, `strconv.ParseInt`, `strconv.FormatInt`, `strings.IndexByte`, `fmt.Sprintf`

## Basic Usage Pattern

```go
import "webtyp.com/fmt"

// Basic string processing
result := tinystring.Convert("MÍ téxtO").Tilde().String()
// Output: "MI textO"

// Type conversion and chaining
result := tinystring.Convert(42).ToUpper().String()
// Output: "42"

// Complex chaining
result := tinystring.Convert("Él Múrcielago Rápido")
    .Tilde()
    .CamelLow()
    .String()
// Output: "elMurcielagoRapido"

// Memory optimization with pointers
text := "Él Múrcielago Rápido"
tinystring.Convert(&text).Tilde().CamelLow().Apply()
// text is now: "elMurcielagoRapido"
```

## Core Operations

**Initialization & Output:**
- `Convert(v any)` - Initialize with any type
- `String()` - Get result as string 
- `Apply()` - Modify original string pointer

**Text Transformations:**
- `Tilde()` - Remove accents/diacritics
- `ToLower()`, `ToUpper()` - Case conversion
- `Capitalize()` - First letter of each word
- `CamelLow()`, `CamelUp()` - camelCase conversion
- `SnakeLow()`, `SnakeUp()` - snake_case conversion

**String Operations:**
- `Split(data, separator)` - Split strings
- `Join(sep...)` - Join string slices
- `Replace(old, new, n...)` - Replace substrings
- `TrimPrefix()`, `TrimSuffix()`, `TrimSpace()` - TrimSpace operations
- `Contains()`, `Count()` - Search operations
- `Repeat(n)` - Repeat strings

**Advanced Features:**
- `Truncate(maxWidth, reservedChars...)` - Smart truncation
- `TruncateName(maxCharsPerWord, maxWidth)` - Name truncation for UI
- `Round(decimals)` - Numeric rounding with `Down()` modifier
- `Thousands()` - Thousand separators
- `Fmt(format, args...)` - sprintf-style formatting
- `Quote()` - Add quotes with escaping

**Type Conversions:**
- `Bool()` - Convert to boolean
- `Int(base...)`, `Uint(base...)` - Integer conversion
- `Float64()` - Float conversion
- `StringErr()` - Get result with error handling

## Installation

```bash
go get webtyp.com/fmt
```

## Architecture Notes

- Manual implementations avoid standard library bloat
- Optimized for binary size over runtime performance
- Thread-safe operations
- Supports pointer optimization to reduce allocations
- WebAssembly-first design philosophy
