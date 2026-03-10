# PLAN: Add `JSONEscape` and `IsZero` to `fmt`

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing` for tests.
- **Testing Runner:** Install and use `gotest`:
  ```bash
  go install github.com/tinywasm/devflow/cmd/gotest@latest
  ```
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.
- **TinyGo Compatible:** No `fmt`, `strings`, `strconv`, `errors` from stdlib.
- **No maps** in exported types (binary bloat in WASM).

## Context

The `tinywasm/fmt` package already has `Field`, `FieldType` (including `FieldStruct`), `Fielder` interface, `Field.Input`, `Field.JSON`, `Builder`, and `Convert` (all published). This plan adds the remaining utilities needed by `tinywasm/json`:

1. **`JSONEscape(s string, b *Builder)`** — JSON string escaping utility. Shared by `tinywasm/json` and any other package that needs JSON-safe strings. Avoids duplicating escaping logic.
2. **`IsZero(v any) bool`** — Zero-value detection for `omitempty` support. Reusable across the ecosystem.

### Why in `fmt`?

- `fmt` already has `Builder`, `Convert`, number/bool parsing. Adding JSON escaping and zero-check here means `tinywasm/json` stays minimal — only the parser and Fielder encode/decode logic.
- Every utility added to `fmt` reduces the total binary size because it's shared, not duplicated.
- `JSONEscape` is a string operation — same domain as `fmt`'s existing string utilities.

## What This Plan Does NOT Cover

- Changes to `orm`, `form`, or `json` — those have their own independent plans.

## Already Completed

The following were implemented in previous iterations and are **not part of this plan**:

- `FieldStruct` constant in `FieldType`
- `Field.JSON string` field
- `Field.Input string` field
- `Fielder` interface (`Schema()`, `Values()`, `Pointers()`)
- `fieldTypeNames` updated with "struct"

---

## Stage 1: Add `JSONEscape` and `IsZero`

← None | Next → [Stage 2](#stage-2-documentation-and-publish)

### 1.1 Add `JSONEscape` to existing string utilities

Add to an appropriate existing file (e.g., `quote.go` which already handles quoting/escaping):

```go
// JSONEscape writes s to b with JSON string escaping (without surrounding quotes).
// Escapes: " → \", \ → \\, newline → \n, carriage return → \r, tab → \t,
// control chars (< 0x20) → \u00XX.
//
// The caller is responsible for writing the surrounding double quotes.
// This design allows the caller to compose JSON strings without extra allocations.
func JSONEscape(s string, b *Builder) {
    for i := 0; i < len(s); i++ {
        c := s[i]
        switch c {
        case '"':
            b.WriteString(`\"`)
        case '\\':
            b.WriteString(`\\`)
        case '\n':
            b.WriteString(`\n`)
        case '\r':
            b.WriteString(`\r`)
        case '\t':
            b.WriteString(`\t`)
        default:
            if c < 0x20 {
                b.WriteString(`\u00`)
                b.WriteByte("0123456789abcdef"[c>>4])
                b.WriteByte("0123456789abcdef"[c&0xf])
            } else {
                b.WriteByte(c)
            }
        }
    }
}
```

### 1.2 Add `IsZero` to existing utilities

Add to an appropriate existing file (e.g., `convert.go` which already handles type conversions):

```go
// IsZero reports whether v is the zero value for its type.
// Supports: string, bool, int (all sizes), uint (all sizes),
// float32, float64, []byte, nil.
// Returns false for unrecognized types.
func IsZero(v any) bool {
    switch val := v.(type) {
    case nil:
        return true
    case string:
        return val == ""
    case bool:
        return !val
    case int:
        return val == 0
    case int8:
        return val == 0
    case int16:
        return val == 0
    case int32:
        return val == 0
    case int64:
        return val == 0
    case uint:
        return val == 0
    case uint8:
        return val == 0
    case uint16:
        return val == 0
    case uint32:
        return val == 0
    case uint64:
        return val == 0
    case float32:
        return val == 0
    case float64:
        return val == 0
    case []byte:
        return len(val) == 0
    }
    return false
}
```

### 1.3 Tests

New file `json_escape_test.go` (or add to `quote_test.go`):

- `TestJSONEscapeEmpty`: Empty string → nothing written.
- `TestJSONEscapePlain`: ASCII string → unchanged.
- `TestJSONEscapeQuotes`: `"hello"` → `\"hello\"`.
- `TestJSONEscapeBackslash`: `a\b` → `a\\b`.
- `TestJSONEscapeNewlines`: `\n`, `\r`, `\t` → escaped.
- `TestJSONEscapeControlChars`: `\x00`..`\x1f` → `\u00XX`.
- `TestJSONEscapeUnicode`: Multi-byte UTF-8 characters pass through unescaped (JSON allows literal UTF-8).

Add to existing test file or new `is_zero_test.go`:

- `TestIsZeroNil`: `nil` → true.
- `TestIsZeroString`: `""` → true, `"x"` → false.
- `TestIsZeroBool`: `false` → true, `true` → false.
- `TestIsZeroInt`: `0` → true, `1` → false. Also `int8`, `int16`, `int32`, `int64`.
- `TestIsZeroUint`: `uint(0)` → true.
- `TestIsZeroFloat`: `0.0` → true, `1.5` → false.
- `TestIsZeroBytes`: `[]byte{}` → true, `[]byte{1}` → false.
- `TestIsZeroUnknown`: Unrecognized type → false.

```bash
gotest
```

---

## Stage 2: Documentation and Publish

← [Stage 1](#stage-1-add-jsonescape-and-iszero) | None →

### 2.1 Update `docs/API_FIELD.md`

Add:
- `Field.JSON`: purpose, format, `ormc` populates it, `tinywasm/json` consumes it.
- `FieldStruct`: purpose, Fielder recursion contract, form ignores it.

### 2.2 Create `docs/API_JSON_ESCAPE.md` or add to `docs/API_STRINGS.md`

Document:
- `JSONEscape(s string, b *Builder)`: escaping rules, caller writes quotes.
- `IsZero(v any) bool`: supported types, returns false for unknown.

### 2.3 Update `README.md`

Ensure all new docs are linked in the index.

### 2.4 Run full test suite

```bash
gotest
```

### 2.5 Publish

```bash
gopush 'add JSONEscape and IsZero utilities for json codec support'
```

---

## Summary

| File | Action |
|------|--------|
| `quote.go` (or new `json_escape.go`) | Add `JSONEscape(s string, b *Builder)` |
| `convert.go` (or new `is_zero.go`) | Add `IsZero(v any) bool` |
| Tests | ~15 new test cases |
| `docs/API_JSON_ESCAPE.md` | Document JSONEscape and IsZero |

All additions are backward-compatible. Zero values preserve existing behavior.
