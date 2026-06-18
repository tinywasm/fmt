# PLAN — Contrato tipado de (de)serialización por campos: `FieldWriter`/`Encodable` (0-alloc, map-free, sin `any`) · BREAKING

> Este plan se despacha vía el workflow CodeJob. Ver skill: `agents-workflow`.
> **Estado:** LISTO PARA REVISIÓN DEL USUARIO.
> **Repo objetivo:** `github.com/tinywasm/fmt`.
> **Tipo:** breaking change (nuevo contrato canónico de serialización; deja legacy el camino
> basado en `any`/`map`/`reflect`).
> **Objetivo:** una **única** forma tipada de cruzar valores Go por un límite (JSON, JS, etc.)
> con **0 asignaciones en el heap Go**, **sin `map`** y **sin `any`** en el camino caliente —
> el mismo principio que usa `tinywasm/json`, llevado a su forma sin `any`.

## Prerequisito (PRIMERO — entorno del agente)

```bash
go install github.com/tinywasm/devflow/cmd/gotest@latest
```

Usar `gotest` (sin argumentos para la suite completa); **NO** `go test` directo.

## Por qué este contrato vive en `fmt` (no en `jsvalue`)

Los **modelos son agnósticos** (compilan backend **y** wasm). El encoder concreto de JS vive en
`tinywasm/jsvalue`, que es **`//go:build wasm`**. Si un modelo implementara una interfaz definida
en `jsvalue`, **rompería el build backend**. Por eso el **contrato** (las interfaces que el
modelo implementa) debe estar en un paquete **agnóstico** = `fmt`. Los encoders concretos
(`jsvalue`, `json`) lo consumen. Es el mismo patrón que `fmt.Fielder` (en `fmt`) consumido por
`tinywasm/json`.

## Problema que resuelve (raíz arquitectónica)

El ecosistema convierte en los límites con **tipo borrado** (`any`):

```go
ToJS(data any) js.Value
ToGo(jsVal js.Value, v any) error
Convert(v ...any) *Conv
Pointers() []any        // ← fmt.Fielder también boxea en any y asigna un slice
```

`any` obliga a redescubrir el tipo en runtime, y eso solo se resuelve con: (a) un `switch` de
tipos infinito, (b) `reflect` (arrastra ~72 KB de tablas en wasm), o (c) `map` (arrastra el
runtime de hashmap de TinyGo, que infla el binario). Las tres son malas. Además, al haber
varios puntos de conversión sin contrato único, **las reglas divergen** (ej. `[]byte` hoy se
codifica de 3 formas distintas en `jsvalue`).

> **Nota de tamaño vs. diseño:** la fuga de tamaño (reflect ~72 KB) se resuelve aparte en
> `jsvalue/docs/PLAN.md`. ESTE plan ataca la **causa raíz** (el `any`/`map` en el límite) con un
> contrato tipado reutilizable, persiguiendo **0-alloc Go-side**.

## Diseño (resuelto — para revisión del usuario)

**Patrón visitor tipado.** El valor **escribe sus propios campos** con llamadas tipadas a un
`FieldWriter`; y **lee sus campos** por nombre desde un `FieldReader`. `fmt` define solo las
interfaces (y un encoder JSON de referencia para validar). Cada límite implementa el writer/reader
concreto. **Sin `any`, sin `map`, sin `reflect`, sin slices en el camino caliente.**

### Interfaces nuevas en `fmt` (archivo nuevo `codec.go`)

```go
package fmt

// FieldWriter recibe los campos de un valor por llamadas TIPADAS.
// Implementaciones: jsvalue (escribe js.Value), json (escribe bytes), etc.
// Reglas: cero `any`, cero `map`, cero asignación en el heap Go (el writer reusa su buffer).
type FieldWriter interface {
	String(name, val string)
	Int(name string, val int64)
	Uint(name string, val uint64)
	Float(name string, val float64)
	Bool(name string, val bool)
	Bytes(name string, val []byte)
	Null(name string)
	// Anidado: objeto hijo que también es Encodable.
	Object(name string, val Encodable)
	// Arrays tipados sin []any: el writer abre el array y el valor empuja elementos.
	Array(name string, n int, each func(i int, a ArrayWriter))
}

// ArrayWriter empuja elementos tipados de un array (sin []any).
type ArrayWriter interface {
	String(val string)
	Int(val int64)
	Float(val float64)
	Bool(val bool)
	Bytes(val []byte)
	Object(val Encodable)
}

// Encodable: un valor que sabe escribir SUS campos (lo genera ormc).
type Encodable interface {
	EncodeFields(w FieldWriter)
}

// FieldReader entrega los campos por nombre, TIPADOS. El bool indica presencia.
// Lee por nombre directo (NO construye un map).
type FieldReader interface {
	String(name string) (string, bool)
	Int(name string) (int64, bool)
	Uint(name string) (uint64, bool)
	Float(name string) (float64, bool)
	Bool(name string) (bool, bool)
	Bytes(name string) ([]byte, bool)
	Object(name string, into Decodable) bool
	Array(name string) (ArrayReader, bool)
}

// ArrayReader recorre un array tipado.
type ArrayReader interface {
	Len() int
	String(i int) string
	Int(i int) int64
	Float(i int) float64
	Bool(i int) bool
	Bytes(i int) []byte
	Object(i int, into Decodable) bool
}

// Decodable: un valor que sabe leer SUS campos (lo genera ormc).
type Decodable interface {
	DecodeFields(r FieldReader) error
}
```

> Ejemplo de lo que generará `ormc` (NO parte de este plan; solo para entender el contrato):
> ```go
> func (u *User) EncodeFields(w fmt.FieldWriter) {
> 	w.String("name", u.Name)
> 	w.Int("age", int64(u.Age))
> }
> func (u *User) DecodeFields(r fmt.FieldReader) error {
> 	if v, ok := r.String("name"); ok { u.Name = v }
> 	if v, ok := r.Int("age"); ok { u.Age = int(v) }
> 	return nil
> }
> ```
> Cero `any`, cero `map`, cero asignación: el writer/reader son concretos y reusan buffers.

### `fmt` NO shippea encoder concreto (evita duplicación / "dos formas")

`fmt` define **solo las interfaces** (el contrato). El encoder/decoder **JSON canónico** vive en
`tinywasm/json`; el de **JS** en `tinywasm/jsvalue`. Si `fmt` trajera su propio encoder JSON
sería una **segunda forma** de serializar JSON = duplicación y deuda (justo lo que este rediseño
elimina). Por eso `fmt` no exporta ningún writer/reader concreto.

Para **validar** que el contrato funciona y es **0-alloc**, los tests de `fmt` definen un
`FieldWriter`/`FieldReader` **mock** (en el `_test.go`, no exportado, no shippeado): el writer
escribe pares `name=val` a un `[]byte` reusado; el reader devuelve valores por nombre desde
slices paralelos (sin `map`). Suficiente para medir asignaciones y cubrir el round-trip del
contrato sin introducir un segundo codec real.

## Pasos de ejecución

### Stage 1 — definir el contrato
1. Crear `codec.go` con las interfaces `FieldWriter`, `ArrayWriter`, `FieldReader`,
   `ArrayReader`, `Encodable`, `Decodable` (tal cual arriba). **Solo interfaces + doc.** Sin
   ninguna implementación concreta exportada.

### Stage 2 — tests del contrato (mock, incluye prueba de 0-alloc)
2. En `codec_test.go`: un `FieldWriter`/`FieldReader` **mock** no exportado (writer → `[]byte`
   reusado con `name=val;`; reader → slices paralelos `[]string` nombre/valor, **sin `map`**).
3. Tipo de ejemplo en el test que implemente `Encodable` y `Decodable` (sin `map`, sin `any`).
4. Round-trip con el mock (primitivos, anidado `Object`, `Array`, `[]byte`).
5. **Test de asignaciones**: `testing.AllocsPerRun(100, func(){ sample.EncodeFields(w) })` con el
   writer/buffer mock pre-creado y reusado → afirmar **0 asignaciones** del heap Go en el camino
   del visitor.
6. Test de decode con el reader mock → verifica que `DecodeFields` asigna los campos correctos.

### Stage 3 — documentación (OBLIGATORIO)
7. **`docs/API_CODEC.md`** (NUEVO): documentar el contrato (`Encodable`/`Decodable`/`FieldWriter`/
   `FieldReader`), las reglas (0-alloc, map-free, sin `any`), el ejemplo de `ormc`, y que los
   encoders concretos viven en `json` (JSON) y `jsvalue` (JS) — `fmt` solo define el contrato.
8. **`docs/CODEC_AND_FIELDER.md`** (YA EXISTE — documento de diseño de la separación de
   responsabilidades **codec vs `Field`/`Fielder`**): revisarlo y **actualizarlo a estado
   "implementado"** una vez que el contrato exista (quitar el "describe la arquitectura objetivo";
   confirmar que la tabla "quién usa qué" sigue exacta). Es la referencia canónica de por qué
   `Field`/`Fielder` se conservan (esquema/DB/validación/UI) y el codec solo cubre serialización.
   NO duplicar su contenido en `API_CODEC.md`: `API_CODEC.md` = API del contrato;
   `CODEC_AND_FIELDER.md` = separación de responsabilidades. Enlazarlos entre sí.
9. **`README.md`**: agregar al índice de documentación **ambos**: `API_CODEC.md` (forma canónica
   y tipada de serialización) y `CODEC_AND_FIELDER.md` (codec vs `Field`/`Fielder`).

## Verificación (repo-local, ejecutable por el agente)

```bash
# 1. El contrato no usa map ni any en su superficie:
grep -nE 'map\[|[^.]\bany\b' codec.go && echo "REVISAR: ¿map/any en el contrato?" || echo "OK: contrato sin map/any"

# 2. fmt NO shippea un encoder concreto (el mock vive solo en _test.go):
ls codec_json.go 2>/dev/null && echo "FALLA: fmt no debe shippear encoder concreto" || echo "OK"

# 3. Tests verdes + 0-alloc:
gotest -run Codec
gotest
```

## Checklist de calidad (obligatorio)

- **0 asignaciones Go-side** en el camino de codificación: validado con `testing.AllocsPerRun`.
  El writer concreto reusa su buffer; nunca retorna `[]any`/`[]Field`/`map` en el hot path.
- **Sin `map`**: prohibido en `codec.go` y `codec_json.go` (TinyGo arrastra el runtime de
  hashmap; infla el binario).
- **Sin `any` en el contrato**: las firmas son tipadas. (`any` solo permitido si fuera
  estrictamente inevitable; en este diseño NO lo es.)
- **Sin `reflect`**.
- **Sin strings hardcodeados repetidos**: separadores/llaves JSON (`{`, `}`, `:`, `,`) y nombres
  → constantes nombradas.
- **Sin duplicación**: reusar `JSONEscape`, `wrIntBase`, `wrFloat64`, etc. del propio `fmt`.

## Tabla de stages

| Stage | Objetivo | Entregable | Criterio de salida |
|---|---|---|---|
| 1 | Contrato tipado | `codec.go` (solo interfaces) | sin `map`/`any`/`reflect`; sin impl concreta |
| 2 | Tests + 0-alloc | mock writer/reader en `_test.go` + `AllocsPerRun==0` | `gotest` verde, 0 allocs |
| 3 | Documentación | `docs/API_CODEC.md` (nuevo) + `docs/CODEC_AND_FIELDER.md` (actualizar a "implementado") + índice en `README.md` | docs presentes, exactas y enlazadas |

## Alcance y coordinación

Este PLAN cubre **solo `fmt`** (las interfaces del contrato + tests con mock). `fmt` NO shippea
encoder concreto. Las librerías que implementan el contrato tienen su propio `docs/PLAN.md`
autocontenido, coordinados por `~/Dev/Project/tinywasm/docs/SIZE_OPTIMIZATION_MASTER_PLAN.md`:

- `orm/docs/PLAN.md` — `ormc` genera `EncodeFields`/`DecodeFields` en los modelos.
- `json/docs/PLAN.md` — `FieldWriter`/`FieldReader` JSON canónicos; migra `Encode`/`Decode` al
  codec para ser **0-alloc** (deja de usar `Pointers() []any` para serializar).
- `jsvalue/docs/PLAN.md` — `FieldWriter`/`FieldReader` sobre `js.Value`; elimina reflect/map/any.

**Este plan es GATE:** debe estar publicado antes de despachar `orm`, `json` y `jsvalue`.

**Una sola forma de serializar:** `json` y `jsvalue` usan AMBOS el mismo contrato (`Encodable`/
`Decodable`). El encoder JSON canónico vive en `json` (tiene el parser); `fmt` no duplica un
segundo JSON.

**Qué NO toca este cambio:** `fmt.Fielder`/`Schema()`/`Pointers()` permanecen, pero **solo** para
su rol de **DB**: el scan posicional de SQL `row.Scan(Pointers()...)` en `orm` (la API de
`database/sql` exige punteros posicionales; un visitor por-nombre no la reemplaza) + schema +
validación. La **serialización** (JSON/JS) deja de usar `Pointers()` y pasa al codec. Son
operaciones distintas, no "dos formas de serializar".
