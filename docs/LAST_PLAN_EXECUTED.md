---
PLAN: "fix: dejar de importar io, la unica dependencia de stdlib que queda"
EXECUTOR: jules
REVIEWER: none
---

> Este plan se despacha con el flujo CodeJob. Ver skill: agents-workflow.

# Plan — `webtyp/fmt`: cerrar la puerta a `io`

## El problema

`error.go` de este mismo paquete abre con esta declaración de intenciones:

> *"Custom error messages to avoid importing standard library packages like
> `errors` or `fmt`. This keeps the binary size minimal for embedded systems and
> WebAssembly"*

Y sin embargo `fmt_template.go` empieza con `import "io"`, para una sola función:

```go
func Fprintf(w io.Writer, format string, args ...any) (n int, err error)
```

`io` importa `errors`, y `errors` importa `internal/reflectlite`.

## Cuánto vale — la parte honesta

**Hoy, cero bytes.** Está medido: quitar este import no reduce el binario, porque
el runtime de TinyGo necesita `internal/reflectlite` para sus aserciones de tipo
y lo enlaza igual. Que nadie persiga esos 10,9 KB creyendo que salen de aquí.

Se hace por otra razón, que sí está medida y es la que importa:

> **El impuesto de la stdlib lo cobra la última puerta que quede abierta.**

En un Worker real, `unicode` y `bytes` entraban por dos caminos distintos. Cerrar
uno solo rendía 2.119 bytes; cerrar los dos, 93.733. Mientras `webtyp/fmt`
—que está en **todos** los binarios del ecosistema— mantenga abierta la vía
`io → errors`, cualquier otro paquete que empiece a usar `errors.Is` o un
`bytes.Buffer` reabre el grifo sin que la guarda del grafo lo note como
regresión: ya estaba abierto.

Este plan cierra la puerta. La ganancia es futura y es de garantía, no de bytes
inmediatos.

## El cambio

En `fmt_template.go`:

```go
// Writer es io.Writer redeclarado aquí. Es estructuralmente idéntico, así que
// cualquier io.Writer lo satisface sin que el llamador cambie una línea — pero
// importar "io" arrastra "errors", y con él internal/reflectlite, a todo binario
// del ecosistema.
type Writer interface {
	Write(p []byte) (n int, err error)
}

func Fprintf(w Writer, format string, args ...any) (n int, err error)
```

**Esto no rompe a nadie.** Go satisface interfaces estructuralmente: quien hoy
pasa un `*os.File`, un `bytes.Buffer` o un `http.ResponseWriter` a `Fprintf`
sigue compilando sin tocar su código.

Revisa además `memory.go`, que según el barrido también usa `io.`; aplícale el
mismo tratamiento. Y comprueba que ningún otro archivo del paquete importe
`errors`, `bytes`, `strings` o `strconv` de la biblioteca estándar.

## Criterios de aceptación

- [ ] `GOOS=js GOARCH=wasm go list -f '{{join .Imports " "}}' .` devuelve
      exactamente `syscall/js unsafe`.
- [ ] `grep -rn '"io"\|"errors"\|"bytes"\|"strings"\|"strconv"' *.go | grep -v _test`
      → vacío.
- [ ] `Fprintf` conserva su firma vista desde el llamador: un test pasa un
      `*bytes.Buffer` de la stdlib —desde un `_test.go`, donde la stdlib es
      legítima— y compila y funciona.
- [ ] La batería actual del repositorio pasa sin modificarse.

**Anti-footgun:** los `_test.go` de este repositorio compilan con Go estándar y
usan `bytes`, `fmt` y `testing` de la stdlib con toda legitimidad —de hecho los
necesitan para comparar contra el comportamiento que este paquete replica—. **No
toques sus imports.** La restricción es sobre el código del paquete.

## Fuera de alcance

Cualquier cambio en el comportamiento de formateo, en `Conv`, o en la API
pública. Este plan sustituye un tipo de parámetro y nada más.
