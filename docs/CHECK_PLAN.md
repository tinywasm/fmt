# PLAN: tinywasm/fmt — Fix Float64 Scientific Notation

## Development Rules

- **Standard Library Only:** No external assertion libraries. Use `testing`.
- **Testing Runner:** Use `gotest` (install: `go install github.com/tinywasm/devflow/cmd/gotest@latest`).
- **Max 500 lines per file.** Subdivide by domain if exceeded.
- **Flat hierarchy.** No subdirectories for library code.
- **TinyGo Compatible:** No stdlib `fmt`, `strings`, `strconv`, `errors`. This IS tinywasm/fmt.
- **No maps** in WASM code (binary bloat).
- **Documentation First:** Update docs before coding.
- **Publishing:** Use `gopush 'message'` after tests pass.

## Context

`fmt.Convert(s).Float64()` falla al parsear strings en notación científica (e.g. `"1e2"`, `"1.5e-3"`, `"2.5E+4"`).

**Root cause:** `parseFloatBase()` en `num_float.go` (línea ~78) itera el integer part con:
```go
for ; i < len(s) && s[i] != '.'; i++ {
    if s[i] < '0' || s[i] > '9' {
        c.wrErr("character", "invalid")  // 'e' y 'E' llegan aquí
        return 0
    }
    ...
}
```

Al encontrar `'e'` o `'E'`, retorna `"character invalid"` en vez de parsear el exponente.

**Impacto:** `tinywasm/json` no puede decodificar campos float con notación científica (`{"f":1e2}`).

---

## Stage 1: Fix `parseFloatBase` en `num_float.go`

Agregar soporte de notación científica al final del parsing, después de parsear la parte entera y decimal:

```go
func (c *Conv) parseFloatBase() float64 {
    c.ResetBuffer(BuffErr)

    s := c.GetString(BuffOut)
    if len(s) == 0 {
        c.wrErr("string", "empty")
        return 0
    }

    var result float64
    var negative bool
    var hasDecimal bool
    var decimalPlaces int
    i := 0

    // Handle sign
    switch s[0] {
    case '-':
        negative = true
        i = 1
        if len(s) == 1 {
            c.wrErr("format", "invalid")
            return 0
        }
    case '+':
        i = 1
        if len(s) == 1 {
            c.wrErr("format", "invalid")
            return 0
        }
    }

    // Parse integer part (stop at '.' OR 'e'/'E')
    for ; i < len(s) && s[i] != '.' && s[i] != 'e' && s[i] != 'E'; i++ {
        if s[i] < '0' || s[i] > '9' {
            c.wrErr("character", "invalid")
            return 0
        }
        result = result*10 + float64(s[i]-'0')
    }

    // Parse decimal part if present
    if i < len(s) && s[i] == '.' {
        hasDecimal = true
        i++ // Skip decimal point
        for ; i < len(s) && s[i] != 'e' && s[i] != 'E'; i++ {
            if s[i] < '0' || s[i] > '9' {
                c.wrErr("character", "invalid")
                return 0
            }
            decimalPlaces++
            result = result*10 + float64(s[i]-'0')
        }
    }

    // Apply decimal places
    if hasDecimal {
        for j := 0; j < decimalPlaces; j++ {
            result /= 10
        }
    }

    // Parse scientific notation exponent if present
    if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
        i++ // skip 'e'/'E'
        if i >= len(s) {
            c.wrErr("format", "invalid")
            return 0
        }
        expNeg := false
        if s[i] == '+' {
            i++
        } else if s[i] == '-' {
            expNeg = true
            i++
        }
        if i >= len(s) {
            c.wrErr("format", "invalid")
            return 0
        }
        var exp int
        for ; i < len(s); i++ {
            if s[i] < '0' || s[i] > '9' {
                c.wrErr("character", "invalid")
                return 0
            }
            exp = exp*10 + int(s[i]-'0')
        }
        // Apply exponent
        mult := 1.0
        for j := 0; j < exp; j++ {
            mult *= 10
        }
        if expNeg {
            result /= mult
        } else {
            result *= mult
        }
    }

    if negative {
        result = -result
    }

    return result
}
```

**Cambios respecto al original:**
1. Integer part loop: agregar `&& s[i] != 'e' && s[i] != 'E'` a la condición.
2. Decimal part loop: agregar `&& s[i] != 'e' && s[i] != 'E'` a la condición.
3. Nuevo bloque al final para parsear el exponente después del mantissa.

---

## Stage 2: Agregar tests en `convert_test.go` o `numeric_test.go`

Verificar si ya existe `TestFloat64ScientificNotation`. Si no, agregar:

```go
func TestFloat64ScientificNotation(t *testing.T) {
    cases := []struct {
        input    string
        expected float64
    }{
        {"1e2", 100.0},
        {"1E2", 100.0},
        {"1.5e2", 150.0},
        {"1.5E+2", 150.0},
        {"1e-2", 0.01},
        {"2.5e-3", 0.0025},
        {"-1e2", -100.0},
        {"1e0", 1.0},
    }
    for _, c := range cases {
        t.Run(c.input, func(t *testing.T) {
            got, err := Convert(c.input).Float64()
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            // Allow small floating point tolerance
            diff := got - c.expected
            if diff < 0 {
                diff = -diff
            }
            if diff > 1e-9 {
                t.Errorf("expected %f, got %f", c.expected, got)
            }
        })
    }
}

func TestFloat64ScientificNotationErrors(t *testing.T) {
    cases := []string{"1e", "1e+", "1e-", "1eX"}
    for _, input := range cases {
        t.Run(input, func(t *testing.T) {
            _, err := Convert(input).Float64()
            if err == nil {
                t.Fatalf("expected error for input %q", input)
            }
        })
    }
}
```

### Verificar

```bash
gotest
```

---

## Stage 3: Publicar

```bash
gopush 'fmt: support scientific notation in Float64/Float32 parsing'
```

Luego actualizar `go.mod` en `tinywasm/json` con la nueva versión de `tinywasm/fmt`.
