# PLAN: Zero-Alloc Parsing API (tinywasm/fmt)

← [README](../README.md)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass and docs are updated.

## Context

`tinywasm/json` decode tiene 20 allocs — el **71%** viene de `fmt.GetConv`/`fmt.Convert` porque:

1. **`Convert(s ...any)`** — el variadic `...any` boxea el string (2 allocs: `[]any` slice + string boxing).
2. **`Int64()`/`Float64()` no devuelven el Conv al pool** — el Conv se pierde, `sync.Pool.New` crea uno nuevo en cada llamada.
3. **`parseIntBase()`/`parseFloatBase()` usan `GetString(BuffOut)`** — aloca un string nuevo cuando ya existe `GetStringZeroCopy` que es 0 allocs y seguro aquí (lectura local).

**Objetivo:** Permitir que json (y otros consumidores) parseen números desde `[]byte` con **0 allocaciones** usando la infraestructura existente de Conv.

**Downstream:** [json PLAN.md](../../json/docs/PLAN.md)

---

## Stage 1: Add `LoadBytes` method

**File:** `memory.go` (junto a `GetString`, `ResetBuffer`, etc.)

```go
// LoadBytes loads raw bytes into the output buffer for subsequent parsing.
// Reuses existing buffer capacity — 0 allocations for data within capacity.
// Use with GetConv() + LoadBytes + Int64()/Float64() + PutConv() to parse
// numbers from byte slices without string creation or variadic boxing.
func (c *Conv) LoadBytes(b []byte) {
	c.ResetBuffer(BuffOut)
	c.ResetBuffer(BuffErr)
	c.out = append(c.out[:0], b...)
	c.outLen = len(b)
	c.kind = K.String
}
```

**Justificación:** Conv ya tiene buffers pre-alocados con capacidad 64 (del pool). JSON numbers son típicamente <20 bytes. `append(c.out[:0], b...)` reutiliza la capacidad existente — 0 allocs.

---

## Stage 2: Use `GetStringZeroCopy` in parsers

**File:** `num_int.go` (line 190)

```go
// BEFORE:
s := c.GetString(BuffOut)

// AFTER:
s := c.GetStringZeroCopy(BuffOut)
```

**File:** `num_float.go` (line 48)

```go
// BEFORE:
s := c.GetString(BuffOut)

// AFTER:
s := c.GetStringZeroCopy(BuffOut)
```

**Justificación:** `GetStringZeroCopy` usa `unsafe.String` — devuelve un string que comparte el buffer sin alocar. Es seguro porque:
- `parseIntString`/`parseFloatBase` solo **leen** el string carácter a carácter.
- El string no escapa de la función — va out of scope antes de cualquier modificación del buffer.
- `PutConv()` se llama después de leer el resultado, no durante el parsing.

---

## Stage 3: Tests

### 3.1 Test `LoadBytes` + `Int64`

```go
func TestLoadBytesInt64(t *testing.T) {
	c := GetConv()
	c.LoadBytes([]byte("42"))
	v, err := c.Int64()
	c.PutConv()
	if err != nil {
		t.Fatal(err)
	}
	if v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}
```

### 3.2 Test `LoadBytes` + `Float64`

```go
func TestLoadBytesFloat64(t *testing.T) {
	c := GetConv()
	c.LoadBytes([]byte("9.5"))
	v, err := c.Float64()
	c.PutConv()
	if err != nil {
		t.Fatal(err)
	}
	if v != 9.5 {
		t.Errorf("expected 9.5, got %f", v)
	}
}
```

### 3.3 Test negative, scientific notation

```go
func TestLoadBytesNegative(t *testing.T) {
	c := GetConv()
	c.LoadBytes([]byte("-100"))
	v, _ := c.Int64()
	c.PutConv()
	if v != -100 {
		t.Errorf("expected -100, got %d", v)
	}
}

func TestLoadBytesScientific(t *testing.T) {
	c := GetConv()
	c.LoadBytes([]byte("1.5e2"))
	v, _ := c.Float64()
	c.PutConv()
	if v != 150 {
		t.Errorf("expected 150, got %f", v)
	}
}
```

### 3.4 Benchmark zero-alloc path

```go
func BenchmarkLoadBytesInt64(b *testing.B) {
	data := []byte("12345")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := GetConv()
		c.LoadBytes(data)
		c.Int64()
		c.PutConv()
	}
}

func BenchmarkLoadBytesFloat64(b *testing.B) {
	data := []byte("3.14159")
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c := GetConv()
		c.LoadBytes(data)
		c.Float64()
		c.PutConv()
	}
}
```

**Expected:** `0 allocs/op` for both benchmarks (pool reuse after warmup).

### 3.5 Verify existing tests still pass

```bash
gotest
```

---

## Stage 4: Update docs

**File:** `docs/API_PARSING.md` — add section for zero-alloc number parsing:

```markdown
### Zero-Alloc Number Parsing from Bytes

For parsing numbers from byte slices without allocations (e.g., JSON parsers):

​```go
c := fmt.GetConv()
c.LoadBytes(numBytes)      // load bytes into buffer (0 alloc)
v, err := c.Int64()        // or c.Float64()
c.PutConv()                // return Conv to pool
​```

This bypasses `Convert(s ...any)` which boxes the string argument.
The pool ensures Conv reuse after warmup.
```

---

## Stage 5: Publish

```bash
gopush 'fmt: add LoadBytes for zero-alloc number parsing, use GetStringZeroCopy in parsers'
```

---

## Summary

| Stage | File(s) | Change |
|-------|---------|--------|
| 1 | `memory.go` | Add `LoadBytes([]byte)` |
| 2 | `num_int.go:190`, `num_float.go:48` | `GetString` → `GetStringZeroCopy` |
| 3 | tests | `LoadBytes` + `Int64`/`Float64` tests + benchmark |
| 4 | docs | Document zero-alloc parsing pattern |
| 5 | — | `gopush` |

**Total code added:** ~10 lines (1 method + 2 line changes). Reutiliza toda la infraestructura existente.
