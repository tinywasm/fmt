# JSON Utilities

Utilities for JSON encoding and zero-value detection. These are used by `tinywasm/json` and other packages to minimize binary size by sharing common logic.

## JSONEscape

`JSONEscape` writes a string to a `Builder` with JSON string escaping (without surrounding quotes).

```go
func JSONEscape(s string, b *fmt.Builder)
```

### Escaping Rules

- `"` → `\"`
- `\` → `\\`
- Newline (`\n`) → `\n`
- Carriage return (`\r`) → `\r`
- Tab (`\t`) → `\t`
- Control characters (< 0x20) → `\u00XX` (hexadecimal)
- All other characters (including UTF-8) are written unescaped.

### Usage

The caller is responsible for writing the surrounding double quotes. This allows composing JSON strings without extra allocations.

```go
b := fmt.Convert()
b.WriteByte('"')
fmt.JSONEscape("hello \"world\"", b)
b.WriteByte('"')
s := b.String() // "\"hello \\\"world\\\"\""
```

## IsZero

`IsZero` reports whether a value is the zero value for its type. This is used for `omitempty` support in the JSON codec.

```go
func IsZero(v any) bool
```

### Supported Types

| Type | Zero Value |
|------|------------|
| `nil` | `true` |
| `string` | `""` |
| `bool` | `false` |
| `int` (all sizes) | `0` |
| `uint` (all sizes) | `0` |
| `float32`, `float64` | `0` |
| `[]byte` | `len(val) == 0` |

Returns `false` for unrecognized types.
