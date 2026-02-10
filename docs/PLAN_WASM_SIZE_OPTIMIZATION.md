# WASM Binary Size Optimization: Eliminating `reflect` and `sync`

## Problem

The `web/client.go` example (a simple counter) compiles to **88.9 KB** in TinyGo mode S. This is excessive for a ~70-line application. The root cause is that `tinywasm/fmt` imports `reflect` (~30-40 KB) and `sync` (~5-10 KB), which are the two heaviest standard library packages in TinyGo WASM.

## Root Cause Analysis

### `reflect` — 3 files (~30-40 KB impact)

| File | Function | Purpose | Used in WASM? |
|------|----------|---------|---------------|
| `convert.go` | `AnyToBuff` default branch | `reflect.ValueOf()` for custom types (`type customInt int`) | Rarely — all common types handled by type switches |
| `num_int.go` | `toInt64Reflect()` | Reflection fallback for custom int types | Rarely |
| `num_float.go` | `toFloat64Reflect()` | Reflection fallback for custom float types | Rarely |

**Key insight**: The reflection code is a **fallback** for rare custom types. All standard Go types (`int`, `string`, `float64`, etc.) are handled by explicit `case` branches in `AnyToBuff`. TinyGo cannot eliminate the `default` branch via dead-code elimination because it's reachable via `any` interface.

### `sync` — 2 files (~5-10 KB impact)

| File | Usage | Purpose | Needed in WASM? |
|------|-------|---------|-----------------|
| `memory.go` | `sync.Pool` | Reuse `Conv` objects to reduce allocations | **No** — WASM is single-threaded |
| `language.go` | `sync.RWMutex` | Thread-safe access to `defLang` global | **No** — WASM is single-threaded |

## Design Decision: `!wasm` Build Tags

### Why `!wasm` instead of separate WASM files?

> [!IMPORTANT]
> Following the user's direction: concurrency (sync) belongs **only to backend** (`!wasm`). This way developers use the same `tinywasm/fmt` package everywhere without worrying about which import to use. The library handles it transparently.

### Alternatives Considered

1. **Separate `fmt_lite` package** — Rejected: forces developers to choose between `fmt` and `fmt_lite`, adds cognitive load
2. **Interface-based injection** — Rejected: adds complexity, needs initialization, not worth it for a fallback path
3. **Build tags on the functions** ✅ — Selected: transparent, follows existing `.back.go`/`.front.go` convention, zero developer friction

## Proposed Changes

### Strategy: Split files using existing naming convention

Follow the existing pattern: `env.back.go` (`!wasm`) / `env.front.go` (`wasm`).

---

### Component 1: `sync` Elimination

#### [MODIFY] [memory.go](memory.go)

Remove `sync` import and `convPool` variable. Keep all buffer methods (they don't use sync). Extract pool logic to build-tag files:

#### [NEW] [memory.back.go](memory.back.go)

```go
//go:build !wasm

package fmt

import "sync"

var convPool = sync.Pool{
    New: func() any {
        return &Conv{
            out:  make([]byte, 0, 64),
            work: make([]byte, 0, 64),
            err:  make([]byte, 0, 64),
        }
    },
}

func GetConv() *Conv {
    c := convPool.Get().(*Conv)
    c.resetAllBuffers()
    c.out = c.out[:0]
    c.work = c.work[:0]
    c.err = c.err[:0]
    c.dataPtr = nil
    c.kind = K.String
    return c
}

func (c *Conv) PutConv() {
    c.resetAllBuffers()
    c.out = c.out[:0]
    c.work = c.work[:0]
    c.err = c.err[:0]
    c.dataPtr = nil
    c.kind = K.String
    convPool.Put(c)
}

func (c *Conv) putConv() {
    c.PutConv()
}
```

#### [NEW] [memory.front.go](memory.front.go)

```go
//go:build wasm

package fmt

// WASM is single-threaded: use simple allocation instead of sync.Pool
func GetConv() *Conv {
    return &Conv{
        out:  make([]byte, 0, 64),
        work: make([]byte, 0, 64),
        err:  make([]byte, 0, 64),
    }
}

func (c *Conv) PutConv() {
    // No-op in WASM — GC handles cleanup
}

func (c *Conv) putConv() {
    // No-op in WASM
}
```

> [!NOTE]
> In WASM single-threaded, `sync.Pool` adds overhead without benefit. Direct allocation is faster because there's no GC pressure from concurrent goroutines.

#### [MODIFY] [language.go](language.go)

Remove `sync` import and mutex. Extract thread-safe access to build-tag files:

#### [NEW] [language.back.go](language.back.go)

```go
//go:build !wasm

package fmt

import "sync"

var defLangMu sync.RWMutex

func setDefaultLang(l lang) {
    defLangMu.Lock()
    defLang = l
    defLangMu.Unlock()
}

func getCurrentLang() lang {
    defLangMu.RLock()
    defer defLangMu.RUnlock()
    return defLang
}
```

#### [NEW] [language.front.go](language.front.go)

```go
//go:build wasm

package fmt

// WASM is single-threaded: no mutex needed
func setDefaultLang(l lang) {
    defLang = l
}

func getCurrentLang() lang {
    return defLang
}
```

---

### Component 2: `reflect` Elimination

#### [MODIFY] [convert.go](convert.go)

Remove `reflect` import. Move the `default` branch of `AnyToBuff` to build-tag files:

#### [NEW] [convert.back.go](convert.back.go)

```go
//go:build !wasm

package fmt

import "reflect"

// anyToBuffFallback handles custom types via reflection (backend only)
func (c *Conv) anyToBuffFallback(dest BuffDest, value any) {
    // Check Stringer interface first
    if stringer, ok := value.(interface{ String() string }); ok {
        c.kind = K.String
        c.WrString(dest, stringer.String())
        return
    }

    // Reflection fallback for custom types
    rv := reflect.ValueOf(value)
    switch rv.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        c.AnyToBuff(dest, rv.Int())
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        c.AnyToBuff(dest, rv.Uint())
    case reflect.Float32, reflect.Float64:
        c.AnyToBuff(dest, rv.Float())
    case reflect.String:
        c.AnyToBuff(dest, rv.String())
    default:
        c.wrErr(D.Type, D.Not, D.Supported)
    }
}
```

#### [NEW] [convert.front.go](convert.front.go)

```go
//go:build wasm

package fmt

// anyToBuffFallback handles unknown types in WASM (no reflect)
func (c *Conv) anyToBuffFallback(dest BuffDest, value any) {
    // Check Stringer interface (still works without reflect)
    if stringer, ok := value.(interface{ String() string }); ok {
        c.kind = K.String
        c.WrString(dest, stringer.String())
        return
    }
    c.wrErr(D.Type, D.Not, D.Supported)
}
```

Then in `convert.go`, replace the `default:` branch of `AnyToBuff` with `c.anyToBuffFallback(dest, value)`.

#### [MODIFY] [num_int.go](num_int.go)

Remove `reflect` import. Move `toInt64Reflect` to build-tag files:

#### [NEW] [num_int.back.go](num_int.back.go)

```go
//go:build !wasm

package fmt

import "reflect"

func (c *Conv) toInt64Reflect(arg any) (int64, bool) {
    rv := reflect.ValueOf(arg)
    switch rv.Kind() {
    case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
        return rv.Int(), true
    case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
        return int64(rv.Uint()), true
    default:
        return 0, false
    }
}
```

#### [NEW] [num_int.front.go](num_int.front.go)

```go
//go:build wasm

package fmt

func (c *Conv) toInt64Reflect(_ any) (int64, bool) {
    return 0, false
}
```

#### [MODIFY] [num_float.go](num_float.go)

Same pattern — extract `toFloat64Reflect`:

#### [NEW] [num_float.back.go](num_float.back.go)

```go
//go:build !wasm

package fmt

import "reflect"

func (c *Conv) toFloat64Reflect(arg any) (float64, bool) {
    rv := reflect.ValueOf(arg)
    switch rv.Kind() {
    case reflect.Float32, reflect.Float64:
        return rv.Float(), true
    default:
        return 0, false
    }
}
```

#### [NEW] [num_float.front.go](num_float.front.go)

```go
//go:build wasm

package fmt

func (c *Conv) toFloat64Reflect(_ any) (float64, bool) {
    return 0, false
}
```

---

## Summary of File Changes

| Action | File | What changes |
|--------|------|-------------|
| MODIFY | `memory.go` | Remove `sync` import, `convPool`, `GetConv`, `PutConv`, `putConv` |
| NEW | `memory.back.go` | `sync.Pool` + `GetConv`/`PutConv` with mutex |
| NEW | `memory.front.go` | Direct allocation `GetConv`/`PutConv` (no-op put) |
| MODIFY | `language.go` | Remove `sync` import, `defLangMu`, `getCurrentLang`. Refactor `OutLang` to use `setDefaultLang`/`getCurrentLang` |
| NEW | `language.back.go` | `sync.RWMutex` + `setDefaultLang`/`getCurrentLang` |
| NEW | `language.front.go` | Direct access `setDefaultLang`/`getCurrentLang` |
| MODIFY | `convert.go` | Remove `reflect` import, replace `default:` branch with `anyToBuffFallback()` |
| NEW | `convert.back.go` | `anyToBuffFallback` with `reflect` |
| NEW | `convert.front.go` | `anyToBuffFallback` without `reflect` (Stringer only) |
| MODIFY | `num_int.go` | Remove `reflect` import, remove `toInt64Reflect` |
| NEW | `num_int.back.go` | `toInt64Reflect` with `reflect` |
| NEW | `num_int.front.go` | `toInt64Reflect` stub (returns 0, false) |
| MODIFY | `num_float.go` | Remove `reflect` import, remove `toFloat64Reflect` |
| NEW | `num_float.back.go` | `toFloat64Reflect` with `reflect` |
| NEW | `num_float.front.go` | `toFloat64Reflect` stub (returns 0, false) |

**Total: 5 modified files + 10 new files**

## Expected Binary Size Reduction

| Before | After (estimated) |
|--------|-------------------|
| 88.9 KB | ~25-35 KB |

Reduction: **~55-65 KB (~60-70%)** by eliminating `reflect` and `sync` from WASM builds.

## Verification Plan

### Automated Tests

go install github.com/tinywasm/devflow/cmd/gotest@latest

1. `gotest` in `tinywasm/fmt` — all existing tests (backend StdLib + WASM) must pass

2. Verify concurrency tests still pass (they run only in `!wasm` via `gotest`)