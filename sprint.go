package fmt

// Sprint converts any value to its string representation.
// This is the equivalent of fmt.Sprint() from the standard library.
// Example: Sprint(42) returns "42"
// Example: Sprint(true) returns "true"
// Example: Sprint("hello") returns "hello"
func Sprint(v any) string {
	return Convert(v).String()
}
