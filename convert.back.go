//go:build !wasm

package fmt

import "reflect"

// anyToBuffFallback handles custom types via reflection (backend only)
func (c *Conv) anyToBuffFallback(dest BuffDest, value any) {
	// Check Stringer interface first
	if stringer, ok := value.(interface{ String() string }); ok {
		c.kind = K.String
		c.WrString(dest, stringer.String())
		return
	}

	// Reflection fallback for custom types
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		c.AnyToBuff(dest, rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		c.AnyToBuff(dest, rv.Uint())
	case reflect.Float32, reflect.Float64:
		c.AnyToBuff(dest, rv.Float())
	case reflect.String:
		c.AnyToBuff(dest, rv.String())
	default:
		c.wrErr("type", "not", "supported")
	}
}
