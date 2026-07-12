# PLAN — `tinywasm/fmt`: bulk-copy safe runs in `JSONEscape` (profiled fix, low priority)

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
> Not urgent — performance only, no bug/breakage. Dispatch whenever convenient.
> Repo rules: `AGENTS.md` at this repo's root — read it first.

## Context (zero-context summary)

`tinywasm/json`'s benchmark (`json/tests/bench_encode_test.go`,
`BenchmarkEncode_tinywasm` vs `BenchmarkEncode_stdlib`) shows encoding a small
4-field struct is barely faster than `encoding/json` (753 ns/op vs 758 ns/op,
same 1 alloc/op) despite `tinywasm/json` having zero reflection — a result
worth investigating rather than accepting, since the whole point of the
zero-reflection codec is to be faster, not merely smaller.

**Real CPU profile taken 2026-07-11** (`go test -bench=BenchmarkEncode_tinywasm
-benchtime=200000x -cpuprofile=...`, then `go tool pprof -top`/`-list`,
against `github.com/tinywasm/fmt@v0.25.2`) found the actual cause. Top of
`-top -nodecount=20`, 160ms total samples:

```
      flat  flat%   sum%        cum   cum%
      50ms 31.25% 31.25%       70ms 43.75%  github.com/tinywasm/fmt.JSONEscape
      30ms 18.75% 50.00%      150ms 93.75%  github.com/tinywasm/json.Encode
      20ms 12.50% 62.50%       30ms 18.75%  github.com/tinywasm/fmt.(*Conv).WriteByte
      20ms 12.50% 75.00%       20ms 12.50%  github.com/tinywasm/fmt.(*Conv).wrByte (inline)
```

`-list JSONEscape` pinpoints it further — of `JSONEscape`'s 70ms cumulative,
40ms (25% of the ENTIRE benchmark) is this single line:

```
quote.go:74:  _ = b.WriteByte(c)     // 20ms flat, 40ms cum — the "safe byte" path
```

`JSONEscape` (`fmt/quote.go:54-78`, current implementation, no build tag —
shared wasm+backend code) writes one byte at a time even for the common case
(a character that needs no escaping):

```go
func JSONEscape(s string, b *Builder) {
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			b.WriteString(`\"`)
		// ... \\, \n, \r, \t cases ...
		default:
			if c < 0x20 {
				b.WriteString(`\u00`)
				_ = b.WriteByte("0123456789abcdef"[c>>4])
				_ = b.WriteByte("0123456789abcdef"[c&0xf])
			} else {
				_ = b.WriteByte(c)   // ← every ordinary byte pays a full function-call chain
			}
		}
	}
}
```

Every "safe" byte pays the full chain `JSONEscape` → `Conv.WriteByte` →
`Conv.wrByte` → `append(c.out[:c.outLen], b)` — 3 function calls plus a
single-element append — INSTEAD of one bulk `append(dst, src...)` for the
whole safe run. `encoding/json` avoids this by scanning for the next special
character (`bytealg`/`IndexByte`, assembly-accelerated) and copying the safe
span in one `WriteString`/`Write` call; that's the actual reason the
zero-reflection codec isn't measurably faster here — reflection savings are
being spent right back on this loop.

**Second finding, arguably bigger:** `jsonWriter.writeKey`
(`json/encode.go:23-30`) calls `JSONEscape` on the JSON **field name** on
every single encode call — `"name"`, `"email"`, `"age"`, `"score"` — even
though these are Go string literals hardcoded by generated (`ormc`) or
hand-written `EncodeFields` bodies and NEVER contain characters needing
escape. `-list jsonWriter.String` shows the split precisely: of `String()`'s
40ms cumulative, 30ms (18.75% of the ENTIRE benchmark) is `writeKey`
(escaping the key), only 10ms is escaping the actual value. Across the whole
benchmark, `writeKey` alone accounts for **37.5% of total CPU time**
(60ms/160ms, confirmed via `-list writeKey`) — more than double the cost of
escaping the real data.

Both findings have the SAME fix: `JSONEscape` doing bulk-copy instead of
per-byte. Fixing it once fixes both hot spots (key-escaping is dominant
today only because it's ALSO paying the per-byte tax for zero actual escape
work).

## Target implementation

Rewrite `fmt/quote.go`'s `JSONEscape` to scan for the next byte needing
escape and bulk-copy the safe span with the already-existing
`Builder.WriteString` (→ `Conv.WrString` → `Conv.wrBytes` →
`append(dst, src...)`, ALREADY a single bulk append — verified in
`fmt/memory.go:44-52`, no new primitive needed):

```go
func JSONEscape(s string, b *Builder) {
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 0x20 && c != '"' && c != '\\' {
			continue // safe byte — part of the current bulk run
		}
		if i > start {
			b.WriteString(s[start:i])
		}
		switch c {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteString(`\u00`)
			_ = b.WriteByte("0123456789abcdef"[c>>4])
			_ = b.WriteByte("0123456789abcdef"[c&0xf])
		}
		start = i + 1
	}
	if start < len(s) {
		b.WriteString(s[start:])
	}
}
```

Escaping RULES are unchanged byte-for-byte (same 5 named cases + same
`\u00XX` control-char path) — only the safe-byte path changes from N
`WriteByte` calls to a single `WriteString` per contiguous safe run. For a
string with zero special characters (the common case, and 100% of the field
names in the benchmark), this collapses the whole function to ONE
`WriteString` call instead of `len(s)` `WriteByte` calls.

## Stages

### Stage 1 — rewrite `JSONEscape`

Replace the function body in `fmt/quote.go` with the target implementation
above. No signature change (`func JSONEscape(s string, b *Builder)` stays
identical) — this is a pure internal optimization, zero API surface change,
zero build-tag change (stays shared wasm+backend code, same as today).

### Stage 2 — output-equivalence tests

Add/extend tests asserting byte-identical output vs. the OLD implementation
for: empty string, string with no special chars (common case — assert it
still produces correct output, this is what regresses if the scan logic has
an off-by-one), string starting/ending with a special char (edge of a safe
run), string that is ENTIRELY special chars (every byte triggers a flush),
consecutive special chars, a control character (`` etc — the `\u00XX`
path), and a realistic mixed string (`"alice@example.com"`, `"Alice \"Bob\""`).

Run with `gotest` (never `go test`).

### Stage 3 — re-profile to confirm the fix (this is the acceptance criterion, not just "tests pass")

```bash
cd json/tests
go test -run=^$ -bench=BenchmarkEncode_tinywasm -benchtime=200000x -cpuprofile=/tmp/cpu_after.prof -o /tmp/encode_after.test .
go tool pprof -top -nodecount=10 /tmp/encode_after.test /tmp/cpu_after.prof
```

Expect `fmt.JSONEscape` to drop out of the top of the profile (or shrink to a
small fraction of its current 43.75% cumulative). This requires bumping
`json`'s `go.mod` to the new `fmt` version first (or using a local
`replace` directive during verification) — see `json/docs/PLAN.md` in this
same monorepo, queued to run right after this plan closes, which re-profiles
end-to-end and updates the benchmark numbers published in `json/README.md`.

### Stage 4 — `docs/API_JSON_ESCAPE.md`

No behavioral change to document (escaping rules identical) — add one line
noting the safe-byte path is now bulk-copied for performance, no observable
difference to callers.

## Anti-footguns (do NOT do)

- Do NOT change the escaping RULES (which characters get escaped, or how) —
  this is a pure performance rewrite; any output difference is a bug.
- Do NOT add a build tag to `quote.go` — it must stay shared wasm+backend
  code exactly as today.
- Do NOT touch `Conv.WriteByte`/`Conv.wrByte`/`Conv.WrString` — they are
  already correct and already do bulk appends where called in bulk; the bug
  was entirely in `JSONEscape` calling them one byte at a time.
- Do NOT touch `jsonWriter.writeKey` in `tinywasm/json` — out of scope for
  this repo; it benefits automatically once `JSONEscape` is fixed, and
  `json/docs/PLAN.md` verifies that.
- Never run `gopush` or `codejob`.

## Stages table

| # | Stage | Files | Done |
|---|---|---|---|
| 1 | Rewrite `JSONEscape` (bulk-copy safe runs) | `quote.go` | ☐ |
| 2 | Byte-identical output tests | `*_test.go` (quote-related) | ☐ |
| 3 | Re-profile, confirm hotspot gone | — | ☐ |
| 4 | Doc note | `docs/API_JSON_ESCAPE.md` | ☐ |
