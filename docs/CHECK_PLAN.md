# PLAN — fmt: map-free, algorithmic snake→Pascal for Go identifiers

> Dispatched via CodeJob. Skill: **agents-workflow**. Single repo: `github.com/tinywasm/fmt`.
> Self-contained: every contract, rule and example is inline.
> **Constraint: no `map` in library code (TinyGo target).** Maps allowed only in `_test.go`.

## 0. TL;DR

`fmt` owns case conversion but has **no working snake→Pascal path**: `CamelUp` does not split on
`_`/`-` (in `toCaseTransformMinimal` the only CamelCase boundaries are whitespace, lower→upper and
digit→letter; `_` is preserved), so `CamelUp("tenant_id")` → `Tenant_id`.

Downstream `ormc` therefore hand-rolled its own `ToPascalCase` **plus a `knownUpper` initialism map**
(`orm/ormc/parse_definition.go:373`). That map is (a) forbidden here (TinyGo), and (b) an unbounded,
unmaintainable list of acronyms.

**The fix is purely algorithmic and carries no data.** Extend the existing CamelCase transform to
treat `_` and `-` as word separators. Acronyms are **not** special-cased — see §1.

## 1. The irreducible asymmetry (why there is no map)

- **Pascal→snake is algorithmic and lossless.** `SnakeLow` already yields `SKU`→`sku`,
  `TenantID`→`tenant_id`. (One acronym-run gap noted in §4, still map-free.)
- **snake→Pascal is lossy for acronyms and cannot be recovered automatically.** `sku`→`SKU` is
  information-theoretically indistinguishable from `abc`→`Abc`: the uppercase-ness was destroyed by
  the snake encoding. Recovering it requires **external knowledge** (a dictionary), which is exactly
  what we refuse to maintain.

**Consequence (accepted):** the generated Go name is a plain per-word capitalization. Acronyms come
out title-cased:

| Column (source of truth) | Generated Go field |
|---|---|
| `id` | `Id` |
| `sku` | `Sku` |
| `tenant_id` | `TenantId` |
| `is_active` | `IsActive` |
| `updated_at` | `UpdatedAt` |
| `api_response` | `ApiResponse` |
| `name` | `Name` |

This is deterministic, reversible (`SnakeLow("TenantId") == "tenant_id"`), and requires zero data.
It is **not** golint-idiomatic (`Id` vs `ID`), which is the price of having no dictionary.

> **Decision (§7): pure form, no exceptions.** A length/vowel heuristic to auto-recover acronym
> casing was evaluated and rejected — it cannot work (acronym-ness is semantic, not orthographic).

## 2. Change (map-free, algorithmic)

### 2.1 Treat `_` and `-` as word separators in CamelCase mode — `fmt/capitalize.go`

In `toCaseTransformMinimal`, for CamelCase mode (`separator == ""`), classify `_` and `-` as
separators alongside whitespace: they start a new word and are **not** emitted. Snake/Kebab mode
(`separator != ""`) is unchanged.

Blast radius is nil: `CamelUp`/`CamelLow` have **no external callers** in the tinywasm tree
(`grep -rln '.CamelUp()\|.CamelLow()' | grep -v /fmt/` → only `fmt/*_test.go`). Existing
space-separated behavior (`CamelUp("hello world")`→`HelloWorld`) is preserved.

Result:
- `Convert("tenant_id").CamelUp()` → `TenantId`
- `Convert("is_active").CamelUp()` → `IsActive`
- `Convert("sku").CamelUp()` → `Sku`
- `Convert("first_name").CamelLow()` → `firstName`

No new methods. `CamelUp`/`CamelLow` already are the public API; after §2.1 they handle snake/kebab
input directly. Call site stays `fmt.Convert(col).CamelUp().String()`.

## 3. Consumer migration (`tinywasm/orm/ormc`, separate repo, after publish)

- **Delete** `ToPascalCase` **and** the `knownUpper` map from `parse_definition.go`.
- Replace `ToPascalCase(fi.ColumnName)` with `fmt.Convert(fi.ColumnName).CamelUp().String()`.
- The generated struct field, `Pointers`, `Encode/DecodeFields`, and the `<Struct>_` typed-fields
  helper all inherit the new casing (`Id`, `Sku`, `TenantId`).

## 4. (Verify, map-free) Pascal→snake acronym runs

`SnakeLow` splits only on lower→upper, so a leading acronym run like `APIResponse` currently yields
`apiresponse` instead of `api_response`. If round-trip tests (§5) surface this, fix it algorithmically
by adding one boundary rule — "split before an uppercase that is followed by a lowercase, when the
previous char is also uppercase" (`API|Response`, `HTTP|Server`). Pure logic, **no map**. Fix only if
a test fails; forward snake→Pascal is the blocking direction.

## 5. Tests (`fmt/goname_test.go`, new — tests MAY use maps/slices)

Table-driven:

| Input | `CamelUp` | `CamelLow` |
|---|---|---|
| `id` | `Id` | `id` |
| `sku` | `Sku` | `sku` |
| `tenant_id` | `TenantId` | `tenantId` |
| `is_active` | `IsActive` | `isActive` |
| `updated_at` | `UpdatedAt` | `updatedAt` |
| `api_response` | `ApiResponse` | `apiResponse` |
| `user123_name` | `User123Name` | `user123Name` |
| `name` | `Name` | `name` |
| `first-name` (kebab) | `FirstName` | `firstName` |
| `hello world` (space, unchanged) | `HelloWorld` | `helloWorld` |

Plus round-trip: for each snake row, `CamelUp` then `SnakeLow` returns the original snake string.
Run `go test -race ./...` on the stdlib target and confirm the WASM/TinyGo build still compiles.

## 6. Acceptance criteria

- `gotest ./...` green; WASM build compiles; **no `map` added to any non-`_test.go` file**.
- `Convert("tenant_id").CamelUp().String() == "TenantId"`, `…("sku")… == "Sku"`.
- Existing `CamelUp`/`CamelLow`/`SnakeLow` tests (space-separated) unchanged and passing.
- Round-trip identity `snake → CamelUp → SnakeLow == snake` holds for the §5 table.
- Published (`gopush`) so `orm` can `go get` and delete `ToPascalCase` + `knownUpper`.

## 7. RESOLVED DECISION — pure algorithmic (no heuristic, no data)

**Chosen:** pure per-word capitalization. `id`→`Id`, `sku`→`Sku`, `tenant_id`→`TenantId`. No map,
no list, no heuristic. Consuming modules use `.Id`, `.Sku`, `.TenantId`.

**Rejected — length/vowel heuristic** (do not re-propose): a rule like "uppercase words ≤3 letters"
or "uppercase words with no vowels" cannot separate acronyms from short common words, because
acronym-ness is *semantic, not orthographic*, and the snake encoding already destroyed the casing.
Counterexamples that break every such rule simultaneously:

| word | want | "≤3 letters"→UPPER | "no vowels"→UPPER |
|---|---|---|---|
| `sku` | `SKU` | SKU ✓ | SKU ✓ |
| `id` | `ID` | ID ✓ | `Id` ✗ (has `i`) |
| `url` | `URL` | URL ✓ | `Url` ✗ (has `u`) |
| `api` | `API` | API ✓ | `Api` ✗ |
| `ip` | `IP` | IP ✓ | `Ip` ✗ |
| `at` (`updated_at`) | `At` | `AT` ✗ | `At` ✓ |
| `is` (`is_active`) | `Is` | `IS` ✗ | `Is` ✓ |
| `to`/`by`/`of` | `To`/… | `TO`/… ✗ | `To`/… ✓ |

No rule gets all rows right; each produces both false positives (`AT`, `IS`) and false negatives
(`Url`, `Api`). A heuristic would also *add* code for *wrong* results — worse for binary size than
the pure path, which adds **zero** logic (§2.1 only reuses `toCaseTransformMinimal`).

## 8. Stages

| # | Stage | Output | Gate |
|---|---|---|---|
| 1 | `_`/`-` separators in CamelCase (§2.1) | `CamelUp`/`CamelLow` snake-aware | §5 table passes |
| 2 | Tests + round-trip + race (§5) | `goname_test.go` | green; WASM compiles; no maps |
| 3 | (If §4 test fails) acronym-run split | `SnakeLow` boundary rule | `api_response` correct |
| 4 | Publish (§6) | tagged module | `orm` can delegate |

## 9. Downstream sequencing

1. **fmt** (this plan) — publish snake-aware `CamelUp`/`CamelLow`.
2. **orm** — `orm/docs/PLAN.md`: wire `// orm:typed_fields` directive **and** delete
   `ToPascalCase`/`knownUpper`, delegate to `fmt.CamelUp`. Publish.
3. **item_catalog** — regenerate; field names/typed-fields helper adopt the §7 casing; update
   `mcp.go` references accordingly (`.Id`/`.Sku`/… per the chosen policy).
