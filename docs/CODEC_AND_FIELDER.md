# Codec tipado vs. `Field`/`Fielder` — separación de responsabilidades

> **Estado:** IMPLEMENTADO. Define la arquitectura del codec tipado (`Encodable`/`Decodable`)
> y su relación con `Field`/`Fielder`.

## Principio

> **Una sola forma de describir el esquema** = `Field` / `Fielder` (Schema/Pointers).
> **Una sola forma de serializar** = el codec tipado (`Encodable`/`Decodable` + `FieldWriter`/
> `FieldReader`).
> No compiten: cubren operaciones **distintas**, cada una en su carril.

`Field` **no desaparece** con el codec. El codec reemplaza a `Field`/`Pointers()` **solo en el
rol de serializar**; todo lo demás sigue usando `Field`.

## Quién usa qué (y para qué)

| Operación | Mecanismo | Quién | ¿El codec lo reemplaza? |
|---|---|---|---|
| **DDL SQL** (`CREATE TABLE`, columnas, PK/Unique/AutoInc vía `Field.DB`) | `Field`/`Schema()` | `sqlt`, `postgres`, `orm`, `indexdb` | ❌ No — el codec solo escribe valores, no conoce columnas ni PKs |
| **Validación** (`Field.Permitted`, `Field.Type`) | `Field`/`Schema()` → `ValidateFields` | `orm`, `form` | ❌ No — el codec no valida |
| **UI / formularios** (`Field.Widget`, seteado por `ormc` del tag `input:`) | `Field`/`Schema()` | `form` | ❌ No — el codec no sabe de widgets |
| **Scan posicional SQL** (`row.Scan(Pointers()...)`) | `Pointers() []any` | `orm/qb.go` | ❌ No — `database/sql` exige punteros en orden de columna |
| **Serialización** (JSON / JS: nombre + valor por campo) | **codec** (`EncodeFields`/`DecodeFields`) | `json`, `jsvalue` | ✅ **Sí** — deja de usar `Pointers()` y pasa al codec |

Solo la última fila migra. `Field`, `Schema()` y `Pointers()` siguen siendo la columna vertebral
del resto.

## Roles finales

- **`Field` / `Schema()`** → metadata de esquema (DDL, validación, UI). Se queda. Es 0-alloc
  (`Schema()` devuelve un `[]Field` global `_schemaX`, no asigna por llamada).
- **`Pointers() []any`** → acceso posicional a valores para `row.Scan(...)` de SQL. Se queda
  (es la API que exige el driver; un visitor por-nombre no la reemplaza).
- **`Encodable` / `Decodable` (codec)** → serialización tipada **0-alloc, map-free, sin `any`**
  (JSON en `json`, JS en `jsvalue`). Nuevo. Reemplaza el uso de `Pointers()` **solo para
  serializar**.

## ¿No son "dos formas"?

No. No son dos formas de hacer **lo mismo**; son herramientas para operaciones **distintas**:

- `Field`/`Fielder` = *"describir el esquema y dar punteros a la DB"*.
- codec = *"serializar valores a un formato"*.

La única superposición es que ambos conocen los **nombres de campo**, pero `ormc` los genera de
la **misma fuente** (el struct Go + tags). Es como tener tags `db:` y `json:` en un struct: una
sola fuente de verdad, dos proyecciones generadas. No es duplicación a mano ni deuda.

## `OmitEmpty`

Hoy `json` evalúa `Field.OmitEmpty` en runtime. Con el codec, `ormc` resuelve el `omitempty` en
**tiempo de generación** y emite la condición directamente:

```go
func (m *User) EncodeFields(w fmt.FieldWriter) {
    if m.Name != "" { w.String("name", m.Name) } // omitempty resuelto en codegen
    w.Int("age", int64(m.Age))
}
```

Más rápido y 0-alloc (no se evalúa metadata en runtime). El directivo sigue declarándose en el
tag; `ormc` lo conoce.

## Referencias

- Contrato del codec: `docs/API_CODEC.md` (interfaces `FieldWriter`/`FieldReader`/`Encodable`/
  `Decodable`).
- Plan de `fmt` (contrato): `docs/PLAN.md`.
- Consumidores: `orm/docs/PLAN.md` (codegen), `json/docs/PLAN.md` (JSON), `jsvalue/docs/PLAN.md`
  (JS).
- Orquestador: `~/Dev/Project/tinywasm/docs/SIZE_OPTIMIZATION_MASTER_PLAN.md`.
