# PLAN: Field v3 — Breaking Change (tinywasm/fmt)

← [README](../README.md)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **Documentation First:** Update docs before coding.

## Context

`fmt.Field` currently carries two layer-specific string fields:

| Field | Consumer | Problem |
|-------|----------|---------|
| `JSON string` | tinywasm/json | Redundant — just copies `json:"..."` tag verbatim. If the dev wants a custom key, they should rename the struct field. |
| `Input string` | tinywasm/form | Redundant — form already resolves input type by field name heuristic. Explicit overrides belong in struct tags parsed at generation time, not in runtime metadata. |

Additionally, there is no shared validation mechanism. Each layer validates independently
(form has `ValidateField` via `input.Permitted`, json/orm have nothing).

### Design decision: Field embeds Permitted, Validate is a method of Field

Validation rules belong in the schema — `Field` already has constraints (`NotNull`, `Unique`)
which ARE validation rules. Taking this to its logical conclusion:

- **Field embeds `Permitted`** — character rules, min/max, format checks live in the schema.
- **`Field.Validate(value string) error`** — each field validates its own value.
- **`ValidateFielder(f Fielder) error`** — generic function iterates schema + pointers, calls `Field.Validate`.
- **No standalone validator functions for character limits** (`ValidateMinLen`, etc.) — all covered by `Permitted`.
- **Complex formats are delegated** — `ormc` generates a `Validate()` method that first calls `fmt.ValidateFielder(m)` for character/length rules, then appends custom standalone format validators (e.g., `form.ValidateEmail(m.Email)`) as needed.

**Downstream dependents (must update AFTER this plan):**
- `tinywasm/json` → [PLAN.md](../../json/docs/PLAN.md)
- `tinywasm/orm` → [PLAN.md](../../orm/docs/PLAN.md)
- `tinywasm/form` → [PLAN.md](../../form/docs/PLAN.md)

---

## Stage 1: Add `Permitted` and `Format` to fmt

**File:** `permitted.go` (new file, replaces `form/input/permitted.go`)

`Permitted` is the core validation engine currently living in `form/input`. Moving it to fmt
makes it available to all 4 layers (form, json, orm, direct validation) without importing form+DOM.

### 1.1 Permitted struct — no maps, ASCII ranges

Replace 3 maps (`valid_letters`, `valid_tilde`, `valid_number`) from `form/input/permitted.go`
with ASCII range checks following `mapping.go` patterns.

Reuse from `mapping.go`:
- `aL []rune` (accented lowercase) — replaces `valid_tilde` map
- `aU []rune` (accented uppercase) — replaces `valid_tilde` map

```go
// Permitted validates strings character-by-character against a configurable whitelist.
//
// Zero value = nothing permitted (strictest). Enable flags to allow character classes.
// Moved from form/input to fmt for cross-layer reuse.
type Permitted struct {
    Letters   bool     // a-z, A-Z, ñ, Ñ
    Tilde     bool     // á, é, í, ó, ú (and uppercase) — uses aL/aU from mapping.go
    Numbers   bool     // 0-9
    Spaces    bool     // ' '
    BreakLine bool     // '\n'
    Tab       bool     // '\t'
    Extra     []rune   // additional allowed characters (e.g., '@', '.', '-')
    NotAllowed []string // forbidden substrings
    Minimum   int      // min length (0 = no limit)
    Maximum   int      // max length (0 = no limit)
    StartWith *Permitted // rules for first character (nil = same as main rules)
}
```

### 1.3 Permitted.Validate — unified validation method

```go
// Validate checks that text conforms to the permitted rules.
// Order: length → forbidden substrings → start-with → characters.
func (p Permitted) Validate(field, text string) error {
    // Length checks (using range to count runes without importing unicode/utf8)
    var count int
    if p.Minimum != 0 || p.Maximum != 0 {
        for range text {
            count++
        }
    }
    if p.Minimum != 0 && count < p.Minimum {
        return Err(field, "minimum", p.Minimum, "chars")
    }
    if p.Maximum != 0 && count > p.Maximum {
        return Err(field, "maximum", p.Maximum, "chars")
    }

    // Forbidden substrings
    for _, na := range p.NotAllowed {
        if Contains(text, na) {
            return Err(field, "text not allowed", na)
        }
    }

    // StartWith check (first rune only)
    if p.StartWith != nil && len(text) > 0 {
        var firstRune rune
        for _, r := range text {
            firstRune = r
            break
        }
        if err := p.StartWith.Validate(field, string(firstRune)); err != nil {
            return Err(field, "start", err)
        }
    }

    // Character-by-character validation — NO MAPS, only range checks
    for _, r := range text {
        if p.isAllowed(r) {
            continue
        }
        return errCharNotAllowed(field, r)
    }

    return nil
}
```

### 1.4 isAllowed — ASCII ranges, reuses mapping.go slices

```go
// isAllowed checks if a rune is permitted using ASCII ranges and slice lookups.
// Follows the same pattern as mapping.go (toUpperRune, isWordSeparatorChar).
func (p Permitted) isAllowed(r rune) bool {
    // ASCII letters: a-z, A-Z (fast path)
    if p.Letters {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
            return true
        }
        if r == 'ñ' || r == 'Ñ' {
            return true
        }
    }

    // Numbers: 0-9 (ASCII range)
    if p.Numbers && r >= '0' && r <= '9' {
        return true
    }

    // Tildes: reuse aL/aU slices from mapping.go
    if p.Tilde {
        for _, a := range aL {
            if r == a {
                return true
            }
        }
        for _, a := range aU {
            if r == a {
                return true
            }
        }
    }

    // Whitespace (individual checks)
    if p.Spaces && r == ' ' {
        return true
    }
    if p.BreakLine && r == '\n' {
        return true
    }
    if p.Tab && r == '\t' {
        return true
    }

    // Extra allowed characters (linear scan, typically 0-5 items)
    for _, c := range p.Extra {
        if r == c {
            return true
        }
    }

    return false
}
```

### 1.5 errCharNotAllowed helper

```go
func errCharNotAllowed(field string, r rune) error {
    switch {
    case r == ' ':
        return Err(field, "space not allowed")
    case r == '\t':
        return Err(field, "tab not allowed")
    case r == '\n':
        return Err(field, "newline not allowed")
    default:
        return Err(field, "character not allowed", string(r))
    }
}
```

### 1.6 Why no maps — comparison table

| Approach | Binary cost | TinyGo friendly | Lookup cost |
|----------|------------|-----------------|-------------|
| `map[rune]bool` (current) | ~500 bytes per map + runtime hash | No — heap alloc, GC | O(1) amortized but hash overhead |
| ASCII range checks | ~20 bytes of instructions | Yes — pure stack | O(1) true |
| `aL`/`aU` slice scan (existing) | 0 extra — already in mapping.go | Yes — no alloc | O(n) but n≤23 |

---

## Stage 2: Update `Field` struct — embed Permitted, add Validate method

**File:** `field.go`

### 2.1 Remove `JSON` and `Input`, embed `Permitted`, add `OmitEmpty`

```go
// BEFORE
type Field struct {
    Name    string
    Type    FieldType
    PK      bool
    Unique  bool
    NotNull bool
    AutoInc bool
    Input   string
    JSON    string
}

// AFTER
type Field struct {
    Name      string
    Type      FieldType
    PK        bool
    Unique    bool
    NotNull   bool
    AutoInc   bool
    OmitEmpty bool      // omit from JSON when zero value
    Permitted            // embedded: validation rules (characters, min/max)
}
```

### 2.2 Add `Field.Validate` method

```go
// Validate checks a string value against this field's constraints.
// Checks NotNull first, then delegates to embedded Permitted.
func (f Field) Validate(value string) error {
    if f.NotNull && value == "" {
        return Err(f.Name, "required")
    }
    if value == "" {
        return nil // empty + not required = ok
    }
    // Only run Permitted validation if any rule is configured
    if f.hasPermittedRules() {
        return f.Permitted.Validate(f.Name, value)
    }
    return nil
}

// hasPermittedRules returns true if any Permitted field is non-zero.
func (f Field) hasPermittedRules() bool {
    return f.Letters || f.Tilde || f.Numbers || f.Spaces ||
        f.BreakLine || f.Tab || len(f.Extra) > 0 ||
        len(f.NotAllowed) > 0 || f.Minimum > 0 || f.Maximum > 0 ||
        f.StartWith != nil
}
```

### 2.3 Update doc comments

```go
// Field describes a single field in a struct's schema.
// It provides type metadata, constraint flags, and validation rules
// used by database (orm), transport (json), UI (form), and validation layers.
//
// Validation rules are embedded via Permitted. When a field has validation
// configured, Field.Validate(value) checks the value against all rules.
// Fields without validation rules pass any value.
```

---

## Stage 3: Add `Validator`, `SafeFielder`, and `ValidateFielder`

**File:** `field.go` (append after Fielder interface)

### 3.1 Interfaces

```go
// Validator is implemented by types that can self-validate.
// Generated by ormc: func (m *X) Validate() error { if err := fmt.ValidateFielder(m); err != nil { return err }; return form.ValidateStructFormats(m) }
// Used by form, json.Decode, and orm pre-insert to enforce data integrity.
type Validator interface {
    Validate() error
}

// SafeFielder combines schema access with validation.
// Handlers that receive user input should accept SafeFielder
// to enforce compile-time validation guarantees.
type SafeFielder interface {
    Fielder
    Validator
}
```

### 3.2 Generic ValidateFielder function

```go
// ValidateFielder validates all fields of a Fielder by iterating Schema + Pointers.
// For each FieldText field, reads the string value and calls Field.Validate.
// For non-text fields with NotNull, checks against zero value.
//
// This is the single validation entry point — ormc-generated Validate()
// methods are one-liners that call this function.
func ValidateFielder(f Fielder) error {
    schema := f.Schema()
    ptrs := f.Pointers()
    for i, field := range schema {
        switch field.Type {
        case FieldText:
            val, _ := ReadStringPtr(ptrs[i])
            if err := field.Validate(val); err != nil {
                return err
            }
        case FieldStruct:
            // Recursive validation for nested structs
            if validator, ok := ptrs[i].(Validator); ok {
                if err := validator.Validate(); err != nil {
                    return err
                }
            } else if fielder, ok := ptrs[i].(Fielder); ok {
                if err := ValidateFielder(fielder); err != nil {
                    return err
                }
            }
        default:
            // Non-text fields: only check NotNull (zero value check)
            if field.NotNull && isZeroPtr(ptrs[i], field.Type) {
                return Err(field.Name, "required")
            }
        }
    }
    return nil
}

// isZeroPtr checks if a pointer points to a zero value.
func isZeroPtr(ptr any, ft FieldType) bool {
    switch ft {
    case FieldInt:
        switch p := ptr.(type) {
        case *int64:
            return *p == 0
        case *int:
            return *p == 0
        case *int32:
            return *p == 0
        case *uint:
            return *p == 0
        case *uint32:
            return *p == 0
        case *uint64:
            return *p == 0
        }
    case FieldFloat:
        switch p := ptr.(type) {
        case *float64:
            return *p == 0
        case *float32:
            return *p == 0
        }
    case FieldBool:
        if p, ok := ptr.(*bool); ok {
            return !*p
        }
    case FieldBlob:
        if p, ok := ptr.(*[]byte); ok {
            return len(*p) == 0
        }
    }
    return false
}
```

> **Note:** `isZeroPtr` is also used by `tinywasm/json` (encode.go:151). It currently lives
> in json. With this change, move it to fmt so both json and field validation share it.

---

## Stage 4: Update `ReadValues` and helpers

**File:** `field.go`

Remove any reference to `field.JSON` or `field.Input` in comments or helper functions.
No functional change needed — `ReadValues`, `ReadStringPtr` don't use those fields.

---

## Stage 5: Update tests

**File:** `field_test.go` and any `*_test.go`

- Remove `Input` and `JSON` from all test `Field` literals.
- Add `OmitEmpty: true` where appropriate.
- Add tests for `Permitted.Validate()` in `permitted_test.go`:
  - Character-class tests (Letters, Numbers, Tilde, Spaces, Extra).
  - Min/Max length tests.
  - NotAllowed substring tests.
  - StartWith recursive validation.
- Add tests for `Field.Validate()` — NotNull + Permitted integration.
- Add tests for `ValidateFielder()` — full Fielder validation.

---

## Stage 6: Update docs

### 6.1 Update `docs/API_FIELD.md`

- Remove `Input` and `JSON` from Field struct documentation.
- Remove "Input Hint Semantics" and "JSON Field Semantics" sections.
- Add `OmitEmpty` documentation.
- Add `Permitted` embed documentation with all fields.
- Add `Field.Validate(value string) error` method documentation.
- Add `Validator`, `SafeFielder` interface documentation.
- Add `ValidateFielder(f Fielder) error` function documentation.

---

## Stage 7: Run tests and publish

```bash
gotest
```

If all pass:

```bash
gopush 'fmt: Field v3 — remove JSON/Input, embed Permitted, Field.Validate, ValidateFielder (breaking change)'
```

**Important:** Note the new version — all downstream libraries need this version in `go.mod`.

---

## Summary

| Stage | File(s) | Action |
|-------|---------|--------|
| 1 | `permitted.go` | New file: `Permitted` struct, no Format enum, no maps, ASCII ranges, reuses `aL`/`aU` from mapping.go |
| 2 | `field.go` | Remove `JSON`/`Input`, embed `Permitted`, add `OmitEmpty`, add `Field.Validate()` |
| 3 | `field.go` | Add `Validator`, `SafeFielder` interfaces, `ValidateFielder()` generic function, `isZeroPtr` |
| 4 | `field.go` | Clean up comments/helpers |
| 5 | `*_test.go` | Update test Field literals, add permitted/validate/field tests |
| 6 | `docs/API_FIELD.md` | Update documentation |
| 7 | — | `gotest` + `gopush` |

## Impact on downstream

| Package | What breaks | What to do |
|---------|------------|------------|
| tinywasm/orm | `ormc` generates `Input:` and `JSON:` in Field literals | Generate `Permitted:` config in schema, generate custom `Validate()` for structural checks |
| tinywasm/json | `parseJSONTag()` reads `field.JSON`, has own `isZeroPtr` | Use `field.Name` as key, `field.OmitEmpty` for omitempty, use `fmt.isZeroPtr` |
| tinywasm/form | `input.Permitted` + `field.Input` | Delete `input/permitted.go`, use `fmt.Permitted`, remove `field.Input` usage |
