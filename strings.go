package fmt

// This file exposes drop-in replacements matching the stdlib "strings" package
// signatures, implemented on top of the existing Convert/Conv chain API.
// It exists so call sites can migrate off "strings" without needing to learn
// the chained API surface, avoiding common mistakes such as
// Convert(s).Contains(x) (Contains/Index/LastIndex are package functions, not methods).
// See docs/API_STRINGS.md for the full equivalence table.

// TrimSpace mirrors strings.TrimSpace.
func TrimSpace(s string) string {
	return Convert(s).TrimSpace().String()
}

// TrimPrefix mirrors strings.TrimPrefix.
func TrimPrefix(s, prefix string) string {
	return Convert(s).TrimPrefix(prefix).String()
}

// TrimSuffix mirrors strings.TrimSuffix.
func TrimSuffix(s, suffix string) string {
	return Convert(s).TrimSuffix(suffix).String()
}

// ToLower mirrors strings.ToLower.
func ToLower(s string) string {
	return Convert(s).ToLower().String()
}

// ToUpper mirrors strings.ToUpper.
func ToUpper(s string) string {
	return Convert(s).ToUpper().String()
}

// Repeat mirrors strings.Repeat.
func Repeat(s string, count int) string {
	return Convert(s).Repeat(count).String()
}

// ReplaceAll mirrors strings.ReplaceAll.
func ReplaceAll(s, old, new string) string {
	return Convert(s).Replace(old, new).String()
}

// ReplaceN mirrors strings.Replace(s, old, new, n).
// Named ReplaceN (not Replace) to avoid colliding with Conv.Replace's
// signature, which accepts values of any type instead of only strings.
func ReplaceN(s, old, new string, n int) string {
	return Convert(s).Replace(old, new, n).String()
}

// JoinSlice mirrors strings.Join(elems, sep).
// Named JoinSlice (not Join) to avoid colliding with Conv.Join.
func JoinSlice(elems []string, sep string) string {
	return Convert(elems).Join(sep).String()
}
