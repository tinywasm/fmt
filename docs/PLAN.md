# PLAN: Add FieldIntSlice to tinywasm/fmt

## Bug
`tinywasm/fmt` lacks a `FieldType` for integer slices (`[]int`).
This prevents `tinywasm/json` from encoding/decoding JSON arrays like `[600,500,400]`.

**Real-world impact**: `tinywasm/pdf` font definitions have `Cw []int` (character widths, up to 256+ entries). Without `FieldIntSlice`, fonts cannot be loaded via `tinywasm/json`.

---

## Stage 1: Add FieldIntSlice constant

**File**: `field.go`

**What to do**:
1. Add `FieldIntSlice` to `FieldType` constants:
   ```go
   const (
       FieldText   FieldType = iota
       FieldInt
       FieldFloat
       FieldBool
       FieldBlob
       FieldStruct
       FieldIntSlice  // []int
   )
   ```
2. Update `fieldTypeNames` slice to include `"intslice"`:
   ```go
   var fieldTypeNames = []string{"text", "int", "float", "bool", "blob", "struct", "intslice"}
   ```

---

## Stage 2: Update isZeroPtr for FieldIntSlice

**File**: `field.go`

**What to do**:
Add case in `isZeroPtr()`:
```go
case FieldIntSlice:
    if p, ok := ptr.(*[]int); ok {
        return len(*p) == 0
    }
```

---

## Stage 3: Update ReadValues for FieldIntSlice

**File**: `field.go`

**What to do**:
Add case in `ReadValues()` if it handles all field types (verify the switch):
```go
case FieldIntSlice:
    if p, ok := ptr.(*[]int); ok {
        values[i] = *p
    }
```

---

## Stage 4: Update tests

**File**: `field_test.go`

**What to do**:
1. Add `FieldIntSlice` to `TestFieldTypeString`:
   ```go
   {FieldIntSlice, "intslice"},
   ```
2. Add `isZeroPtr` test for `FieldIntSlice`:
   ```go
   var sl []int
   if !isZeroPtr(&sl, FieldIntSlice) { t.Error("nil slice should be zero") }
   sl = []int{1}
   if isZeroPtr(&sl, FieldIntSlice) { t.Error("non-empty slice should not be zero") }
   ```

---

## Validation
1. `go test ./...` passes
2. `gotest` (WASM) passes
