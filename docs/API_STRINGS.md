# Strings Package Equivalents

Replace common `strings` package functions with fmt equivalents:

| Go Standard | fmt Equivalent |
|-------------|----------------------|
| `strings.Builder` | `c:= Convert() c.Write(a) c.Write(b) c.String()` |
| `strings.Contains()` | `Contains(s, substr)` |
| `strings.Index()` | `Index(s, substr)` |
| `strings.LastIndex()` | `LastIndex(s, substr)` |
| `strings.Join()` | `Convert(slice).Join(sep).String()` |
| `strings.Repeat()` | `Convert(s).Repeat(n).String()` |
| `strings.Replace()` | `Convert(s).Replace(old, new).String()` |
| `strings.Split()` | `Convert(s).Split(sep).String()` |
| `strings.ToLower()` | `Convert(s).ToLower().String()` |
| `strings.ToUpper()` | `Convert(s).ToUpper().String()` |
| `strings.TrimSpace()` | `Convert(s).TrimSpace().String()` |
| `strings.TrimPrefix()` | `Convert(s).TrimPrefix(prefix).String()` |
| `strings.TrimSuffix()` | `Convert(s).TrimSuffix(suffix).String()` |
| `strings.HasPrefix()` | `HasPrefix(s, prefix)` |
| (Utility) | `HasUpperPrefix(s)` |
| `strings.HasSuffix()` | `HasSuffix(s, suffix)` |

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

// ⚠️ Note: Index, Contains, LastIndex and HasUpperPrefix are global functions, not methods.
// Do NOT use: Convert(s).Contains(substr) // ❌ Incorrect, will not compile
// Use:        Index(s, substr)            // ✅ Correct
//             Contains(s, substr)         // ✅ Correct
//             LastIndex(s, substr)        // ✅ Correct
//             HasUpperPrefix(s)           // ✅ Correct

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