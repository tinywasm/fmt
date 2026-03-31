# Plan: Separate DB metadata into FieldDB struct

## Problem

`fmt.Field` is a god-struct mixing concerns from 3 layers:
- **DB**: PK, Unique, AutoInc (only relevant to orm, sqlite, postgres, indexdb)
- **Form/UI**: Widget (only relevant to form)
- **Universal**: Name, Type, NotNull, OmitEmpty, Permitted (used by all)

`formonly` structs carry `PK: false, AutoInc: false` — meaningless metadata. DB structs carry `Widget: nil`. The consumer cannot tell which fields are semantically relevant.

## Solution

Extract DB-only flags into `FieldDB` struct. Field gets `DB *FieldDB` pointer (nil when not a DB struct).

## Changes to field.go

### New type

```go
type FieldDB struct {
    PK      bool
    Unique  bool
    AutoInc bool
}
```

### Updated Field struct

```go
type Field struct {
    Name      string
    Type      FieldType
    NotNull   bool
    OmitEmpty bool
    Widget    Widget
    Permitted
    DB *FieldDB  // nil for formonly/transport structs
}
```

Remove from Field: `PK`, `Unique`, `AutoInc` (moved to FieldDB).

### Helper methods on Field (migration aid)

To avoid nil-check boilerplate in every consumer, add convenience methods:

```go
func (f Field) IsPK() bool      { return f.DB != nil && f.DB.PK }
func (f Field) IsUnique() bool   { return f.DB != nil && f.DB.Unique }
func (f Field) IsAutoInc() bool  { return f.DB != nil && f.DB.AutoInc }
```

### Update ValidateFields

Replace direct field access:
- `field.PK` → `field.IsPK()`
- `field.AutoInc` → `field.IsAutoInc()`

Affected lines in ValidateFields: the delete branch (PK check), the create skip (PK+AutoInc), and the PK-required checks.

### Update field_test.go

Update tests that assert `f.PK`, `f.Unique`, `f.AutoInc` to use either `f.DB.PK` or `f.IsPK()`.

## Files affected in tinywasm/fmt

- field.go (struct change + helpers + ValidateFields update)
- field_test.go (test assertions)

## Consumer impact (other modules — see their own plans)

| Module | Files | Change |
|--------|-------|--------|
| tinywasm/orm | ormc.go, ormc_generate.go, db.go | Generate `DB: &fmt.FieldDB{...}` instead of flat fields |
| tinywasm/form | form.go, test files | `field.PK` → `field.IsPK()` |
| tinywasm/sqlite | translate.go, tests | `f.PK` → `f.IsPK()` |
| tinywasm/postgres | translate.go | `f.PK` → `f.IsPK()` |
| tinywasm/indexdb | adapter.go, tests | `f.PK` → `f.IsPK()` |
| tinywasm/user | models_orm.go (regenerate) | Generated code uses new format |
| tinywasm/mcp | model_orm.go (regenerate) | Generated code uses new format |
| tinywasm/skills | models_orm.go (regenerate) | Generated code uses new format |

## Execution order

1. Add `FieldDB` struct and helper methods to field.go
2. Move PK, Unique, AutoInc from Field to FieldDB
3. Update ValidateFields to use helpers
4. Update field_test.go
5. `go test ./...` in fmt
6. Publish fmt (consumers depend on this)
