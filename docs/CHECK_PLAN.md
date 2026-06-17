# PLAN — Extraer i18n a `fmt/lang` (diccionario opt-in) · BREAKING CHANGE

> Este plan se despacha vía el workflow CodeJob. Ver skill: `agents-workflow`.
> **Estado:** LISTO PARA DESPACHO.
> **Repo objetivo:** `github.com/tinywasm/fmt`.
> **Tipo:** breaking change (mueve la API de i18n a un subpaquete `fmt/lang`).
> **Objetivo:** que el core `fmt` NO arrastre el diccionario/traducción por defecto.
> El que quiera traducir importa `fmt/lang` explícitamente (opt-in).

## Prerequisito (PRIMERO — entorno del agente)

El agente NO puede probar en el navegador. Toda verificación es con `gotest`, que no viene
preinstalado en el entorno aislado del agente. Instalarlo antes de cualquier otra cosa:

```bash
go install github.com/tinywasm/devflow/cmd/gotest@latest
```

`gotest` corre backend (stdlib) **y** WASM con build tags, más `-vet`/`-race`/`-cover`. Usar
`gotest` (sin argumentos para toda la suite, o `gotest -run TestX`); **NO** `go test` directo.

## Problema (medido)

El diccionario multilingüe está **cableado en el core** de `fmt`: cualquier camino de error o
de formato lo alcanza, aunque el binario no quiera traducir.

- `Err(...)` / `wrErr(...)` → `SmartArgs`/`processTranslatedArgs` → `lookupWord` (diccionario).
- `wrFormat` (motor de `Sprintf`/`Errf`) recibe un `lang` y el verbo `%L` llama a `lookupWord`.
- `Html(...)` → `SmartArgs` → `lookupWord`.
- Peor aún: un `init()` aguas abajo (`form/words.go`) llama `RegisterWords(...)` de forma
  **automática**, así que **importar `form` arrastra el diccionario + los datos** sin que
  nadie lo pida.

TinyGo ya elimina código muerto por binario (un `Convert().String()` pesa 325 B). Por eso el
problema NO es código duplicado: es **acoplamiento de alcanzabilidad**. La única forma de que
el diccionario no entre a un binario es que el core deje de llamarlo por defecto.

## Principio rector

> La traducción es **opt-in**. El core `fmt` emite el texto crudo por defecto (sin tocar
> diccionario). Quien quiera traducir importa `github.com/tinywasm/fmt/lang`, que instala un
> hook de traducción en el core y expone `RegisterWords`, `OutLang`, `Translate`. Si nadie
> importa `fmt/lang`, todo el motor de i18n (diccionario, sort, idiomas, datos) queda
> **inalcanzable** y no entra al binario.

Esto **no** cambia el comportamiento por defecto observable cuando no hay palabras
registradas: hoy, sin `RegisterWords`, `lookupWord` ya devuelve el texto crudo. El cambio es
que la traducción deja de ser automática y pasa a requerir importar `fmt/lang`.

## Decisión arquitectónica (resuelta)

1. **`Sprintf`/`Fprintf`/`Sscanf`/`Printf`/`Println` se quedan en root `fmt`** (son intuitivos
   y muy usados). NO se mueven. Solo dejan de depender del diccionario.
2. **Hook de traducción en el core** (nil por defecto):
   ```go
   // core fmt (nuevo archivo translate_hook.go)
   // translateWord traduce una palabra al idioma actual. nil = sin traducción (texto crudo).
   var translateWord func(word string) (string, bool)

   // SetTranslator instala el traductor global. Lo llama fmt/lang en su init().
   // Pasar nil restaura el modo sin traducción.
   func SetTranslator(fn func(word string) (string, bool)) { translateWord = fn }
   ```
   El core usa `translateWord` así (helper único, reutilizado por todos los caminos):
   ```go
   func tr(word string) string {
       if translateWord != nil {
           if t, ok := translateWord(word); ok {
               return t
           }
       }
       return word
   }
   ```
3. **El tipo `lang` y las constantes `EN..RU` se mueven a `fmt/lang`.** El core deja de
   conocer idiomas. La selección de idioma por argumento (`Err(ES, ...)`, `Translate(ES, ...)`)
   queda disponible **solo** vía `fmt/lang` (`lang.Translate(lang.ES, ...)`). El core `Err`
   no acepta selector de idioma: traduce palabra por palabra con el hook global.
4. **`fmt/lang` instala el hook en su `init()`** → importar `fmt/lang` ES el opt-in. El hook
   cierra sobre el idioma global actual de `fmt/lang`.

### Qué se MUEVE a `fmt/lang` (paquete nuevo `github.com/tinywasm/fmt/lang`)

| Origen (core) | Símbolos |
|---|---|
| `dictionary.go` | `DictEntry`, `entry`, `dictEntries`, `langCount`, `RegisterWords`, `lookupWord`, `sortDict`, `quicksort`, `partition`, `compareCaseInsensitive` |
| `language.go` | tipo `lang`, constantes `EN ES ZH HI AR PT FR DE RU`, `defLang`, `OutLang`, `langParser`, `mapLangCode` |
| `language.back.go` / `language.front.go` | `defLangMu`, `setDefaultLang`, `getCurrentLang` (mantener el split `!wasm`/`wasm`) |
| `translation.go` | `Translate`, `SmartArgs`, `detectLanguage`, `processTranslatedArgs`, `shouldAddSpace` |
| `env.back.go` / `env.front.go` | SOLO `getSystemLang` (detección de idioma del sistema/navegador). `Println`/`Printf`/`isWasm` se QUEDAN en core |

### Qué se QUEDA en el core `fmt` (y cómo se reescribe)

- **`error.go`**: `Err`, `Errf`, `wrErr`, `StringErr`, `Error`, `getError`.
  - `Err(msgs ...any)`: reescribir para NO usar `SmartArgs`. Escribe los args separados por
    espacio en `BuffErr`; para cada `string` aplica `tr(...)`; para números/bool/error usa la
    lógica que ya tiene `wrErr`. (Mover/compartir la lógica de `wrErr`; no duplicar.)
  - `Errf`: igual que hoy pero `wrFormat` ya no recibe `lang`.
  - `wrErr`: reemplazar `lookupWord(v, getCurrentLang())` por `tr(v)`.
- **`fmt_template.go`** (`wrFormat`): quitar el parámetro `currentLang lang`. El verbo `%L`
  pasa a usar `tr(strVal)` en vez de `lookupWord(strVal, currentLang)`. Ajustar los callers
  (`Sprintf`, `Fprintf`, `Errf`, y `SmartArgs` que se va a `fmt/lang`).
- **`html.go`** (`Html`): hoy usa `SmartArgs`. Reescribir para usar el motor de formato del
  core (`wrFormat`) + `tr(...)` para palabras, sin depender de `detectLanguage`/`lang`. La
  detección de "¿es format string?" puede quedarse en core (no es i18n).
- **`translate_hook.go`** (NUEVO): el hook `translateWord`, `SetTranslator`, helper `tr`.

### Cómo `fmt/lang` reescribe lo que era método de `*Conv`

`OutLang`, `langParser`, `mapLangCode` eran métodos de `*Conv` que usaban internos privados
(`splitStr`, `changeCase`, `ResetBuffer`, `WrString`, `GetString`). En `fmt/lang` NO pueden ser
métodos de `Conv` (otro paquete). Reescribirlos como funciones libres usando SOLO la API
**pública** de `fmt`:

- `c.splitStr(s, ".")` → `fmt.Split(s, ".")` (ya existe la función libre pública `Split`).
- `c.changeCase(true, ...)` / minúsculas → `fmt.Convert(s).ToLower().String()`.
- Donde se necesitaba un buffer temporal → usar `fmt.Convert(...)`.

`fmt/lang` instala el hook:
```go
package lang

import "github.com/tinywasm/fmt"

func init() {
    fmt.SetTranslator(func(word string) (string, bool) {
        return lookupWord(word, getCurrentLang())
    })
}
```

> ⚠️ **Sin ciclo de imports:** `fmt/lang` importa `fmt` (correcto). El core `fmt` **NUNCA**
> importa `fmt/lang`. Verificar que no haya import del core hacia el subpaquete.

## API pública resultante (resumen para consumidores)

| Antes (root `fmt`) | Después |
|---|---|
| `fmt.OutLang(...)` | `lang.OutLang(...)` |
| `fmt.Translate(...)` | `lang.Translate(...)` |
| `fmt.RegisterWords(...)` | `lang.RegisterWords(...)` |
| `fmt.DictEntry{...}` | `lang.DictEntry{...}` |
| `fmt.ES`, `fmt.EN`, ... | `lang.ES`, `lang.EN`, ... |
| `fmt.Err(...)`, `fmt.Sprintf(...)`, `fmt.Html(...)` | **igual** (en root; traducen solo si se importó `fmt/lang`) |

## Pasos de ejecución

### Stage 1 — hook en el core
1. Crear `translate_hook.go`: `translateWord`, `SetTranslator(fn)`, helper `tr(word) string`.

### Stage 2 — desacoplar el core del diccionario
2. `error.go`: `Err`/`wrErr` usan `tr(...)` en vez de `SmartArgs`/`lookupWord`. `Err` deja de
   aceptar selector de idioma (sin `detectLanguage`).
3. `fmt_template.go`: `wrFormat` sin parámetro `lang`; `%L` usa `tr(...)`. Ajustar callers.
4. `html.go`: `Html` sin `SmartArgs`/`lang`; usar `wrFormat` + `tr(...)`.

### Stage 3 — crear el subpaquete `fmt/lang`
5. Crear el dir `lang/` con `package lang`. Mover dictionary/language/translation y
   `getSystemLang` (ver tabla "Qué se MUEVE"). Reescribir `OutLang`/`langParser`/`mapLangCode`
   como funciones libres usando la API pública de `fmt`. Mantener el split `*.back.go`/
   `*.front.go` para `getCurrentLang`/`setDefaultLang`/`getSystemLang`.
6. `lang/translate.go` con el `init()` que llama `fmt.SetTranslator(...)`.

### Stage 4 — tests y verificación
7. Mover los tests de i18n al paquete `lang` (`translation_test.go`, `dictionary_test.go`,
   `messagetype_test.go` si toca idioma, `capitalize_translate_test.go` si aplica, los tests
   de `OutLang`). Los tests del core que comprobaban traducción automática deben:
   - o importar `fmt/lang` (y así habilitar el hook), o
   - actualizarse para esperar texto crudo (comportamiento por defecto sin `fmt/lang`).
8. `gotest` verde en `fmt` **y** en `fmt/lang`.

### Stage 5 — documentación (OBLIGATORIO — no omitir)
Esta es una API pública que cambia de paquete; la doc DEBE reflejarlo. Archivos a actualizar:

9. **`docs/TRANSLATE.md`** — es la doc de i18n. Reescribir todos los ejemplos:
   `import "github.com/tinywasm/fmt/lang"` y `lang.OutLang(...)`, `lang.Translate(...)`,
   `lang.RegisterWords(...)`, `lang.ES`/`lang.EN`. Explicar el modelo **opt-in**: sin importar
   `fmt/lang` no hay traducción (texto crudo); importarlo instala el traductor global.
10. **`README.md`** — la feature "Multilingual error messages" debe aclarar que la traducción
    es **opt-in** vía `fmt/lang`. Actualizar cualquier ejemplo con `OutLang`/`Translate`/
    `RegisterWords` al nuevo import. `Err`/`Sprintf`/`Html` siguen en root (sin cambio de import).
11. **`docs/API_ERRORS.md`** — `Err` ya no acepta selector de idioma (`Err(ES, ...)`); la
    selección de idioma por llamada es `lang.Translate(lang.ES, ...)`. Documentar que `Err`
    traduce solo si se importó `fmt/lang`.
12. **`docs/API_HTML.md`** — `Html` sigue en root; aclarar que traduce palabras solo con
    `fmt/lang` importado.
13. **`docs/API_FMT.md`** y **`docs/MESSAGE_TYPES.md`** — actualizar referencias a `OutLang`/
    `Translate`/idiomas que hayan quedado obsoletas (repuntar a `fmt/lang`).
14. **`lang/README.md`** (NUEVO) — doc del subpaquete: API (`OutLang`, `Translate`,
    `RegisterWords`, `DictEntry`, constantes de idioma), el `init()` que instala el hook, y los
    9 idiomas soportados. Enlazar desde el README raíz en el índice de documentación.

> Verificación de doc: `grep -rn 'fmt\.OutLang\|fmt\.Translate\|fmt\.RegisterWords\|fmt\.DictEntry'
> README.md docs/*.md` debe quedar vacío (ya no se referencian desde el namespace `fmt.`).

## Verificación (repo-local, ejecutable por el agente)

```bash
# 1. El core NO referencia el diccionario (criterio principal):
grep -rn 'lookupWord\|RegisterWords\|dictEntries\|sortDict' *.go | grep -v _test && echo "FALLA: core aún toca el diccionario" || echo "OK: core sin diccionario"

# 2. No hay ciclo: el core NO importa el subpaquete lang:
grep -rn 'tinywasm/fmt/lang' *.go | grep -v _test && echo "FALLA: ciclo de import" || echo "OK: sin ciclo"

# 3. fmt/lang compila y depende solo del core fmt:
GOOS=js GOARCH=wasm go list -deps ./lang | grep tinywasm

# 4. Un binario que usa Err SIN importar fmt/lang no linkea el diccionario.
#    (Comprobación de humo; la validación de tamaño fina es aguas abajo en el edge.)
GOOS=js GOARCH=wasm go list -deps . | grep -i 'sort' || echo "OK"

# 5. Tests verdes en ambos paquetes:
gotest
gotest ./lang

# 6. La doc ya no referencia la API i18n desde el namespace `fmt.`:
grep -rn 'fmt\.OutLang\|fmt\.Translate\|fmt\.RegisterWords\|fmt\.DictEntry' README.md docs/*.md && echo "FALLA: doc desactualizada" || echo "OK: doc migrada a fmt/lang"
```

## Checklist de calidad (obligatorio)

- **Sin strings hardcodeados repetidos:** códigos de idioma (`"en"`, `"es"`, ...), nombres de
  idioma y separadores que se repitan → constantes nombradas en `fmt/lang`. Nada de literales
  duplicados en la lógica.
- **Sin duplicación lógica:** la lógica de `wrErr` y la de `Err` deben compartir el mismo
  helper de escritura de args (no copiar el switch de tipos). La lógica i18n se MUEVE, no se
  copia: no dejar copias en core y en `lang`.
- **Reglas tinywasm:**
  - Nada de stdlib pesado en código wasm: usar `tinywasm/fmt` (no `errors`/`strconv`/`strings`).
    `fmt/lang` debe usar la API pública del core, no reimplementar split/case.
  - Embebido por valor (no punteros) para tipos `dom` (no aplica aquí, pero respetar).
  - El core `fmt` debe quedar libre de `lang`/diccionario/`reflect`/`regexp`.
  - Mantener el patrón `*.back.go` (`//go:build !wasm`) / `*.front.go` (`//go:build wasm`) para
    el mutex y la detección de idioma del sistema.

## Tabla de stages

| Stage | Objetivo | Entregable | Criterio de salida |
|---|---|---|---|
| 1 | Hook en core | `translate_hook.go` (`SetTranslator`, `tr`) | compila |
| 2 | Core sin diccionario | `error.go`/`fmt_template.go`/`html.go` reescritos | `grep` (verif. 1) limpio |
| 3 | Subpaquete `fmt/lang` | `lang/` con i18n + `init()` que instala el hook | `go list ./lang` OK, sin ciclo |
| 4 | Tests | tests de i18n movidos/ajustados | `gotest` y `gotest ./lang` verdes |
| 5 | Documentación | `TRANSLATE.md`, `README.md`, `API_ERRORS.md`, `API_HTML.md`, `API_FMT.md`, `MESSAGE_TYPES.md`, `lang/README.md` | `grep` de doc (verif. abajo) limpio |

## Nota (fuera de este repo — fase aguas abajo del master plan)

Estos consumidores deberán migrar `fmt.X` → `lang.X` (NO es parte de este plan; se coordina en
`docs/SIZE_OPTIMIZATION_MASTER_PLAN.md`):

- `tinywasm/form/words.go`: su `init()` que llama `RegisterWords` debe pasar a `lang.RegisterWords`
  y, idealmente, dejar de ser un `init()` automático (registro opt-in por la app) para que
  importar `form` no arrastre i18n.
- Llamadas a `OutLang` (~23), `RegisterWords` (~19), `Translate` (~8) y constantes de idioma en
  el ecosistema → repuntar a `github.com/tinywasm/fmt/lang`.

La comprobación de tamaño del edge (`edge.wasm`) se hace al final del master plan, no aquí.
