# PLAN: fmt — ValidateFields con action byte

← [README](../README.md)

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** If exceeded, subdivide by domain.
- **Flat hierarchy.** No subdirectories for library code.
- **Documentation First:** Update docs before coding.

## Contexto

`ValidateFields` no tiene noción de la acción CRUD.
Con `action byte` decide **cuáles campos evaluar** según la operación.

`Field.Validate(value)` **no cambia** — valida calidad del dato (NotNull, Permitted).
La action es lógica de orquestación, no de validación individual.

## Reglas por action en ValidateFields

| Action | PK + AutoInc | PK sin AutoInc | NotNull | Permitted |
|--------|-------------|----------------|---------|-----------|
| `'c'` create | skip (DB asigna) | requerido | requerido | aplica |
| `'u'` update | requerido | requerido | requerido | aplica |
| `'d'` delete | requerido | requerido | skip | skip |
| otro/desconocido | requerido | requerido | requerido | aplica |

No hay caso especial para `'r'` — reads no pasan por `ValidateFields`.
Si llega un action desconocido, se valida con las reglas más estrictas (`'u'`).

---

## Stage 1: Actualizar `ValidateFields` — agregar `action byte`

**File:** `field.go`

`Field.Validate(value string) error` no cambia. Solo cambia `ValidateFields`:

```go
// ANTES:
func ValidateFields(f Fielder) error {
    schema := f.Schema()
    ptrs := f.Pointers()
    for i, field := range schema {
        switch field.Type {
        case FieldText:
            val, _ := ReadStringPtr(ptrs[i])
            if err := field.Validate(val); err != nil {
                return err
            }
        case FieldStruct:
            if validator, ok := ptrs[i].(Validator); ok {
                if err := validator.Validate(); err != nil {
                    return err
                }
            } else if fielder, ok := ptrs[i].(Fielder); ok {
                if err := ValidateFields(fielder); err != nil {
                    return err
                }
            }
        default:
            if field.NotNull && isZeroPtr(ptrs[i], field.Type) {
                return Err(field.Name, "required")
            }
        }
    }
    return nil
}

// DESPUÉS:
func ValidateFields(action byte, f Fielder) error {
    schema := f.Schema()
    ptrs := f.Pointers()
    for i, field := range schema {
        // 'd' delete: solo PK requerido, skip todo lo demás
        if action == 'd' {
            if field.PK {
                switch field.Type {
                case FieldText:
                    val, _ := ReadStringPtr(ptrs[i])
                    if val == "" {
                        return Err(field.Name, "required")
                    }
                default:
                    if isZeroPtr(ptrs[i], field.Type) {
                        return Err(field.Name, "required")
                    }
                }
            }
            continue
        }

        // 'c' create: skip PK+AutoInc (DB asigna)
        if action == 'c' && field.PK && field.AutoInc {
            continue
        }

        switch field.Type {
        case FieldText:
            val, _ := ReadStringPtr(ptrs[i])

            // PK siempre requerido (en 'c' sin AutoInc, en 'u', y cualquier otro)
            if field.PK && val == "" {
                return Err(field.Name, "required")
            }

            if err := field.Validate(val); err != nil {
                return err
            }

        case FieldStruct:
            if validator, ok := ptrs[i].(Validator); ok {
                if err := validator.Validate(action); err != nil {
                    return err
                }
            } else if fielder, ok := ptrs[i].(Fielder); ok {
                if err := ValidateFields(action, fielder); err != nil {
                    return err
                }
            }

        default:
            // PK siempre requerido
            if field.PK && isZeroPtr(ptrs[i], field.Type) {
                return Err(field.Name, "required")
            }
            if field.NotNull && isZeroPtr(ptrs[i], field.Type) {
                return Err(field.Name, "required")
            }
        }
    }
    return nil
}
```

---

## Stage 2: Actualizar interfaz `Validator`

**File:** `field.go`

```go
// ANTES:
type Validator interface {
    Validate() error
}

// DESPUÉS:
type Validator interface {
    Validate(action byte) error
}
```

`SafeFielder` no cambia — solo combina `Fielder` + `Validator`.

---

## Stage 3: Actualizar tests

- Actualizar todos los tests que llaman `ValidateFields(data)` → `ValidateFields('c', data)`
- Agregar tests por action:
  - `'c'`: PK+AutoInc omitido, PK sin AutoInc requerido, NotNull requerido, Permitted aplica
  - `'u'`: PK requerido, NotNull requerido, Permitted aplica
  - `'d'`: solo PK requerido, todo lo demás skip
  - Action desconocido (ej: `'x'`): mismas reglas que `'u'`

```bash
gotest
```

---

## Stage 4: Actualizar documentación

**File:** `docs/API_FIELD.md`

- Actualizar `Validator` interface: `Validate() error` → `Validate(action byte) error`
- Actualizar `ValidateFields()`: `ValidateFields(f)` → `ValidateFields(action, f)`
- Actualizar ejemplo de uso:

```go
// ANTES:
func (u *User) Validate() error {
    return fmt.ValidateFields(u)
}

// DESPUÉS:
func (u *User) Validate(action byte) error {
    return fmt.ValidateFields(action, u)
}
```

- Agregar tabla de reglas por action en la sección `ValidateFields() Helper`
- Documentar que `Field.Validate(value)` no cambia — la action es orquestación---

## Stage 5: Interfaz `Model` — estándar agnóstico

**File:** `field.go`

Mover la interfaz `Model` desde `orm` a `fmt` para que sea el estándar agnóstico
utilizado por todas las capas (DB, Form, API). Un `Model` es un `Fielder` que
tiene un nombre identificador (`ModelName`).

```go
type Model interface {
    Fielder
    ModelName() string
}
```

**Justificación:**
- `fmt` ya define la estructura (`Fielder`) y validación (`Validator`).
- `ModelName` es el metadato mínimo necesario para identificar el recurso en cualquier transporte.
- Al estar en `fmt`, el paquete `orm` y `form` pueden depender de una interfaz común sin crear ciclos.
