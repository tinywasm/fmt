---
PLAN: "fix: Sprint(error) devolvía cadena vacía y borraba todo diagnóstico"
EXECUTOR: jules
REVIEWER: none
STATUS: DONE
---

> ✅ **EJECUTADO** — `sprint.go` con la rama de error + `sprint_error_test.go`.
> Barrido `Convert(...).String()`: `strings.go`, `language.go` y `fmt_template.go`
> reciben tipos concretos (string/rune) — no aplica; `lang/translation.go:92`
> (subpackage aparte, interpolación de mensajes traducidos) queda sin tocar por el
> radio mínimo del plan (criterio 5: solo sprint.go + tests). gotest: vet ✅,
> race ✅, tests ✅, wasm ✅.

> Plan autocontenido: todo lo necesario para ejecutarlo está aquí.
> Se despacha con el flujo CodeJob. Ver skill: agents-workflow.
>
> ⚠️ **`docs/PLAN.md` de este repo está ocupado** por un plan de rendimiento en
> `JSONEscape` marcado *low priority*. Este plan es **urgente** y vive aparte.
> Para despacharlo hay que cerrar o mover ese otro primero.
>
> Reglas del repo: `AGENTS.md` en la raíz — léelo antes de tocar nada.

# Plan — `fmt.Sprint(err)` debe imprimir el error

## 1. El defecto, medido

```go
err := errors.New("web/public está versionado — el entregable corre peligro")

fmt.Sprint(err)        // ""                                    ← BUG
fmt.Sprintf("%v", err) // "web/public está versionado — …"      ← correcto
```

Reproducible con este programa contra `github.com/tinywasm/fmt v0.25.5`:

```go
package main

import (
	"errors"
	stdfmt "fmt"
	twfmt "github.com/tinywasm/fmt"
)

func main() {
	err := errors.New("boom")
	stdfmt.Printf("Sprint  = %q\n", twfmt.Sprint(err))       // imprime ""
	stdfmt.Printf("Sprintf = %q\n", twfmt.Sprintf("%v", err)) // imprime "boom"
}
```

## 2. Por qué pasa

Tres líneas encadenadas, todas en este repo:

```go
// sprint.go:8
func Sprint(v any) string {
	return Convert(v).String()
}

// convert.go:39-42 — un error va al buffer de ERROR, no al de salida
if _, isError := val.(error); isError {
	return c.wrErr(val.(error).Error())
}

// convert.go:186-191 — y String() devuelve "" cuando ese buffer tiene contenido
func (c *Conv) String() string {
	if c.hasContent(BuffErr) {
		c.putConv()
		return ""      // ← aquí muere el mensaje
	}
	...
}
```

El contrato de `Conv` es deliberado: el error se recupera con `StringErr()`.
El defecto **no** está en `Conv`; está en que `Sprint` —cuyo trabajo es
convertir cualquier valor a texto para mostrarlo— se apoya en él sin tratar el
caso más frecuente que va a recibir: un `error`.

## 3. El daño aguas abajo (contexto, no se arregla aquí)

`tinywasm/app` construye **todas** las líneas del TUI y del canal SSE así
(`app/logs.go:45`, `SprintMessages`):

```go
for i, m := range messages {
	if i > 0 { res += " " }
	res += fmt.Sprint(m)     // ← cada error del ecosistema se convierte en ""
}
```

Resultado real observado en el arranque del demonio: una línea de log
**totalmente vacía** donde debía ir el aviso de que el entregable de release
estaba en peligro. El patrón `logger("algo falló:", err)` está por todo el
ecosistema; hoy publica el prefijo y tira el motivo.

## 4. El cambio

**Un solo archivo: `sprint.go`.**

```go
// Sprint convierte v en texto para mostrarlo.
//
// El caso error se trata aquí y no en Convert: Convert mete el mensaje en el
// buffer de error a propósito (String() devuelve "" y el texto se recupera con
// StringErr()), un contrato que otros llamadores usan. Sprint tiene el trabajo
// contrario —imprimir lo que le den—, y lo que más le llega es un error.
func Sprint(v any) string {
	if err, ok := v.(error); ok && err != nil {
		return err.Error()
	}
	return Convert(v).String()
}
```

### Prohibido

- **NO** toques `Conv.String()` (`convert.go:186`). Devolver el mensaje de error
  ahí cambia el contrato de `Convert`/`StringErr` y rompe a todo el que hoy usa
  `String() == ""` como señal de fallo. El radio de impacto es todo el
  ecosistema.
- **NO** toques `Convert` (`convert.go:39-42`).
- **NO** cambies `Sprintf`: ya funciona (`%v` con error da el mensaje) y hay que
  mantenerlo así.

### Si `Sprint` tiene hermanos con el mismo camino

Comprueba y aplica el mismo tratamiento **solo** a las funciones cuyo trabajo
sea imprimir un valor suelto:

```sh
grep -rn "Convert(.*)\.String()" --include="*.go" . | grep -v tests/
```

Cada acierto fuera de `sprint.go`: decide si esa función existe para *mostrar*
(entonces trátale el error igual) o para *convertir con validación* (entonces
déjala como está). Anota la decisión en el commit.

## 5. Tests — van todos en `tests/`

Archivo nuevo: **`tests/sprint_error_test.go`**.

```go
func TestSprintDevuelveElMensajeDelError(t *testing.T) {
	err := errors.New("boom")
	if got := fmt.Sprint(err); got != "boom" {
		t.Errorf("Sprint(err) = %q, se esperaba %q", got, "boom")
	}
}

func TestSprintConErrorEnvueltoDevuelveLaCadenaCompleta(t *testing.T)
	// fmt.Errf("contexto:", errors.New("boom")) → el texto completo, no ""

func TestSprintNoRompeLosTiposQueYaFuncionaban(t *testing.T)
	// string, int, bool, float64, nil → mismo resultado que antes del cambio

func TestSprintConErrorNilTipadoNoEntraPorLaRamaDeError(t *testing.T)
	// var e error = nil; Sprint(e) NO debe entrar en la rama (por eso el `&& err != nil`)
```

Y el test **con forma de consumidor**, que es el que prueba de verdad la API
(regla de oro del `CONSTRUCTION_HARNESS`): reproducir el patrón de log real,

```go
func TestPatronDeLogDelEcosistemaNoPierdeElError(t *testing.T) {
	// mismo bucle que app.SprintMessages
	sprintMessages := func(messages ...any) string {
		res := ""
		for i, m := range messages {
			if i > 0 { res += " " }
			res += fmt.Sprint(m)
		}
		return res
	}
	got := sprintMessages("assetmin flush error:", errors.New("disco lleno"))
	want := "assetmin flush error: disco lleno"
	if got != want { t.Errorf("got %q, want %q", got, want) }
}
```

Ese test es la razón de ser del plan: si vuelve a fallar, el ecosistema se queda
ciego otra vez.

## 6. Restricciones del repo

- Sin librería estándar en código que compile a WASM: nada de `errors`,
  `strconv`, `strings` en el código de producción de este repo. **Los tests sí
  pueden** usar `errors` y `testing`.
- Todos los `*_test.go` van en `tests/`. Verificable:
  `find . -name "*_test.go" -not -path "./tests/*" -not -path "./.git/*"` → vacío.
- Sin cadenas repetidas: si necesitas un literal más de una vez, constante con
  nombre.

## 7. Criterios de aceptación

| # | Comprobación | Resultado esperado |
|---|---|---|
| 1 | `go test ./...` | verde |
| 2 | `fmt.Sprint(errors.New("boom"))` | `"boom"` |
| 3 | `fmt.Sprintf("%v", errors.New("boom"))` | `"boom"` (sin regresión) |
| 4 | `fmt.Sprint("hola")`, `fmt.Sprint(42)`, `fmt.Sprint(true)` | sin cambios |
| 5 | `git diff --stat` | toca `sprint.go` y `tests/`, **nada más** |
| 6 | `grep -n "hasContent(BuffErr)" convert.go` | sigue igual que antes |

## 8. Etapas

| # | Etapa | Archivos |
|---|---|---|
| 1 | Test que reproduce el fallo (rojo) | `tests/sprint_error_test.go` |
| 2 | Arreglo en `Sprint` | `sprint.go` |
| 3 | Test de no regresión de los otros tipos | `tests/sprint_error_test.go` |
| 4 | Barrido de `Convert(...).String()` y decisión anotada | — |
