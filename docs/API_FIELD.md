# Field and Fielder API

The `field` API provides a standard way to describe struct schemas, access their values, and perform validation without using runtime reflection. This is essential for maintaining a small binary size in WebAssembly environments.

## FieldType

`FieldType` represents the abstract storage type of a struct field.

| Constant | String | Description |
|----------|--------|-------------|
| `FieldText` | `"text"` | String or text data |
| `FieldInt` | `"int"` | Integer numbers |
| `FieldFloat` | `"float"` | Floating point numbers |
| `FieldBool` | `"bool"` | Boolean values |
| `FieldBlob` | `"blob"` | Binary data |
| `FieldStruct` | `"struct"` | Nested struct |
| `FieldIntSlice` | `"intslice"` | Slice of integers (`[]int`) |
| `FieldStructSlice` | `"structslice"` | Slice of nested structs (`[]Fielder`) |

### String()

Returns the string representation of the `FieldType`.

```go
fmt.FieldInt.String() // returns "int"
```

## Field

`Field` describes a single field in a struct's schema with its metadata, constraints, and validation rules.

```go
type Field struct {
    Name      string
    Type      FieldType
    NotNull   bool
    OmitEmpty bool      // omit from JSON when zero value
    Widget    Widget    // semantic input type; nil = no UI binding (set by ormc from `input:` tag)
    DB        *FieldDB  // nil for formonly/transport structs
    Permitted           // embedded: validation rules (characters, min/max)
}
```

## FieldDB

`FieldDB` contains database-specific metadata, extracted from `Field` to keep transport/UI structs lean. When `DB` is `nil`, the field has no database concerns (e.g., `formonly` structs).

```go
type FieldDB struct {
    PK      bool
    Unique  bool
    AutoInc bool
}
```

### Helper methods on Field

Convenience methods to avoid nil-check boilerplate:

```go
func (f Field) IsPK() bool      { return f.DB != nil && f.DB.PK }
func (f Field) IsUnique() bool   { return f.DB != nil && f.DB.Unique }
func (f Field) IsAutoInc() bool  { return f.DB != nil && f.DB.AutoInc }
```

## Widget Interface

`Widget` is the contract for a semantic input type. It is implemented by `tinywasm/form/input` types and by custom project inputs defined in `web/inputs/`. Set by `ormc` code generation from the `input:` struct tag.

```go
type Widget interface {
    Type() string                              // Semantic type name (e.g., "email", "textarea")
    Validate(value string) error               // Semantic validation for this input type
    Clone(parentID, name string) Widget        // Returns a positioned instance; pass ("","") for a bare template
}
```

`Field.Validate()` calls `Widget.Validate()` before `Permitted.Validate()`. If the widget fails, `Permitted` is not evaluated. If `Widget` is `nil`, only `Permitted` rules apply.

```go
// Example: field with Widget + additive Permitted rules
Field{
    Name:      "email",
    Widget:    input.NewEmail(),            // validates email format
    Permitted: fmt.Permitted{Minimum: 10}, // additionally enforces min length
}
```

**Why `Clone(parentID, name)` instead of `Clone()`:** The form layer always needs positioned instances for rendering. A parameterless `Clone()` would force a separate `Build()` method — two constructors for the same concern. With `Clone(parentID, name)`, consumers that only need a bare template pass `("", "")`, and form consumers get a positioned instance directly.

**Why not include `RenderHTML()`:** Interface Segregation — ORM and validation consumers need only `Type()` and `Validate()`. Rendering is handled by `tinywasm/form`, which type-asserts `field.Widget.Clone(parentID, name).(input.Input)` for HTML output.

### Validation (Permitted)

Validation rules are embedded in the `Field` via the `Permitted` struct. This includes character-level whitelisting, length constraints, and structural format checks.

| Field | Type | Description |
|-------|------|-------------|
| `Letters` | `bool` | Allows `a-z`, `A-Z`, `ñ`, `Ñ` |
| `Tilde` | `bool` | Allows accented characters (`á`, `é`, etc.) |
| `Numbers` | `bool` | Allows `0-9` |
| `Spaces` | `bool` | Allows `' '` |
| `BreakLine` | `bool` | Allows `\n` |
| `Tab` | `bool` | Allows `\t` |
| `Extra` | `[]rune` | Additional allowed characters |
| `NotAllowed` | `[]string`| Forbidden substrings |
| `Minimum` | `int` | Minimum length (runes) |
| `Maximum` | `int` | Maximum length (runes) |
| `StartWith` | `*Permitted`| Rules for the first character |

### Field.Validate()

Checks a string value against the field's constraints in order: `NotNull` → `Widget.Validate()` → `Permitted`.

```go
err := field.Validate("some value")
```

## Fielder Interface

The `Fielder` interface is the shared contract between various layers (like ORM and UI forms) to interact with structs.

```go
type Fielder interface {
    Schema() []Field
    Pointers() []any
}
```

### Contract

- `Schema()` and `Pointers()` MUST return slices of the same length.
- The i-th element in each slice corresponds to the same struct field.
- `Pointers()` returns pointers to fields for reading (dereference) and writing.

## Validator and SafeFields

```go
// Validator can self-validate
type Validator interface {
    Validate(action byte) error
}

// Model describes a resource with a schema and a name.
type Model interface {
    Fielder
    ModelName() string
}

// SafeFields combines Fielder and Validator
type SafeFields interface {
    Fielder
    Validator
}
```

### ValidateFields() Helper

Generic function that iterates through a `Fielder`'s schema and pointers to perform full validation based on the action. It handles `FieldText` (calling `Field.Validate`) and `FieldStruct` (recursive validation).

| Action | PK + AutoInc | PK without AutoInc | NotNull | Permitted |
|--------|--------------|--------------------|---------|-----------|
| `'c'` create | skip (DB assigns) | required | required | applies |
| `'u'` update | required | required | required | applies |
| `'d'` delete | required | required | skip | skip |
| other/unknown | required | required | required | applies |

```go
err := fmt.ValidateFields('u', myFielder)
```

## Conversion and Reading Helpers

### ReadValues()

For consumers that need a `[]any` of values (like ORM for SQL arguments):

```go
vals := fmt.ReadValues(myFielder.Schema(), myFielder.Pointers())
```

### ReadStringPtr()

High-performance codecs can read string values without boxing to `any`:

```go
if val, ok := fmt.ReadStringPtr(ptrs[i]); ok {
    // use val (string) directly
}
```

### isZeroPtr()

Checks if a pointer points to its type's zero value.

```go
if fmt.isZeroPtr(ptr, fieldType) {
    // value is zero
}
```

## Example Implementation

```go
type User struct {
    ID   string
    Name string
}

func (u *User) Schema() []fmt.Field {
    return []fmt.Field{
        {Name: "id", Type: fmt.FieldText, DB: &fmt.FieldDB{PK: true}},
        {Name: "name", Type: fmt.FieldText, NotNull: true, Permitted: fmt.Permitted{Letters: true}},
    }
}

func (u *User) Pointers() []any {
    return []any{&u.ID, &u.Name}
}

func (u *User) Validate(action byte) error {
    return fmt.ValidateFields(action, u)
}
```
