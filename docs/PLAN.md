# PLAN: Añadir FieldRaw al sistema de tipos de schema

## Problema

El sistema de tipos de `tinywasm/fmt` no tiene forma de expresar que un campo `string` contiene **JSON pre-serializado** que debe emitirse inline, sin quotes.

Actualmente todos los campos string son `FieldText`. Cuando `tinywasm/json` encodea un `FieldText` cuyo valor ya es JSON (ej. `{"tools":[...]}`), lo trata como string opaca y lo quotea → double-encoding:

```
campo contiene: {"tools":[...]}
json emite:     "\"tools\":[...]}"   ← incorrecto
```

Este patrón aparece en cualquier módulo que necesite componer JSON estructurado (MCP, respuestas HTTP, etc.). El problema está en el schema: no existe el tipo semántico correcto para expresarlo.

## Justificación

`FieldType` es el contrato central del ecosistema. Define QUÉ es un campo — no cómo se renderiza ni cómo se almacena. `FieldRaw` es un tipo semántico legítimo: "este campo es JSON válido pre-serializado, trátalo como valor JSON, no como string".

Añadirlo en `tinywasm/fmt` es correcto porque:
- `tinywasm/json` lo consume para decidir cómo encodear
- `tinywasm/orm` lo consume para decidir qué generar
- No añade dependencias nuevas
- El resto del ecosistema lo adopta naturalmente via regeneración de ORM

## Fix

En `field.go`, añadir `FieldRaw` al enum y a `fieldTypeNames`:

```go
const (
    FieldText      FieldType = iota // Any string
    FieldInt                        // Any integer
    FieldFloat                      // Any float
    FieldBool                       // Boolean
    FieldBlob                       // Binary data ([]byte)
    FieldStruct                     // Nested struct (implements Fielder)
    FieldIntSlice                   // []int
    FieldStructSlice                // []Fielder
    FieldRaw                        // Pre-serialized JSON — emitted inline, no quoting
)

var fieldTypeNames = []string{"text", "int", "float", "bool", "blob", "struct", "intslice", "structslice", "raw"}
```

## Impacto en downstream

| Módulo | Cambio requerido |
|---|---|
| `tinywasm/json` | En encode: campo `FieldRaw` → emitir el string value como bytes sin quotes. En decode: leer JSON value y guardarlo como string. |
| `tinywasm/orm` | Reconocer tag `json:",raw"` en un campo string → generar `fmt.FieldRaw` en el schema. |
| `tinywasm/mcp` | Añadir `json:",raw"` a campos `Result`, `Error`, `Tools` → regenerar `model_orm.go`. |

Cada módulo tiene su propio PLAN con los detalles de su cambio.
