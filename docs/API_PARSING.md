# Parsing API

The `fmt` package provides tools for parsing structured strings, such as Go struct tags and key-value pairs.

## Struct Tag Parsing

### TagValue
`TagValue(key string) (string, bool)`

Searches for the value of a specific key in a Go struct tag-like string.

```go
val, found := fmt.Convert(`json:"name" Label:"Nombre"`).TagValue("Label")
// val: "Nombre", found: true
```

### TagPairs
`TagPairs(key string) []KeyValue`

Parses a Go struct tag-like string and extracts multiple key-value pairs from a specific tag's value. This is useful for tags that contain comma-separated key-value pairs.

- Pairs are separated by `,`.
- Key and value within a pair are separated by `:`.
- If a pair does not contain a `:`, it is considered invalid and skipped.

```go
pairs := fmt.Convert(`options:"key1:text1,key2:text2"`).TagPairs("options")
// pairs: []KeyValue{{Key: "key1", Value: "text1"}, {Key: "key2", Value: "text2"}}
```

## Key-Value Extraction

### ExtractValue
`ExtractValue(delimiters ...string) (string, error)`

Extracts the value part from a "key:value" string.

```go
// Default delimiter (:)
val, err := fmt.Convert("key:value").ExtractValue()
// val: "value", err: nil

// Custom delimiter
val, err := fmt.Convert("key=value").ExtractValue("=")
// val: "value", err: nil
```

## Types

### KeyValue
Represents a simple key-value pair extracted from a string.

```go
type KeyValue struct {
    Key   string // The identifier or key
    Value string // The associated value or text
}
```