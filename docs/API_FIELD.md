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
    PK        bool
    Unique    bool
    NotNull   bool
    AutoInc   bool
    OmitEmpty bool      // omit from JSON when zero value
    Permitted           // embedded: validation rules (characters, min/max)
}
```

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

Checks a string value against the field's constraints (`NotNull` and `Permitted`).

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
        {Name: "id", Type: fmt.FieldText, PK: true},
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
