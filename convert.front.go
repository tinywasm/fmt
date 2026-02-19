//go:build wasm

package fmt

// anyToBuffFallback handles unknown types in WASM (no reflect)
func (c *Conv) anyToBuffFallback(dest BuffDest, value any) {
	// Check Stringer interface (still works without reflect)
	if stringer, ok := value.(interface{ String() string }); ok {
		c.kind = K.String
		c.WrString(dest, stringer.String())
		return
	}
	c.wrErr("type", "not", "supported")
}
