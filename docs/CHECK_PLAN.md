# PLAN: Fix `%q` Verb — Missing String Escape in `Sprintf`

## Principio de diseño de `tinywasm/fmt`

`tinywasm/fmt` es una librería **minimalista** orientada a tamaño de binario mínimo.
Reglas estrictas:
- **Cero código muerto**: cada función debe ser utilizada; sin helpers que no se referencien
- **Reutilización máxima**: si la lógica ya existe, se invoca — no se duplica
- **Sin imports externos** para funcionalidad core

El fix de este bug debe reutilizar `Conv.Quote()` que ya existe en `quote.go`.
No se crea ninguna función nueva.

---

## Bug Confirmado

**Test que replica el bug:** `fmt_q_escape_test.go` → `TestSprintfQ_EscapesSpecialChars`

Ejecutar para ver el fallo actual:
```
go test -run TestSprintfQ_EscapesSpecialChars -v
```

Output real del fallo (bug confirmado):
```
--- FAIL: TestSprintfQ_EscapesSpecialChars/double_quotes_must_be_escaped
    got:  "{"key":"value"}"       ← JSON inválido, comillas sin escapar
    want: "{\"key\":\"value\"}"

--- FAIL: TestSprintfQ_EscapesSpecialChars/json_array_payload_(real_MCP_state_bug)
    got:  "[{"tab_title":"BUILD",...}]"   ← causó footer TUI vacío
    want: "[{\"tab_title\":\"BUILD\",...}]"
```

**El test debe pasar al 100% después del fix. Es la condición de aceptación.**

---

## Root Cause

`fmt_template.go` ~línea 373–377:

```go
case 'q':
    switch v := arg.(type) {
    case string:
        return "\"" + v + "\""  // BUG: concatenación cruda, sin escapar
```

`Conv.Quote()` en `quote.go` ya escapa correctamente `"→\"`, `\→\\`, `\n→\n`,
`\r→\r`, `\t→\t`. El `%q` no la usa — es código duplicado e incorrecto.

## Impacto Real

`tinywasm/app/daemon.go` usa `fmt.Sprintf(..., %q, stateJSON)` para construir
la respuesta `tinywasm/state`. `stateJSON` es un array JSON con comillas internas.
El `%q` roto produce JSON inválido → `tinywasm/json.Decode` falla silenciosamente
→ `reconstructRemoteHandlers` nunca se llama → **footer TUI permanece vacío**.

---

## Fix (una línea, reutiliza `Conv.Quote()`)

**Archivo:** `fmt_template.go` ~línea 376

```go
// Antes
case string:
    return "\"" + v + "\""

// Después
case string:
    return Convert(v).Quote().String()
```

`Convert(v).Quote()` usa el código existente en `quote.go`. Sin nueva función,
sin nuevo import, sin código muerto. El tamaño de binario no aumenta porque
`Conv.Quote()` ya es referenciada por otros callers.

---

## Archivo Afectado

| Archivo | Línea | Cambio |
|---|---|---|
| `fmt_template.go` | ~376 | `return "\"" + v + "\""` → `return Convert(v).Quote().String()` |

Un solo cambio de una línea.

---

## Condición de Aceptación

El agente que ejecute este plan debe:

1. Aplicar el fix de una línea en `fmt_template.go`
2. Ejecutar: `go test -run TestSprintfQ_EscapesSpecialChars -v`
   → **todos los sub-tests deben pasar (PASS)**
3. Ejecutar: `go test ./...`
   → **ningún test existente debe romperse**

El test `fmt_q_escape_test.go` ya existe y falla con el bug. No hay que crearlo.
