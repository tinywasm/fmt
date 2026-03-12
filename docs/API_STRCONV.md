# Strconv Package Equivalents

Replace `strconv` package functions for type conversions:

| Go Standard | fmt Equivalent |
|-------------|----------------------|
| `strconv.Itoa()` | `Convert(i).String()` |
| `strconv.Atoi()` | `Convert(s).Int()` |
| `strconv.ParseFloat()` | `Convert(s).Float64()` |
| `strconv.ParseBool()` | `Convert(s).Bool()` |
| `strconv.FormatFloat()` | `Convert(f).Round(n).String()` |
| `strconv.Quote()` | `Convert(s).Quote().String()` |

## Type Conversions

```go
// String to numbers => Int,Int32,Int64,Uint,Uint32,Uint64,Float32,Float64 eg:
result, err := Convert("123").Int()        // out: 123, nil
result, err := Convert("456").Uint()       // out: 456, nil
result, err := Convert("3.14").Float64()     // out: 3.14, nil
result, err := Convert("1.5e3").Float64()    // out: 1500, nil (scientific notation)

// Numbers to string
Convert(42).String()      // out: "42"
Convert(3.14159).String() // out: "3.14159"

// Boolean conversions
result, err := Convert("true").Bool()  // out: true, nil
result, err := Convert(42).Bool()      // out: true, nil (non-zero = true)
result, err := Convert(0).Bool()       // out: false, nil

// String quoting
Convert("hello").Quote().String()           // out: "\"hello\""
Convert("say \"hello\"").Quote().String()  // out: "\"say \\\"hello\\\"\""
```

## Number Formatting

```go
// Decimal rounding: keep N decimals, round or truncate
// By default, rounds using "round half to even" (bankers rounding)
// Pass true as the second argument to truncate (no rounding), e.g.:
Convert("3.14159").Round(2).String()        // "3.14" (rounded)
Convert("3.155").Round(2).String()          // "3.16" (rounded)
Convert("3.14159").Round(2, true).String()  // "3.14" (truncated, NOT rounded)
Convert("3.159").Round(2, true).String()    // "3.15" (truncated, NOT rounded)

// Formatting with thousands separator (EU default)
Convert(2189009.00).Thousands().String()        // out: "2.189.009"
// Anglo/US style (comma, dot)
Convert(2189009.00).Thousands(true).String()    // out: "2,189,009"
```

## Low-level Writing (Performance)

High-performance codecs can write numbers directly to the output buffer without creating high-level `any` boxes:

```go
c := fmt.Convert()
c.WriteInt(123)
c.WriteFloat(3.14)
result := c.String() // "1233.14"
```