# PLAN: Widget Interface — tinywasm/fmt

**Module:** `github.com/tinywasm/fmt`
**Breaking change:** Yes — adds a new field to `Field` struct and a new interface.
**Execution order:** Must be completed and published BEFORE `tinywasm/form` and `tinywasm/orm` plans are executed.
**Downstream dependencies:** `tinywasm/form` (PLAN_INPUT_TAG.md) and `tinywasm/orm` (PLAN_INPUT_TAG.md) both depend on the `fmt.Widget` interface defined here. Do not begin those plans until this one is published and its new version is available via `go get`.

---

## Context

`tinywasm/fmt` is the base package for the tinywasm ecosystem. It defines:
- `Field` struct — schema metadata for a single struct field (type, constraints, validation rules)
- `Fielder` / `Model` / `Validator` interfaces — contracts used by orm, form, and json layers
- `ValidateFields()` — central validation entry point called by ormc-generated `Validate(action byte)` methods

Currently `Field` has no concept of *input type* — the UI layer (`tinywasm/form`) infers the input type by matching field names via a registry. This is implicit, magic, and hard to maintain.

The goal of this plan is to add a `Widget` interface to `fmt` and a corresponding `Widget Widget` field to `Field`, so that the input type is declared explicitly in the schema and validated through a unified code path.

---

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (`go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **TinyGo Compatible:** No `fmt`, `strings`, `strconv`, `errors` from stdlib. Use `tinywasm/fmt` itself (be careful of self-import — use the package's own functions directly, not importing itself).
- **No `reflect` at runtime.**

---

## What Changes

### 1. New interface: `Widget` in `field.go`

Add this interface to `field.go`:

```go
// Widget is the contract for a semantic input type.
// It is implemented by tinywasm/form/input types and custom project inputs.
// Stored in Field.Widget; set by ormc code generation from the `input:` struct tag.
//
// Implementations must be stateless templates — Clone() returns a fresh copy
// for thread-safe use per request in the form layer.
//
// RenderHTML is intentionally excluded: rendering is a concern of tinywasm/form,
// not of the base schema package.
type Widget interface {
    Type() string                  // Semantic input type (e.g., "email", "textarea"). Industry standard: HTML `type` attribute, JSON Schema `type` field.
    Validate(value string) error   // Semantic validation for this input type
    Clone() Widget             // Returns a fresh template instance (no parentID/name)
}
```

**Why not include `RenderHTML()`:** Interface Segregation — `fmt.Widget` has two distinct consumers:
- `orm` and validation: need only `Name()` + `Validate()`
- `form` and UI (SSR/WASM): need `Name()` + `Validate()` + `RenderHTML()`

If `RenderHTML()` is included in `fmt.Widget`, every custom input implementation — even backend-only or validation-only ones — must implement HTML rendering even when irrelevant. An interface should declare only what all its implementors need across all their use contexts.

`dom.Component` already declares `RenderHTML()` and `input.Input` already embeds it. The `form` package type-asserts `field.Widget.(dom.Component)` for rendering — no need for `fmt` to re-declare it. This is not a backend-vs-WASM concern: the entire ecosystem is isomorphic. It is purely about keeping interface contracts minimal and role-specific.

### 2. Add `Widget Widget` to `Field` struct in `field.go`

```go
type Field struct {
    Name      string
    Type      FieldType
    PK        bool
    Unique    bool
    NotNull   bool
    AutoInc   bool
    OmitEmpty bool
    Widget Widget  // ← NEW: set by ormc from `input:` tag. nil = no UI binding.
    Permitted             // embedded: Permitted validation rules (min/max/chars)
}
```

### 3. Update `Field.Validate()` in `field.go`

Update the method to call `Widget.Validate()` before `Permitted.Validate()`:

```go
func (f Field) Validate(value string) error {
    if f.NotNull && value == "" {
        return Err(f.Name, "required")
    }
    if value == "" {
        return nil
    }
    if f.Widget != nil {
        if err := f.Widget.Validate(value); err != nil {
            return err
        }
    }
    if f.hasPermittedRules() {
        return f.Permitted.Validate(f.Name, value)
    }
    return nil
}
```

**Composition logic:**
- `Widget.Validate()` checks semantic validity (e.g., email format, valid IP, etc.)
- `Permitted.Validate()` checks character rules and min/max length
- Both run independently; `Permitted` rules are additive (e.g., `input:"email,min=10"`)

### 4. No changes needed to `ValidateFields()`, `Fielder`, `Model`, `Validator`, or any other interface

`ValidateFields()` already calls `field.Validate(val)` per field — it will automatically benefit from the new `Widget.Validate()` call without modification.

---

## Files to Modify

| File | Change |
|---|---|
| `field.go` | Add `Widget` interface, add `Widget Widget` to `Field`, update `Field.Validate()`, update doc comment on `Validator` interface (line ~82) to remove reference to `form.ValidateStructFormats(m)` — that call is eliminated by the orm plan |

No other files in `tinywasm/fmt` require changes.

---

## Tests

Add tests in `field_test.go` (create if it doesn't exist):

1. `TestFieldValidate_WithWidget_Valid` — field with `Widget` set, valid value → no error
2. `TestFieldValidate_WithWidget_Invalid` — field with `Widget` set, invalid value → error from Widget
3. `TestFieldValidate_WithWidgetAndPermitted` — Widget passes but Permitted.Minimum fails → error from Permitted
4. `TestFieldValidate_NilWidget` — field with `Widget = nil`, existing Permitted rules still apply
5. `TestFieldValidate_NotNull_EmptyValue` — `Widget.Validate()` must NOT be called when value is empty and NotNull check already returns error

Use a local stub implementing `Widget` in the test file — do not import `tinywasm/form/input` in tests:

```go
type stubInput struct{ kind string }
func (s stubInput) Type() string              { return s.kind }
func (s stubInput) Clone() Widget         { return s }
func (s stubInput) Validate(v string) error   {
    if v == "invalid" { return Err(s.kind, "invalid value") }
    return nil
}
```



