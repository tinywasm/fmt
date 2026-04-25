# PLAN: tinywasm/fmt — Add missing dictionary words

## Goal
Add the words required by `tinywasm/app` Stage 3 error messages to the built-in dictionary.
All words follow Noun+Adjective order so they produce natural output in both EN and ES.

## Dependency
This plan must be executed before `tinywasm/app` Stage 3.

## Words to Add

The following words are missing from `dictionary/dictionary.go` and are needed for
the "not inside a Go project" error message in `tinywasm/app`.

Composed message pattern: `"Directory", "Go", "Not", "Initialized"`
- EN output: "Directory Go Not Initialized"
- ES output: "Directorio Go No Inicializado"

| EN | ES | Section |
|----|-----|---------|
| `"Directory"` | `"Directorio"` | D |
| `"Initialized"` | `"Inicializado"` | I |
| `"Project"` | `"Proyecto"` | P |
| `"Root"` | `"Raíz"` | R |
| `"Run"` | `"Ejecutar"` | R |

Note: `"Not"` and `"Go"` are already handled — `"Not"` exists in the dictionary,
`"Go"` passes through untranslated (proper noun / language name).

## Files to Modify

| File | Change |
|------|--------|
| [dictionary/dictionary.go](../dictionary/dictionary.go) | Add the 5 words above in their correct alphabetical sections |

## Steps

- [ ] Add `"Directory"` / `"Directorio"` in section D (after `"Debugging"`)
- [ ] Add `"Initialized"` / `"Inicializado"` in section I (after `"Install"`)
- [ ] Add `"Project"` / `"Proyecto"` in section P (after `"Production"`)
- [ ] Add `"Root"` / `"Raíz"` in section R (after `"Right"`)
- [ ] Add `"Run"` / `"Ejecutar"` in section R (after `"Root"`)
- [ ] Run `gotest ./dictionary/...` — all must pass
- [ ] Notify `tinywasm/app` Stage 3 that the dependency is satisfied
