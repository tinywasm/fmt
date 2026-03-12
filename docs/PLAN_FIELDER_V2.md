# PLAN: Fielder v2 — Breaking Change (tinywasm/fmt)

← [README](../README.md)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.

## Context

The `Fielder` interface currently has 3 methods: `Schema()`, `Values()`, `Pointers()`.
`Values()` forces interface boxing of every field value — strings (16 bytes) don't fit in
the 8-byte interface data word and escape to the heap. This causes N+1 allocations per call,
making downstream libraries (`json`, `orm`, `form`) slower than stdlib `encoding/json` despite
avoiding `reflect`.

This plan modifies `fmt` to remove `Values()` from the interface and add helpers so consumers
can read values efficiently through `Pointers()`.

**Downstream dependents (must update AFTER this plan):**
- `tinywasm/json` → [PLAN_FIELDER_V2.md](../../json/docs/PLAN_FIELDER_V2.md)
- `tinywasm/orm` → [PLAN_FIELDER_V2.md](../../orm/docs/PLAN_FIELDER_V2.md)
- `tinywasm/form` → [PLAN_FIELDER_V2.md](../../form/docs/PLAN_FIELDER_V2.md)

---

## Stage 1: Remove `Values()` from Fielder interface

**File:** `field.go`

### 1.1 Update interface

```go
// BEFORE
type Fielder interface {
    Schema() []Field
    Values() []any
    Pointers() []any
}

// AFTER
type Fielder interface {
    Schema() []Field
    Pointers() []any
}
```

### 1.2 Update doc comment

Remove references to `Values()` from the contract comment (lines 44-48).
Keep the contract:
- `Schema()` and `Pointers()` MUST return slices of the same length.
- The i-th element in each slice corresponds to the same struct field.
- `Pointers()` returns pointers to fields for reading (dereference) and writing (scan/sync).

---

## Stage 2: Add `ReadValues` helper

**File:** `field.go` (append after Fielder interface)

```go
// ReadValues reads field values through Pointers by dereferencing based on FieldType.
// Used by consumers that need []any (e.g., orm for SQL args).
// Hot-path consumers (json, form) should read through Pointers directly to avoid boxing.
func ReadValues(schema []Field, ptrs []any) []any {
    vals := make([]any, len(schema))
    for i, f := range schema {
        switch f.Type {
        case FieldText:
            if p, ok := ptrs[i].(*string); ok {
                vals[i] = *p
            }
        case FieldInt:
            switch p := ptrs[i].(type) {
            case *int64:
                vals[i] = *p
            case *int:
                vals[i] = *p
            case *int32:
                vals[i] = *p
            case *uint:
                vals[i] = *p
            case *uint32:
                vals[i] = *p
            case *uint64:
                vals[i] = *p
            }
        case FieldFloat:
            switch p := ptrs[i].(type) {
            case *float64:
                vals[i] = *p
            case *float32:
                vals[i] = *p
            }
        case FieldBool:
            if p, ok := ptrs[i].(*bool); ok {
                vals[i] = *p
            }
        case FieldBlob:
            if p, ok := ptrs[i].(*[]byte); ok {
                vals[i] = *p
            }
        case FieldStruct:
            vals[i] = ptrs[i] // pointer to nested struct IS the Fielder
        }
    }
    return vals
}
```

---

## Stage 3: Add `ReadStringPtr` helper

**File:** `field.go` (append)

This helper allows json/form to read a string value from a pointer without boxing into `any`:

```go
// ReadStringPtr reads a string from a typed pointer.
// Returns the string value and true if the pointer is *string, or ("", false) otherwise.
func ReadStringPtr(ptr any) (string, bool) {
    if p, ok := ptr.(*string); ok {
        return *p, true
    }
    return "", false
}
```

---

## Stage 4: Add `WriteInt` and `WriteFloat` to Conv

These allow json to write numbers directly to a Builder without going through `Convert(val).String()` which boxes the value into `any`.

**File:** `num_int.go` (append)

```go
// WriteInt writes an int64 as decimal text to the Conv output buffer.
func (c *Conv) WriteInt(v int64) {
    if v == 0 {
        c.WriteByte('0')
        return
    }
    if v < 0 {
        c.WriteByte('-')
        v = -v
    }
    // Write digits in reverse, then reverse
    start := len(c.out)
    for v > 0 {
        c.out = append(c.out, byte('0'+v%10))
        v /= 10
    }
    // Reverse the digits
    digits := c.out[start:]
    for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
        digits[i], digits[j] = digits[j], digits[i]
    }
}
```

**File:** `num_float.go` (append)

```go
// WriteFloat writes a float64 as decimal text to the Conv output buffer.
// Uses the existing wrFloat64 logic.
func (c *Conv) WriteFloat(v float64) {
    c.wrFloat64(v)
}
```

---

## Stage 5: Update docs

### 5.1 Update `docs/API_FIELD.md`

- Remove `Values()` from the Fielder interface documentation.
- Document `ReadValues()` as the helper for consumers that need `[]any`.
- Document `ReadStringPtr()`.

### 5.2 Update `docs/API_STRCONV.md`

- Document `WriteInt()` and `WriteFloat()` methods on Conv.

---

## Stage 6: Update tests

**File:** Any test files that implement Fielder mocks — remove `Values()` method from them.

Search all `*_test.go` files for `Values()` implementations and remove them.

```bash
grep -rn "func.*Values().*\[\]any" *.go *_test.go
```

---

## Stage 7: Run tests and publish

```bash
gotest
```

If all pass:

```bash
gopush 'fmt: remove Values() from Fielder, add ReadValues/WriteInt/WriteFloat helpers (breaking change v2)'
```

**Important:** Note the new version tag — downstream libraries need this exact version in their `go.mod`.

---

## Summary

| Stage | File(s) | Action |
|-------|---------|--------|
| 1 | `field.go` | Remove `Values()` from Fielder interface |
| 2 | `field.go` | Add `ReadValues()` helper |
| 3 | `field.go` | Add `ReadStringPtr()` helper |
| 4 | `num_int.go`, `num_float.go` | Add `WriteInt()`, `WriteFloat()` |
| 5 | `docs/API_FIELD.md`, `docs/API_STRCONV.md` | Update documentation |
| 6 | `*_test.go` | Remove Values() from test mocks |
| 7 | — | `gotest` + `gopush` |
