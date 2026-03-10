# PLAN: Add `JSON`, `FieldStruct` to `fmt.Field`

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing` and `reflect` for test helpers only.
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

The `tinywasm/fmt` package already has `Field`, `FieldType`, `Fielder` interface, and `Field.Input` (all published). This plan adds:

1. **`Field.JSON string`** — JSON key + modifiers (e.g., `"email,omitempty"`). Populated by `ormc` from `json:"..."` tags. Used by `tinywasm/json` to serialize/deserialize without reflect.
2. **`FieldStruct` constant** — New `FieldType` for nested struct fields. Allows `tinywasm/json` to recurse into nested `Fielder` types.

Both additions are **backward-compatible**: zero values preserve existing behavior.

## What This Plan Does NOT Cover

- Changes to `orm`, `form`, or `json` — those have their own independent plans.

---

## Stage 1: Add `JSON` Field and `FieldStruct`

← None | Next → [Stage 2](#stage-2-documentation-and-publish)

### 1.1 Update `field.go`

Add `JSON string` to the `Field` struct and `FieldStruct` to `FieldType`:

```go
const (
    FieldText   FieldType = iota // Any string
    FieldInt                     // Any integer
    FieldFloat                   // Any float
    FieldBool                    // Boolean
    FieldBlob                    // Binary data ([]byte)
    FieldStruct                  // Nested struct (implements Fielder)
)

type Field struct {
    Name    string
    Type    FieldType
    PK      bool
    Unique  bool
    NotNull bool
    AutoInc bool
    Input   string // UI hint for form layer
    JSON    string // JSON key + modifiers ("email,omitempty"). Empty = use Field.Name as key.
}
```

**`Field.JSON` semantics:**
- `""` (empty) → json codec uses `Field.Name` as JSON key (default).
- `"email"` → json codec uses `"email"` as key.
- `"email,omitempty"` → json codec uses `"email"` as key and omits zero values.
- `"-"` → json codec skips this field entirely.
- Format is identical to Go's `json:"..."` tag. `ormc` copies the tag value verbatim.

**`FieldStruct` semantics:**
- `Values()[i]` for a `FieldStruct` field returns a value that implements `fmt.Fielder`.
- `Pointers()[i]` returns a pointer to a value that implements `fmt.Fielder`.
- The json codec recurses into the nested `Fielder` for encoding/decoding.
- The form layer ignores `FieldStruct` fields (forms are flat).

### 1.2 Update `FieldType.String()`

```go
var fieldTypeNames = []string{"text", "int", "float", "bool", "blob", "struct"}
```

### 1.3 Tests

Update `field_test.go`:

- `TestFieldTypeStringStruct`: Verify `FieldStruct.String()` returns `"struct"`.
- `TestFieldJSONEmpty`: Verify `Field{}` has `JSON == ""`.
- `TestFieldJSONKey`: Verify `Field{JSON: "email"}` stores the value.
- `TestFieldJSONOmitEmpty`: Verify `Field{JSON: "email,omitempty"}` stores modifiers.
- `TestFieldJSONExclude`: Verify `Field{JSON: "-"}` stores exclusion marker.
- `TestFieldStructType`: Verify `Field{Type: FieldStruct}` has correct type.

```bash
gotest
```

---

## Stage 2: Documentation and Publish

← [Stage 1](#stage-1-add-json-field-and-fieldstruct) | None →

### 2.1 Update `docs/API_FIELD.md`

Add documentation for:
- `Field.JSON`: purpose, value format, populated by `ormc`, consumed by `tinywasm/json`.
- `FieldStruct`: purpose, `Fielder` recursion contract, form ignores it.

### 2.2 Update `README.md`

Ensure docs index links to `API_FIELD.md`.

### 2.3 Run full test suite

```bash
gotest
```

### 2.4 Publish

```bash
gopush 'add JSON field and FieldStruct type to Field for json codec integration'
```

---

## Summary

| File | Action |
|------|--------|
| `field.go` | Add `JSON string` to `Field`, add `FieldStruct` to `FieldType` constants |
| `field_test.go` | Add 6 tests |
| `docs/API_FIELD.md` | Document JSON and FieldStruct |

Backward-compatible. Zero values preserve existing behavior.
