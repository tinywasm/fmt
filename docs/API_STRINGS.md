# Strings Package Equivalents

Replace common `strings` package functions with fmt equivalents:

| Go Standard | fmt Equivalent |
|-------------|----------------------|
| `strings.Builder` | `c:= Convert() c.Write(a) c.Write(b) c.String()` |
| `strings.Contains()` | `Contains(s, substr)` |
| `strings.Index()` | `Index(s, substr)` |
| `strings.LastIndex()` | `LastIndex(s, substr)` |
| `strings.Join()` | `Convert(slice).Join(sep).String()` or `JoinSlice(elems, sep)` |
| `strings.Repeat()` | `Convert(s).Repeat(n).String()` or `Repeat(s, n)` |
| `strings.Replace()` | `Convert(s).Replace(old, new).String()` or `ReplaceN(s, old, new, n)` |
| `strings.ReplaceAll()` | `ReplaceAll(s, old, new)` |
| `strings.Split()` | `Split(s, separator...)` |
| `strings.ToLower()` | `Convert(s).ToLower().String()` or `ToLower(s)` |
| `strings.ToUpper()` | `Convert(s).ToUpper().String()` or `ToUpper(s)` |
| `strings.TrimSpace()` | `Convert(s).TrimSpace().String()` or `TrimSpace(s)` |
| `strings.TrimPrefix()` | `Convert(s).TrimPrefix(prefix).String()` or `TrimPrefix(s, prefix)` |
| `strings.TrimSuffix()` | `Convert(s).TrimSuffix(suffix).String()` or `TrimSuffix(s, suffix)` |
| `strings.HasPrefix()` | `HasPrefix(s, prefix)` |
| (Utility) | `HasUpperPrefix(s)` |
| `strings.HasSuffix()` | `HasSuffix(s, suffix)` |
| (Utility) | `Matches(s, terms...)` |
| (Utility) | `MatchesAny(s, terms...)` |

## Drop-in stdlib-signature wrappers (strings.go)

`Contains`, `Index`, `LastIndex`, `HasPrefix`, `HasSuffix`, `Split` already
match the stdlib `strings` signature and are used directly.

For the chain-only operations (`TrimSpace`, `TrimPrefix`, `TrimSuffix`,
`ToLower`, `ToUpper`, `Repeat`, `Replace`, `Join`), `strings.go` adds thin
package-level functions with the exact stdlib signature, implemented on top
of the `Convert(...)` chain. This avoids the most common migration mistake —
calling a chain method as if it were a package function (or vice versa)
without knowing the full chain API:

| Go Standard | fmt Equivalent |
|-------------|----------------------|
| `strings.TrimSpace(s)` | `TrimSpace(s)` |
| `strings.TrimPrefix(s, prefix)` | `TrimPrefix(s, prefix)` |
| `strings.TrimSuffix(s, suffix)` | `TrimSuffix(s, suffix)` |
| `strings.ToLower(s)` | `ToLower(s)` |
| `strings.ToUpper(s)` | `ToUpper(s)` |
| `strings.Repeat(s, count)` | `Repeat(s, count)` |
| `strings.ReplaceAll(s, old, new)` | `ReplaceAll(s, old, new)` |
| `strings.Replace(s, old, new, n)` | `ReplaceN(s, old, new, n)` |
| `strings.Join(elems, sep)` | `JoinSlice(elems, sep)` |

`ReplaceN` and `JoinSlice` are named differently from `Replace`/`Join`
because those names are already taken by `Conv.Replace`/`Conv.Join` (chain
methods with a different signature: `Conv.Replace` accepts `any` values, and
`Conv.Join` takes the slice as the receiver instead of a parameter).

## Other String Transformations

```go
Convert("hello world").CamelLow().String() // out: "helloWorld"
Convert("hello world").CamelUp().String()  // out: "HelloWorld"
Convert("hello world").SnakeLow().String() // out: "hello_world"
Convert("hello world").SnakeUp().String()  // out: "HELLO_WORLD"
```

## String Search & Operations

```go
// Search and count
pos := Index("hello world", "world")                  // out: 6 (first occurrence)
found := Contains("hello world", "world")              // out: true
count := Count("abracadabra", "abra")       // out: 2

// Prefix / Suffix checks
isPref := HasPrefix("hello", "he")          // out: true
isUpper := HasUpperPrefix("Hello")          // out: true
isSuf := HasSuffix("file.txt", ".txt")      // out: true

// Note: this library follows the standard library semantics for prefixes/suffixes:
// an empty prefix or suffix is considered a match (HasPrefix(s, "") == true,
// HasSuffix(s, "") == true).

// Find last occurrence (useful for file extensions)
pos := LastIndex("image.backup.jpg", ".")             // out: 12
if pos >= 0 {
    extension := "image.backup.jpg"[pos+1:]           // out: "jpg"
}

// Multi-term searching (AND/OR semantics)
// Both perform internal ToLower() normalization on content and terms.
foundAll := Matches("Hello World", "hello", "world")   // out: true (AND)
foundOne := MatchesAny("Hello World", "hello", "xyz")  // out: true (OR)
noMatch  := Matches("Hello World", "hello", "xyz")     // out: false (AND)

// ⚠️ Note: Index, Contains, LastIndex, HasUpperPrefix, Matches and MatchesAny are global functions, not methods.
// Do NOT use: Convert(s).Contains(substr) // ❌ Incorrect, will not compile
// Use:        Index(s, substr)            // ✅ Correct
//             Contains(s, substr)         // ✅ Correct
//             LastIndex(s, substr)        // ✅ Correct
//             HasUpperPrefix(s)           // ✅ Correct
//             Matches(s, terms...)        // ✅ Correct
//             MatchesAny(s, terms...)     // ✅ Correct

// Replace operations
Convert("hello world").Replace("world", "Go").String() // out: "hello Go"
Convert("test 123 test").Replace(123, 456).String()    // out: "test 456 test"
```

## String Splitting & Joining

```go
// Split strings (always use Convert(...).Split(...))
parts := Convert("apple,banana,cherry").Split(",")
// out: []string{"apple", "banana", "cherry"}

parts := Convert("hello world new").Split()  // Handles whitespace
// out: []string{"hello", "world", "new"}

// Join slices
Convert([]string{"Hello", "World"}).Join().String()    // out: "Hello World"
Convert([]string{"a", "b", "c"}).Join("-").String()    // out: "a-b-c"
```

## String Trimming & Cleaning

```go
// TrimSpace operations
Convert("  hello  ").TrimSpace().String()                    // out: "hello"
Convert("prefix-data").TrimPrefix("prefix-").String()   // out: "data"
Convert("file.txt").TrimSuffix(".txt").String()         // out: "file"

// Repeat strings
Convert("Go").Repeat(3).String()                        // out: "GoGoGo"
```