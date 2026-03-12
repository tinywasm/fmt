package fmt

// =============================================================================
// FLOAT OPERATIONS - All float parsing, conversion and formatting
// =============================================================================

// Float64 converts the value to a float64.
// Returns the converted float64 and any error that occurred during conversion.
func (c *Conv) Float64() (float64, error) {
	val := c.parseFloatBase()
	if c.hasContent(BuffErr) {
		return 0, c
	}
	return val, nil
}

// toFloat64 converts various float types to float64
func (c *Conv) toFloat64(arg any) (float64, bool) {
	switch v := arg.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	default:
		// Try reflection for custom types (e.g., type customFloat float64)
		return c.toFloat64Reflect(arg)
	}
}

// Float32 converts the value to a float32.
// Returns the converted float32 and any error that occurred during conversion.
func (c *Conv) Float32() (float32, error) {
	val := c.parseFloatBase()
	if c.hasContent(BuffErr) {
		return 0, c
	}
	if val > 3.4028235e+38 {
		return 0, c.wrErr("number", "overflow")
	}
	return float32(val), nil
}

// parseFloatBase parses the buffer as a float64, similar to parseIntBase for ints.
// It always uses the buffer output and handles errors internally.
func (c *Conv) parseFloatBase() float64 {
	c.ResetBuffer(BuffErr)

	s := c.GetStringZeroCopy(BuffOut)
	if len(s) == 0 {
		c.wrErr("string", "empty")
		return 0
	}

	var result float64
	var negative bool
	var hasDecimal bool
	var hasDigits bool
	var decimalPlaces int
	i := 0

	// Handle sign
	switch s[0] {
	case '-':
		negative = true
		i = 1
		if len(s) == 1 {
			c.wrErr("format", "invalid")
			return 0
		}
	case '+':
		i = 1
		if len(s) == 1 {
			c.wrErr("format", "invalid")
			return 0
		}
	}

	// Parse integer part
	for ; i < len(s) && s[i] != '.' && s[i] != 'e' && s[i] != 'E'; i++ {
		if s[i] < '0' || s[i] > '9' {
			c.wrErr("character", "invalid")
			return 0
		}
		hasDigits = true
		result = result*10 + float64(s[i]-'0')
	}

	// Parse decimal part if present
	if i < len(s) && s[i] == '.' {
		hasDecimal = true
		i++ // Skip decimal point
		for ; i < len(s) && s[i] != 'e' && s[i] != 'E'; i++ {
			if s[i] < '0' || s[i] > '9' {
				c.wrErr("character", "invalid")
				return 0
			}
			hasDigits = true
			decimalPlaces++
			result = result*10 + float64(s[i]-'0')
		}
	}

	// Apply decimal places
	if hasDecimal {
		for j := 0; j < decimalPlaces; j++ {
			result /= 10
		}
	}

	if !hasDigits {
		c.wrErr("format", "invalid")
		return 0
	}

	// Parse scientific notation exponent if present
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++ // skip 'e'/'E'
		if i >= len(s) {
			c.wrErr("format", "invalid")
			return 0
		}
		expNeg := false
		if s[i] == '+' {
			i++
		} else if s[i] == '-' {
			expNeg = true
			i++
		}
		if i >= len(s) {
			c.wrErr("format", "invalid")
			return 0
		}
		var exp int
		for ; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				c.wrErr("character", "invalid")
				return 0
			}
			// Cap exponent and prevent overflow during parsing
			if exp < 1000 { // Still allow parsing but cap later
				exp = exp*10 + int(s[i]-'0')
			}
		}

		// Cap exponent to prevent DOS and handle infinity correctly
		// float64 max is ~1.8e308, min positive is ~5e-324
		if exp > 400 {
			exp = 400
		}

		// Apply exponent
		if result != 0 {
			mult := 1.0
			for j := 0; j < exp; j++ {
				mult *= 10
			}
			if expNeg {
				result /= mult
			} else {
				result *= mult
			}
		}
	}

	if negative {
		result = -result
	}

	return result
}

// wrFloat32 writes a float32 to the buffer destination.
func (c *Conv) wrFloat32(dest BuffDest, val float32) {
	c.wrFloatBase(dest, float64(val), 3.4028235e+38)
}

// wrFloat64 writes a float64 to the buffer destination.
func (c *Conv) wrFloat64(dest BuffDest, val float64) {
	c.wrFloatBase(dest, float64(val), 1.7976931348623157e+308)
}

// wrFloatBase contains the shared logic for writing float values.
func (c *Conv) wrFloatBase(dest BuffDest, val float64, maxInf float64) {
	// Handle special cases
	if val != val { // NaN
		c.WrString(dest, "NaN")
		return
	}
	if val == 0 {
		c.WrString(dest, "0")
		return
	}

	// Handle infinity
	if val > maxInf {
		c.WrString(dest, "+Inf")
		return
	}
	if val < -maxInf {
		c.WrString(dest, "-Inf")
		return
	}

	// Handle negative numbers
	negative := val < 0
	if negative {
		c.WrString(dest, "-")
		val = -val
	}

	// Check if it's effectively an integer
	if val < 1e15 && val == float64(int64(val)) {
		c.wrIntBase(dest, int64(val), 10, false)
		return
	}

	// For numbers with decimal places, use a precision-limited approach
	// Round to 6 decimal places to avoid precision issues
	scaled := val * 1000000
	rounded := int64(scaled + 0.5)

	intPart := rounded / 1000000
	fracPart := rounded % 1000000

	// Write integer part
	c.wrIntBase(dest, intPart, 10, false)

	// Write fractional part if non-zero
	if fracPart > 0 {
		c.WrString(dest, ".")

		// Build fractional string using local array to avoid buffer conflicts
		var digits [6]byte
		temp := fracPart
		for i := 0; i < 6; i++ {
			digits[i] = byte(temp%10) + '0'
			temp /= 10
		}

		// Find the start position (skip leading zeros in the array)
		start := 0
		for start < 6 && digits[start] == '0' {
			start++
		}

		// Write digits in reverse order (correct order), skipping leading zeros
		if start < 6 {
			for i := 5; i >= start; i-- {
				c.wrByte(dest, digits[i])
			}
		}
	}
}

// WriteFloat writes a float64 as decimal text to the output buffer.
func (c *Conv) WriteFloat(v float64) {
	c.wrFloat64(BuffOut, v)
}
