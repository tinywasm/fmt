# PLAN: tinywasm/fmt — Contrato de seguridad por whitelist

## Contexto

Tras el fix de `hasPermittedRules` en `field.go` y `hasCharRules` en `permitted.go`
(separación de longitud y caracteres, ya publicado), el framework valida correctamente los
caracteres permitidos en cada campo. Esto tiene una implicación de seguridad importante que
debe quedar documentada y reforzada en el código.

Nota sobre nombres: `Permitted.hasCharRules()` (en `permitted.go`) y
`Field.hasPermittedRules()` (en `field.go`) son métodos en structs distintos — ambos ya
actualizados para excluir Minimum/Maximum. No se renombran: cada uno describe su contexto.

## Hallazgo

Los widgets estándar (`input.Text`, `input.Textarea`, `input.Email`) usan validación positiva
(whitelist): solo se permiten los caracteres explícitamente habilitados. Los caracteres
peligrosos para HTML (`<`, `>`, `&`, `"`, `'`) **no están en ningún whitelist** por defecto.

### Consecuencia directa

Datos validados por `ValidateFields` con widgets estándar no pueden contener chars HTML
peligrosos. El `EscapeHTML` aplicado en la capa de salida sobre esos datos es defensa en
profundidad, no el mecanismo primario de protección.

### Riesgo si no se documenta

Un developer podría:
1. Crear un widget custom con `Extra: []rune{'<', '>'}` sin darse cuenta del riesgo XSS.
2. Saltarse `ValidateFields` en el handler asumiendo que "el form del cliente ya validó".
3. Usar `FieldRaw` para datos externos (API, DB) y escribirlos en HTML sin escape —
   estos nunca pasan por `Permitted`.

---

## Decisiones de diseño

### 1. Validate no muta — nunca

`ValidateFields` y `Field.Validate` solo verifican; no transforman el dato. Si el valor que
entró y el que el schema ve son distintos, `Permitted` deja de ser fuente de verdad.

### 2. Seguridad por whitelist positivo (no por sanitización)

No se agrega `SanitizeAndValidate` ni `Field.Sanitize`. El motivo:
- La sanitización en entrada crea falsa sensación de que el framework protege la capa de salida.
- Encodear en salida es responsabilidad de quien escribe al contexto (HTML, SQL, JSON).
- La whitelist ya rechaza los chars peligrosos antes de que lleguen al handler.

### 3. Configuración insegura debe ser explícita

Si un campo necesita aceptar `<` o `>` (ej: editor de texto rico), el developer debe
agregarlo explícitamente a `Extra`. `NoHTML()` es la herramienta del framework para declarar
en código que se es consciente del riesgo y se mitiga — sin ella, el principio solo existe
en documentación.

---

## Cambios a implementar

### 1. Documentación de contrato en `permitted.go`

Agregar al bloque de doc de `Permitted`:

```go
// Security contract: the whitelist is positive — only explicitly enabled characters pass.
// HTML-dangerous characters (<, >, &, ", ') are not included in any standard widget's
// whitelist. Data validated through standard widgets is therefore safe for HTML output
// without additional escaping.
//
// If a custom widget adds dangerous chars to Extra, it must document the XSS risk
// and the caller is responsible for output encoding (e.g., fmt.Convert(v).EscapeHTML()).
```

### 2. Documentación en `field.go`

Agregar al bloque de doc de `Field.Validate`:

```go
// Security note: fields using standard widgets (Text, Textarea, Email) have whitelists
// that exclude HTML-dangerous characters. ValidateFields provides implicit XSS protection
// for form-submitted data. Data from external sources (DB reads, API responses) bypasses
// this check and must be encoded at the output layer.
```

### 3. Helper `Permitted.NoHTML()` — retorna copia modificada

Receiver por valor (consistente con el resto de `Permitted`). Retorna la copia modificada
para poder usarse en cadena durante la construcción del widget:

```go
// NoHTML adds HTML-dangerous characters to NotAllowed as an explicit safety layer.
// Returns a modified copy. Use when Extra contains characters that could appear in
// HTML injection attempts and the widget cannot be restricted further.
//
// Example:
//
//   t.Permitted = t.Permitted.NoHTML()
func (p Permitted) NoHTML() Permitted {
    p.NotAllowed = append(append([]string{}, p.NotAllowed...), "<", ">", "&")
    return p
}
```

Uso en widget custom:

```go
func RichTextarea() Input {
    t := &richTextarea{}
    t.Letters = true
    t.Extra = []rune{':', ';', '$', '#', '!', '?', '-'}
    t.Permitted = t.Permitted.NoHTML() // explícito: bloquear HTML aunque Extra sea amplio
    t.InitBase("", "", "textarea")
    return t
}
```

---

## Archivos a modificar

| Archivo | Cambio |
|---------|--------|
| `permitted.go` | Doc del contrato de seguridad + agregar `NoHTML()` con receiver por valor |
| `field.go` | Doc en `Field.Validate` sobre fuentes externas vs form data |

## Tests a agregar

| Test | Archivo | Verifica |
|------|---------|----------|
| `TestPermitted_NoHTML_BlocksInjection` | `permitted_security_test.go` | `Permitted{Letters:true}.NoHTML().Validate("f", "hel<lo")` → error |
| `TestPermitted_NoHTML_AllowsNormal` | `permitted_security_test.go` | `Permitted{Letters:true}.NoHTML().Validate("f", "hello")` → ok |
| `TestWidget_Text_RejectsHTML` | `permitted_security_test.go` | `input.Text().Validate("<script>")` → error |
| `TestWidget_Textarea_RejectsHTML` | `permitted_security_test.go` | `input.Textarea().Validate("<b>bold</b>")` → error |

## Orden de ejecución

1. Agregar `NoHTML()` en `permitted.go`
2. Agregar docs en `permitted.go` y `field.go`
3. Agregar tests de seguridad
4. Publicar via `gopush`
