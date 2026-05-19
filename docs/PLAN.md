# PLAN: tinywasm/fmt — Permitted validation fix + tag-driven tilde control

## Problema raíz

`Field.hasPermittedRules()` retorna `true` cuando `Minimum > 0`, aunque ningún flag de
caracteres esté activo. Esto provoca que `Permitted.Validate` se ejecute con todos los flags
en `false`, rechazando **cualquier** carácter como inválido.

### Caso concreto (goflare-demo)

Schema generado por ormc:
```go
{Name: "nombre", Widget: input.Text(), Permitted: fmt.Permitted{Minimum: 2}}
```

Flujo en `Field.Validate("María")`:
1. `Widget.Validate("María")` → ok (Text() tiene Letters=true, Tilde=true)
2. `hasPermittedRules()` → **true** (Minimum=2 > 0)
3. `Permitted.Validate("nombre", "María")` → **error: character not allowed M**
   - porque el `Permitted` del Field tiene Letters=false, Tilde=false, etc.

El usuario ve 422 Unprocessable Entity aunque el dato es válido.

---

## Decisión de diseño (opción B elegida)

**Separar validación de longitud de validación de caracteres en `hasPermittedRules()`.**

`hasPermittedRules()` solo debe retornar `true` cuando hay flags de caracteres explícitos.
La longitud (Minimum/Maximum) siempre se valida dentro de `Permitted.Validate` si están
configuradas — sin necesitar que `hasPermittedRules()` lo active.

### Cambio en `field.go`

```go
// Antes
func (f Field) hasPermittedRules() bool {
    return f.Letters || f.Tilde || f.Numbers || f.Spaces ||
        f.BreakLine || f.Tab || len(f.Extra) > 0 ||
        len(f.NotAllowed) > 0 || f.Minimum > 0 || f.Maximum > 0 ||
        f.StartWith != nil
}

// Después
func (f Field) hasPermittedRules() bool {
    return f.Letters || f.Tilde || f.Numbers || f.Spaces ||
        f.BreakLine || f.Tab || len(f.Extra) > 0 ||
        len(f.NotAllowed) > 0 || f.StartWith != nil
    // Minimum/Maximum NO activan char-validation: solo implican chequeo de longitud,
    // que ya ocurre dentro de Permitted.Validate cuando esos campos son > 0.
}
```

Y en `Field.Validate`, mover la verificación de longitud antes del guard:

```go
func (f Field) Validate(value string) error {
    if f.NotNull && value == "" {
        return Err(f.Name, "required")
    }
    if value == "" {
        return nil
    }
    // Longitud: siempre verificar si está configurada, independientemente de char rules
    if f.Minimum > 0 || f.Maximum > 0 {
        if err := f.Permitted.validateLength(f.Name, value); err != nil {
            return err
        }
    }
    if f.Widget != nil {
        if err := f.Widget.Validate(value); err != nil {
            return err
        }
    }
    if f.hasPermittedRules() {
        return f.Permitted.validateChars(f.Name, value)
    }
    return nil
}
```

Esto implica dividir `Permitted.Validate` en dos métodos internos:
- `validateLength(field, text) error` — solo Minimum/Maximum
- `validateChars(field, text) error` — NotAllowed + StartWith + char-by-char

El método público `Validate` mantiene su contrato llamando ambos.

---

## Configuración de tilde via etiquetas de struct

### Objetivo

Permitir que `input.Text()` sea configurable desde la etiqueta del campo en el struct:

```go
// Por defecto (tilde permitida — español es el caso común)
Nombre string `input:"required,min=2"`

// Opt-out explícito (e.g. campos técnicos: usernames, códigos)
Username string `input:"required,min=3,notilde"`
```

### Dónde vive la lógica

La etiqueta `input:"..."` es parseada por `tinywasm/form` al registrar campos.
El resultado configura el widget. Por tanto:

- `tinywasm/fmt`: expone `Permitted` con `Tilde bool` (ya existe) — sin cambios aquí.
- `tinywasm/form`: parsea etiqueta `notilde` y llama `widget.SetTilde(false)` si el widget lo soporta.
- `tinywasm/form/input`: `text` expone `SetTilde(bool)` para que `form` lo configure.

**Responsabilidad de tinywasm/fmt**: mantener `Permitted` limpio, documentar que `Tilde bool`
permite/desactiva tildes, nada más. El parsing de etiquetas es responsabilidad de `tinywasm/form`.

---

## Documentación de `Permitted`

Agregar en `permitted.go` un comentario de referencia rápida:

```
// Regla: si solo Minimum/Maximum están configurados (sin flags de chars),
// Validate solo verifica longitud — nunca rechaza caracteres.
// Para restringir caracteres, habilitar al menos un flag (Letters, Numbers, etc.)
// o agregar entradas en Extra/NotAllowed.
```

---

## Archivos a modificar

| Archivo | Cambio |
|---------|--------|
| `field.go` | `hasPermittedRules()` excluye Minimum/Maximum; `Validate` verifica longitud por separado |
| `permitted.go` | Añadir `validateLength` y `validateChars` internos; actualizar comentario |

## Tests a agregar

- `field_validate_test.go`: campo con solo `Minimum>0` + `Widget=input.Text()` acepta letras y tildes
- `permitted_test.go`: `Permitted{Minimum:2}.Validate("ab")` → ok; `Permitted{}.Validate("x")` → ok (sin reglas = todo permitido)

## Orden de ejecución

1. Modificar `permitted.go` (añadir métodos internos)
2. Modificar `field.go` (ajustar `hasPermittedRules` y `Validate`)
3. Agregar tests
4. Publicar via `gopush`
