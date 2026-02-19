# Translation API Simplification — `tinywasm/fmt`

> **This document is an LLM-executable prompt.**
> Follow every decision below **exactly**. No extra scope, no speculative refactors.

---

## Goal

Replace the static `D` struct + `LocStr` array system with a **string-key lookup engine**.

**Before:**
```go
Translate(ES, D.Format, D.Invalid).Capitalize().String()
Err(FR, D.Empty, D.String)
```

**After:**
```go
Translate(ES, "format", "invalid").Capitalize().String()
Err(FR, "empty", "string")
Err("fr", "empty", "string") // string language code still supported
```

---

## Core Design (All Locked ✅)

| # | Topic | Decision |
|---|-------|----------|
| Q1 | `D` struct + `LocStr` | **Hard break** — both removed entirely. No deprecated code left. |
| Q2 | Dictionary data structure | Sorted `[]entry` (struct: `key string` + `translations [9]string`). Binary search (O log n). No maps. |
| Q3 | Default dictionary | **None.** No translations are loaded by default. The developer opts in by importing `dictionary/` or registering their own words. |
| Q4 | Optional built-in dictionary | Lives in sub-package **`dictionary/dictionary.go`** — contains all 100 words × 9 languages (EN, ES, ZH, HI, AR, PT, FR, DE, RU). Loaded via blank import. |
| Q5 | Custom words API | `RegisterWords([]DictEntry)` — bulk registration. `DictEntry` has named fields for all 9 languages. |
| Q6 | Lookup case normalisation | ASCII in-place (reuses `changeCaseASCIIInPlace` logic). Zero allocation. |
| Q7 | Unknown key fallback | Pass-through as-is. |
| Q8 | `LocStr` | **Removed** — no dead code. |
| Q9 | Integration point | `processTranslatedArgs` switch — `string` case calls `lookupWord`. |
| Q10 | `langCount` | Fixed at `9`. Empty slots fall back to EN. |
| C1 | `lang` constants | All 9 constants stay in `language.go` (EN…RU). No import cycle. |
| C2 | `DictEntry` fields | All 9 language fields (`EN`, `ES`, `FR`, `DE`, `ZH`, `HI`, `AR`, `PT`, `RU`). |

---

## Architecture

### `fmt` root package — engine only, zero translations

The root package provides the **API and lookup engine**. No translations are registered by default.

#### `dictionary.go` (modify — keep filename)

The translation data (`var D = struct{...}` and all `LocStr` literals) is **removed** from this file — it moves entirely to `dictionary/dictionary.go`.

This file is repurposed to contain **only the engine**:

```go
// DictEntry describes one translatable word.
type DictEntry struct {
    Key string // lowercase English key, e.g. "empty"
    EN  string
    ES  string
    FR  string
    DE  string
    ZH  string
    HI  string
    AR  string
    PT  string
    RU  string
}

const langCount = 9

// internal entry
type entry struct {
    key          string
    translations [langCount]string
}

var dictEntries []entry // sorted by key

// RegisterWords adds word entries to the lookup engine. Safe to call from init().
// Called by dictionary/ sub-package and by developers registering custom words.
func RegisterWords(entries []DictEntry) { ... }

// lookupWord returns the translation for a word in the target language.
// Normalises word to lowercase (ASCII fast path, zero allocation).
// Returns false if no match; caller must pass-through the original string.
func lookupWord(word string, l lang) (string, bool) { ... }

func sortDict() { /* sort dictEntries by key */ }
```

> **WASM constraint:** No `map` anywhere — only slices and structs.

#### `language.go` (modify)

- Remove `type LocStr [9]string`. All 9 lang constants (EN…RU) stay unchanged.

#### `translation.go` (modify `processTranslatedArgs`)

Replace the `case LocStr:` arm with a string-lookup branch:

```go
case string:
    if translated, ok := lookupWord(v, currentLang); ok {
        c.WrString(dest, translated)
    } else {
        c.WrString(dest, v) // pass-through
    }
```

Remove the `case LocStr:` arm entirely.

#### `error.go` (modify `wrErr`)

Same change: remove `case LocStr:`, add string-lookup branch.

#### `parse.go` (no change)

`KeyValue` struct stays as-is. Not used by the engine directly.

---

### `dictionary/` sub-package — optional built-in translations

**One file: `dictionary/dictionary.go`**

Contains all current translations from `fmt/dictionary.go` **moved here** and reformatted as `[]DictEntry` with all 9 language fields. The `fmt/dictionary.go` file loses the translation data but gains the engine. This sub-package has the data.

Its `init()` calls `fmt.RegisterWords(builtinDict)` to register all ~100 words at once.

```go
package dictionary

import fmt "github.com/tinywasm/fmt"

func init() {
    fmt.RegisterWords([]fmt.DictEntry{
        {Key: "all",    EN: "All",    ES: "Todo",  FR: "Tout",  DE: "Alle",  ZH: "所有", HI: "सभी", AR: "كل",  PT: "Todo",  RU: "Все"},
        {Key: "empty",  EN: "Empty",  ES: "Vacío", FR: "Vide",  DE: "Leer",  ZH: "空",   HI: "खाली", AR: "فارغ", PT: "Vazio", RU: "Пустой"},
        // ... all ~100 words from current dictionary.go
    })
}
```

**Usage:**
```go
// Project that needs translations:
import _ "github.com/tinywasm/fmt/dictionary"

// Project using only custom words — does NOT import dictionary/
fmt.RegisterWords([]fmt.DictEntry{
    {Key: "invoice", EN: "invoice", ES: "factura"},
})
```

---

## Files Changed

> **Constraint:** No new files in the `fmt` package root.

| File | Action | Notes |
|------|--------|-------|
| `dictionary.go` | **MODIFY** | Translation data removed (moved to `dictionary/`). Engine added: `DictEntry`, `entry`, `langCount`, `dictEntries`, `lookupWord`, `RegisterWords`, `sortDict`. No `init()`, no default data. |
| `language.go` | **MODIFY** | Remove `type LocStr [9]string` only. All 9 lang constants stay. |
| `translation.go` | **MODIFY** | `processTranslatedArgs`: remove `LocStr` case, add string lookup via `lookupWord`. |
| `error.go` | **MODIFY** | `wrErr`: remove `LocStr` case, add string lookup via `lookupWord`. |
| `translation_test.go` | **MODIFY** | Add string-key tests (see Verification). |
| `dictionary_test.go` | **MODIFY** | Adapt existing tests to new engine API (`RegisterWords`, `lookupWord`). |
| `docs/TRANSLATE.md` | **REWRITE** | Document new API. |
| `dictionary/dictionary.go` | **NEW** (sub-package) | All ~100 words × 9 languages moved from old `dictionary.go`. `init()` calls `RegisterWords`. |
| `dictionary/dictionary_test.go` | **NEW** (sub-package) | Verifies all words load for all 9 languages. |

---

## Verification Plan

### 1. Regression
```bash
cd ~/Dev/Project/tinywasm/fmt
gotest ./...
```
All existing tests must pass. Coverage ≥ current baseline.

### 2. String key lookup with custom words
```go
func TestStringKeyCustom(t *testing.T) {
    RegisterWords([]DictEntry{
        {Key: "empty", EN: "Empty", ES: "Vacío"},
    })
    OutLang(ES)
    got := Translate("empty").String()
    if got != "Vacío" {
        t.Fatalf("got %q", got)
    }
}
```

### 3. Pass-through for unknown key
```go
func TestStringKeyPassthrough(t *testing.T) {
    got := Translate("unknownTerm").String()
    if got != "unknownTerm" {
        t.Fatalf("got %q", got)
    }
}
```

### 4. Case-insensitive lookup
```go
func TestStringKeyNormalization(t *testing.T) {
    RegisterWords([]DictEntry{{Key: "empty", EN: "Empty", ES: "Vacío"}})
    OutLang(ES)
    for _, input := range []string{"empty", "Empty", "EMPTY"} {
        got := Translate(input).String()
        if got != "Vacío" {
            t.Fatalf("input %q: got %q", input, got)
        }
    }
}
```

### 5. Optional dictionary sub-package
```go
// dictionary/dictionary_test.go
package dictionary_test

import (
    _ "github.com/tinywasm/fmt/dictionary"
    fmt "github.com/tinywasm/fmt"
)

func TestBuiltinDictLoaded(t *testing.T) {
    cases := []struct{ l fmt.lang; key, want string }{
        {fmt.ES, "empty",  "Vacío"},
        {fmt.FR, "empty",  "Vide"},
        {fmt.DE, "empty",  "Leer"},
        {fmt.ZH, "empty",  "空"},
    }
    for _, tc := range cases {
        got := fmt.Translate(tc.l, tc.key).String()
        if got != tc.want {
            t.Fatalf("lang %v key %q: want %q got %q", tc.l, tc.key, tc.want, got)
        }
    }
}
```

### 6. Benchmark
```bash
go test -bench=BenchmarkTranslate -benchmem ./...
```
Accept ±10% vs pre-change baseline.

---

## Plan B — Migration of Dependent Packages

After `tinywasm/fmt` is published at **v0.18.0**, scan `~/Dev/Project/tinywasm/` for old API usage.

**Search patterns:**
```
D\.[A-Z]\w+           # D.Format, D.Invalid, ...
LocStr\{              # LocStr literal definitions
```

**Migration rules:**
| Old | New |
|-----|-----|
| `D.SomeWord` | `"someword"` (lowercase English key) |
| `Err(FR, D.Empty, D.String)` | `Err(FR, "empty", "string")` |
| `Translate(ES, D.Format, D.Invalid)` | `Translate(ES, "format", "invalid")` |
| `LocStr{...}` custom dict | `fmt.RegisterWords([]fmt.DictEntry{{Key:..., EN:..., ES:...}})` |

Steps for the migration LLM:
1. `grep -r "D\." ~/Dev/Project/tinywasm --include="*.go" -l`
2. Replace patterns per file manually.
3. Add `import _ "github.com/tinywasm/fmt/dictionary"` where built-in words are needed.
4. `gotest ./...` per package.
5. `gopush` per package.

---

## Publishing

```bash
cd ~/Dev/Project/tinywasm/fmt
# Rewrite docs/TRANSLATE.md first (required by project rules)
gopush  # tags as v0.18.0
```
