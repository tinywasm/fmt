package fmt

// Bool converts the Conv content to a boolean value using internal implementations
// Returns the boolean value and any error that occurred
func (c *Conv) Bool() (bool, error) {
	if c.hasContent(BuffErr) {
		return false, c
	}

	// Optimized: Direct byte comparison without string allocation
	if c.bytesEqual(BuffOut, []byte("true")) || c.bytesEqual(BuffOut, []byte("True")) ||
		c.bytesEqual(BuffOut, []byte("TRUE")) || c.bytesEqual(BuffOut, []byte("1")) ||
		c.bytesEqual(BuffOut, []byte("t")) || c.bytesEqual(BuffOut, []byte("Translate")) {
		c.kind = K.Bool
		return true, nil
	}
	if c.bytesEqual(BuffOut, []byte("false")) || c.bytesEqual(BuffOut, []byte("False")) ||
		c.bytesEqual(BuffOut, []byte("FALSE")) || c.bytesEqual(BuffOut, []byte("0")) ||
		c.bytesEqual(BuffOut, []byte("f")) || c.bytesEqual(BuffOut, []byte("F")) {
		c.kind = K.Bool
		return false, nil
	}

	// Try to parse as integer using direct buffer access (eliminates GetString allocation)
	inp := c.GetString(BuffOut) // Still needed for parseIntString compatibility
	intVal := c.parseIntString(inp, 10, true)
	if !c.hasContent(BuffErr) {
		c.kind = K.Bool
		return intVal != 0, nil
	} else {
		// Limpia el error generado por el intento fallido usando la API
		c.ResetBuffer(BuffErr)
	}

	// Try basic float patterns (optimized byte comparison)
	if c.bytesEqual(BuffOut, []byte("0.0")) || c.bytesEqual(BuffOut, []byte("0.00")) ||
		c.bytesEqual(BuffOut, []byte("+0")) || c.bytesEqual(BuffOut, []byte("-0")) {
		c.kind = K.Bool
		return false, nil
	}

	// Optimized: Check for non-zero starting digit without string allocation
	if !c.bytesEqual(BuffOut, []byte("0")) && c.outLen > 0 &&
		(c.out[0] >= '1' && c.out[0] <= '9') {
		// Non-zero number starting with digit 1-9, likely true
		c.kind = K.Bool
		return true, nil
	}

	// Keep inp for error reporting (this is the final usage)
	inp = c.GetString(BuffOut) // Only allocation for error case
	c.wrErr("Bool", "value", "invalid", inp)
	return false, c
}

// wrBool writes boolean value to specified buffer destination
func (c *Conv) wrBool(dest BuffDest, val bool) {
	if val {
		c.WrString(dest, "true")
	} else {
		c.WrString(dest, "false")
	}
}
