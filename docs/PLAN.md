# `fmt.ErrType` — Error Wrapping con `errors.Is` traversal

## Contexto y problema

`fmt.Err(...)` devuelve `*Conv`, un objeto del pool de conversión. Su propósito es crear **valores centinela** (`var ErrNotFound = fmt.Err(...)`) con soporte multilingüe. No está diseñado para ser un error de larga vida con cadena de wrapping.

El resultado es que hoy **ningún consumidor puede hacer esto** de forma idiomática:

```go
return fmt.ErrType(cause, ErrSyncFailed)
// errors.Is(result, ErrSyncFailed) == true  ← actualmente imposible con fmt.Err
```

Cada librería que necesite wrapping (hoy `orm`, mañana `app`, `postgres`, etc.) termina copiando un `syncError`-like struct local. Eso viola el principio de **una sola fuente de verdad** y hace crecer el binario.

---

## Por qué `*Conv` no puede llevar el campo `wrapped`

`*Conv` es gestionado por un `sync.Pool` (backend) o alojado directamente (WASM). El ciclo de vida es:

```
GetConv() → SmartArgs → ... → putConv() → convPool.Put(c)
```

Un `*Conv` puede ser reutilizado en cualquier momento tras `putConv()`. Si añadimos `wrapped error` al struct, ese puntero sobreviviría al `Put`, creando:

1. **Race conditions**: otro goroutine obtiene el objeto del pool con un campo `wrapped` colgado.
2. **Retención de memoria**: el GC no puede liberar el error apuntado mientras `*Conv` viva en el pool.
3. **Semántica rota**: los errores centinela (`var ErrNotFound`) se crean una vez y nunca se devuelven al pool. Si añadimos `wrapped` al struct, sería un campo siempre `nil` en los centinela — peso sin uso.

**Conclusión:** `*Conv` no es el lugar correcto. Necesitamos un tipo separado, liviano, de larga vida.

---

## Propuesta: `fmt.ErrType`

El nombre refleja la intención: el `sentinel` es la **identidad de tipo** del error — `ErrSyncFailed` es el "tipo sync fallido", `ErrNotFound` es el "tipo no encontrado". `ErrType(cause, sentinel)` crea un error que *es de tipo* `sentinel`.

### Nuevo tipo `wrappedErr` (interno)

```go
// wrappedErr es un error de larga vida que une un mensaje con una
// cadena Unwrap() compatible con errors.Is / errors.As (Go 1.20+).
// Es un tipo separado de *Conv — no usa el pool.
type wrappedErr struct {
    msg  string  // mensaje ya construido (sin alojamiento extra)
    errs []error // causa + centinela(s)
}

func (e *wrappedErr) Error() string   { return e.msg }
func (e *wrappedErr) Unwrap() []error { return e.errs }
```

### API pública

```go
// ErrType crea un error que:
//   - muestra: cause.Error() + ": " + sentinel.Error()
//   - permite: errors.Is(result, sentinel) == true
//   - permite: errors.Is(result, cause)    == true  (si cause lo soporta)
//
// El sentinel actúa como la identidad de tipo del error — análogo a la
// categoría, clase o "tipo" al que pertenece el error resultante.
//
// Si cause es nil, retorna sentinel directamente (sin wrapping).
//
// Ejemplo:
//   return fmt.ErrType(dbErr, ErrSyncFailed)
//
func ErrType(cause, sentinel error) error {
    if cause == nil {
        return sentinel
    }
    return &wrappedErr{
        msg:  cause.Error() + ": " + sentinel.Error(),
        errs: []error{cause, sentinel},
    }
}
```

> [!NOTE]
> `ErrType` no usa `GetConv()` ni el pool. Es una asignación única de un struct pequeño (~40 B en 64-bit). El binario no crece con tablas de traducción.

---

## Adición a `API_ERRORS.md`

| Go Standard | fmt Equivalent |
|---|---|
| `errors.New()` | `Err(message)` |
| `fmt.Errorf()` | `Errf(format, args...)` |
| `fmt.Errorf("%w: %w", cause, sentinel)` | `ErrType(cause, sentinel)` ← **nuevo** |

---

## Impacto en binario

| Situación | Bytes extra |
|---|---|
| Sin `ErrType` (cada lib copia su struct) | ~80–120 B por paquete consumidor |
| Con `ErrType` en `fmt` | ~60 B una sola vez en `fmt` |
| Ahorro por paquete adicional (`orm`, `postgres`, `app`…) | ~80–120 B cada uno |

El tipo `wrappedErr` es tan pequeño que el compilador lo puede inlinear. Sin pool, sin goroutines, sin tabla de traducción.

---

## Casos de uso concretos habilitados

### `tinywasm/orm` — migración post-publicación

```go
// Antes (workaround local hoy en errors.go):
type syncError struct { cause, sentinel error }
func (e *syncError) Error() string   { ... }
func (e *syncError) Unwrap() []error { return []error{e.cause, e.sentinel} }
func wrapSyncErr(cause error) error  { return &syncError{...} }

// Después — todo el boilerplate desaparece:
func wrapSyncErr(cause error) error { return fmt.ErrType(cause, ErrSyncFailed) }
```

### `tinywasm/postgres` / `tinywasm/sqlt`

```go
var ErrCompileFailed = fmt.Err("compile", "failed")
return fmt.ErrType(parseErr, ErrCompileFailed)
// errors.Is(err, ErrCompileFailed) == true en el consumidor
```

### `tinywasm/app` — sin cambios en el llamador

```go
if errors.Is(err, orm.ErrSyncFailed) { /* 500 */ }
if errors.Is(err, orm.ErrNotFound)   { /* 404 */ }
```
Hoy esto no funciona cuando el error viene envuelto. Con `ErrType` funciona de forma transparente.

---

## Archivos a modificar en `tinywasm/fmt`

#### [NEW] `error_wrap.go`
Contiene `wrappedErr` + `ErrType`. Sin build tags — idéntico en WASM y backend.

#### [MODIFY] `docs/API_ERRORS.md`
Añade la fila `ErrType` a la tabla y un ejemplo de uso.

---

## Migración en `tinywasm/orm` (post-publicación de `fmt`)

1. Publicar `tinywasm/fmt` con `ErrType`.
2. En `orm/errors.go`: eliminar `syncError` y `wrapSyncErr`.
3. En `orm/sync.go`: `wrapSyncErr(err)` → `fmt.ErrType(err, ErrSyncFailed)` (o mantener el helper de una línea).
4. Bump de versión de `fmt` en `orm/go.mod`.

---

## Verificación

```bash
# En tinywasm/fmt
go test ./... -race -count=1
go test -run TestErrType -v
```

| Caso | Esperado |
|---|---|
| `errors.Is(ErrType(c, s), s)` | `true` |
| `errors.Is(ErrType(c, s), c)` | `true` (cuando `c` es un centinela) |
| `ErrType(c, s).Error()` | `"cause msg: sentinel msg"` |
| `errors.Is(ErrType(c, s), otroErr)` | `false` |
| `ErrType(nil, s)` | retorna `s` directamente |

---

## Pregunta abierta

> [!IMPORTANT]
> **¿`cause == nil` debe retornar `sentinel` directamente, o mejor panics con mensaje claro?**
> La propuesta actual retorna `sentinel` — es la opción más segura en WASM donde no hay stack trace recuperable.
