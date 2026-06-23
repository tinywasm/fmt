# PLAN — Corregir `codec_test.go` y agregar `Raw()` a las interfaces de codec

> **Repo:** `github.com/tinywasm/fmt`
> **Archivos:** `codec.go`, `codec_test.go`
> **Tipo:** bug fix en tests + extensión de interfaces
> **Prerequisito:** ninguno — primer eslabón de la cadena
> **Verificación:** `gotest`

## Errores actuales

```
codec_test.go:81: mockFieldWriter.Array tiene Array(string, int, func(int, ArrayWriter))
                   pero la interfaz exige  Array(string, int) ArrayWriter
codec_test.go:265: w.Array("tags", len(u.Tags), func(...)) — firma con callback obsoleta
```

`codec_test.go` usa el API viejo de `Array` basado en callback. La interfaz
`FieldWriter` ya fue actualizada al API de retorno `Array(name, n) ArrayWriter`,
pero el archivo de test no se actualizó.

Adicionalmente, `ArrayWriter` no tiene `Close()` en su interfaz, pero
`jsonArrayWriter` (en `tinywasm/json`) sí lo implementa. Sin `Close()` en la
interfaz, el caller no puede cerrar el array de forma portátil.

---

## Fix 1 — Agregar `Close()` a `ArrayWriter` en `codec.go`

```go
type ArrayWriter interface {
    String(val string)
    Int(val int64)
    Float(val float64)
    Bool(val bool)
    Bytes(val []byte)
    Object(val Encodable)
    Close()   // ← NUEVO: finaliza el array (escribe ']' en JSON, libera pool)
}
```

---

## Fix 2 — Actualizar `codec_test.go`

### 2a — `mockFieldWriter.Array` — cambiar firma a la nueva API

```go
// ANTES (firma vieja con callback):
func (m *mockFieldWriter) Array(name string, n int, each func(i int, a ArrayWriter)) {
    ...
}

// DESPUÉS (nueva API: devuelve ArrayWriter):
func (m *mockFieldWriter) Array(name string, n int) ArrayWriter {
    m.buf.WriteString(name)
    m.buf.WriteString("=[")
    return &mockArrayWriter{m: m}
}
```

### 2b — `mockArrayWriter` — agregar `Close()`

```go
func (a *mockArrayWriter) Close() {
    a.m.buf.WriteString("];")
}
```

Notar que `mockArrayWriter` también debe emitir comas entre elementos. Agregar
campo `first bool` e inicializarlo en `true`:

```go
type mockArrayWriter struct {
    m     *mockFieldWriter
    first bool
}

func newMockArrayWriter(m *mockFieldWriter) *mockArrayWriter {
    return &mockArrayWriter{m: m, first: true}
}

func (a *mockArrayWriter) maybeComma() {
    if !a.first {
        a.m.buf.WriteByte(',')
    }
    a.first = false
}
```

Actualizar cada método (`String`, `Int`, etc.) para llamar `a.maybeComma()` primero.

### 2c — `sampleUser.EncodeFields` — actualizar al nuevo API

```go
func (u *sampleUser) EncodeFields(w FieldWriter) {
    w.String("name", u.Name)
    w.Int("age", int64(u.Age))
    if len(u.Tags) > 0 {
        aw := w.Array("tags", len(u.Tags))
        for _, tag := range u.Tags {
            aw.String(tag)
        }
        aw.Close()
    }
}
```

---

## Fix 3 — Agregar `Raw()` a las interfaces en `codec.go`

### Agregar `Raw` a `FieldWriter`

```go
type FieldWriter interface {
    String(name, val string)
    Int(name string, val int64)
    Uint(name string, val uint64)
    Float(name string, val float64)
    Bool(name string, val bool)
    Bytes(name string, val []byte)
    Null(name string)
    Raw(name, val string)             // ← NUEVO: emite val inline sin escaping
    Object(name string, val Encodable)
    Array(name string, n int) ArrayWriter
}
```

### Agregar `Raw` a `FieldReader`

```go
type FieldReader interface {
    String(name string) (string, bool)
    Int(name string) (int64, bool)
    Uint(name string) (uint64, bool)
    Float(name string) (float64, bool)
    Bool(name string) (bool, bool)
    Bytes(name string) ([]byte, bool)
    Object(name string, into Decodable) bool
    Array(name string) (ArrayReader, bool)
    Raw(name string) (string, bool)   // ← NUEVO: devuelve el valor JSON crudo
}
```

---

## Fix 4 — Actualizar mocks en `codec_test.go` para implementar `Raw()`

### `mockFieldWriter.Raw`

```go
func (m *mockFieldWriter) Raw(name, val string) {
    m.buf.WriteString(name)
    m.buf.WriteByte('=')
    m.buf.WriteString(val)
    m.buf.WriteByte(';')
}
```

### `mockFieldReader.Raw`

```go
func (r *mockFieldReader) Raw(name string) (string, bool) {
    return r.String(name)  // el mock trata raw igual que string
}
```

---

## Verificación

```bash
cd ~/Dev/Project/tinywasm/fmt
gotest
```

Todos los tests deben pasar.

## Checklist

- [ ] `Close()` agregado a `ArrayWriter` en `codec.go`
- [ ] `Raw(name, val string)` agregado a `FieldWriter` en `codec.go`
- [ ] `Raw(name string) (string, bool)` agregado a `FieldReader` en `codec.go`
- [ ] `mockFieldWriter.Array` usa nueva firma `(name, n) ArrayWriter`
- [ ] `mockArrayWriter.Close()` implementado
- [ ] `mockArrayWriter` emite comas entre elementos (campo `first bool`)
- [ ] `sampleUser.EncodeFields` usa `aw := w.Array(...)` + loop + `aw.Close()`
- [ ] `mockFieldWriter.Raw()` implementado
- [ ] `mockFieldReader.Raw()` implementado
- [ ] `gotest` verde
