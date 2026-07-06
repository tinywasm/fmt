package fmt

// =============================================================================
// FORMAT PRECISION OPERATIONS - Decimal rounding and precision control
// =============================================================================

// Round rounds or truncates the current numeric value to the specified number of decimal places.
//
// - If the optional 'down' parameter is omitted or false, it applies "round half to even" (bankers rounding, like Go: 2.5 → 2, 3.5 → 4).
// - If 'down' is true, it truncates (floors) the value without rounding.
//
// Example:
//
//	Convert("3.14159").Round(2)        // "3.14" (rounded)
//	Convert("3.145").Round(2)          // "3.14" (rounded)
//	Convert("3.155").Round(2)          // "3.16" (rounded)
//	Convert("3.14159").Round(2, true)  // "3.14" (truncated)
//
// If the value is not numeric, returns "0" with the requested number of decimals.
func (t *Conv) Round(decimals int, down ...bool) *Conv {
	if t.hasContent(BuffErr) {
		return t
	}
	roundDown := false
	if len(down) > 0 && down[0] {
		roundDown = true
	}
	t.applyRoundingToNumber(BuffOut, decimals, roundDown)
	// If the result is not numeric, set to zero with correct decimals
	str := t.GetString(BuffOut)
	if !t.isNumericString(str) || str == "" || str == "-" {
		t.ResetBuffer(BuffOut)
		t.WrString(BuffOut, "0")
		if decimals > 0 {
			t.WrString(BuffOut, ".")
			for i := 0; i < decimals; i++ {
				t.WrString(BuffOut, "0")
			}
		}
	}
	return t
}

// applyRoundingToNumber rounds the current number to specified decimal places
// Universal method with dest-first parameter order - follows buffer API architecture
func (t *Conv) applyRoundingToNumber(dest BuffDest, decimals int, roundDown bool) *Conv {
	if t.hasContent(BuffErr) {
		return t
	}

	// Get current string representation
	currentStr := t.GetString(dest)

	// Find decimal point
	dotIndex := func() int {
		for i := range len(currentStr) {
			if currentStr[i] == '.' {
				return i
			}
		}
		return -1
	}()

	// If no decimal point, add zeros if needed
	if dotIndex == -1 {
		if decimals > 0 {
			t.WrString(dest, ".")
			for i := 0; i < decimals; i++ {
				t.wrByte(dest, '0')
			}
		}
		return t
	}

	// Calculate required length
	var targetLen int
	if decimals == 0 {
		targetLen = dotIndex // No decimal point for 0 decimals
	} else {
		targetLen = dotIndex + 1 + decimals // Include decimal point and decimal places
	}

	// If we need to truncate or round
	if len(currentStr) > targetLen {
		if roundDown {
			// Simple truncation for roundDown (floor behavior)
			t.ResetBuffer(dest)
			t.WrString(dest, currentStr[:targetLen])
		} else {
			// Implement Go's round half to even (bankers rounding)
			var firstDiscarded byte = '0'
			var moreNonZero bool
			var lastKeptIdx int
			if decimals == 0 {
				// For decimals==0, first discarded is the first digit after the dot
				if dotIndex+1 < len(currentStr) {
					firstDiscarded = currentStr[dotIndex+1]
				}
				for i := dotIndex + 2; i < len(currentStr); i++ {
					if currentStr[i] != '0' && currentStr[i] != '.' {
						moreNonZero = true
						break
					}
				}
				lastKeptIdx = dotIndex - 1
			} else {
				if targetLen < len(currentStr) {
					firstDiscarded = currentStr[targetLen]
				}
				for i := targetLen + 1; i < len(currentStr); i++ {
					if currentStr[i] != '0' && currentStr[i] != '.' {
						moreNonZero = true
						break
					}
				}
				lastKeptIdx = targetLen - 1
			}
			shouldRoundUp := false
			if firstDiscarded > '5' {
				shouldRoundUp = true
			} else if firstDiscarded < '5' {
				shouldRoundUp = false
			} else if firstDiscarded == '5' {
				if moreNonZero {
					shouldRoundUp = true
				} else {
					// Check if the last kept digit is odd (for even rounding)
					for lastKeptIdx >= 0 && currentStr[lastKeptIdx] == '.' {
						lastKeptIdx--
					}
					if lastKeptIdx >= 0 && (currentStr[lastKeptIdx]-'0')%2 == 1 {
						shouldRoundUp = true
					} else {
						shouldRoundUp = false
					}
				}
			}

			if shouldRoundUp {
				// Rounding up
				var roundedBytes []byte
				if decimals == 0 {
					roundedBytes = []byte(currentStr[:dotIndex])
				} else {
					roundedBytes = []byte(currentStr[:targetLen])
				}
				carry := 1
				for i := len(roundedBytes) - 1; i >= 0 && carry > 0; i-- {
					if roundedBytes[i] == '.' {
						continue
					}
					if roundedBytes[i] >= '0' && roundedBytes[i] <= '9' {
						digit := int(roundedBytes[i]-'0') + carry
						if digit > 9 {
							roundedBytes[i] = '0'
							carry = 1
						} else {
							roundedBytes[i] = byte(digit) + '0'
							carry = 0
						}
					}
				}
				t.ResetBuffer(dest)
				if carry > 0 {
					t.WrString(dest, "1")
				}
				t.wrBytes(dest, roundedBytes)
			} else {
				// Truncation (no rounding up)
				t.ResetBuffer(dest)
				if decimals == 0 {
					t.WrString(dest, currentStr[:dotIndex])
				} else {
					t.WrString(dest, currentStr[:targetLen])
				}
			}
		}
	} else if len(currentStr) < targetLen {
		// Add trailing zeros
		zerosNeeded := targetLen - len(currentStr)
		for i := 0; i < zerosNeeded; i++ {
			t.wrByte(dest, '0')
		}
	}

	return t
}

// wrFloatWithPrecision formats a float with specified precision and writes to buffer destination
// Universal method with dest-first parameter order - follows buffer API architecture
func (c *Conv) wrFloatWithPrecision(dest BuffDest, value float64, precision int) {
	// Handle special cases
	if value != value { // NaN
		c.WrString(dest, "NaN")
		return
	}

	if value == 0 {
		if precision > 0 {
			c.WrString(dest, "0.")
			for i := 0; i < precision; i++ {
				c.wrByte(dest, '0')
			}
		} else {
			c.WrString(dest, "0")
		}
		return
	}

	// Handle infinity
	if value > 1.7976931348623157e+308 {
		c.WrString(dest, "+Inf")
		return
	}
	if value < -1.7976931348623157e+308 {
		c.WrString(dest, "-Inf")
		return
	}

	// Handle negative numbers
	negative := value < 0
	if negative {
		c.WrString(dest, "-")
		value = -value
	}

	// Scale value by precision to get required decimal places
	multiplier := 1.0
	for i := 0; i < precision; i++ {
		multiplier *= 10
	}

	scaled := value * multiplier
	rounded := int64(scaled + 0.5) // Round to nearest integer

	// Extract integer and fractional parts
	intPart := rounded
	for i := 0; i < precision; i++ {
		intPart /= 10
	}

	fracPart := rounded - intPart*int64(multiplier)

	// Write integer part
	c.WrIntBase(dest, intPart, 10, true)

	// Write fractional part if precision > 0
	if precision > 0 {
		c.WrString(dest, ".")

		// Convert fractional part to string with leading zeros
		// Build digits array in reverse order to avoid allocations
		var digits [20]byte // Support up to 20 decimal places
		temp := fracPart
		for i := 0; i < precision; i++ {
			digits[i] = byte(temp%10) + '0'
			temp /= 10
		}

		// Write digits in reverse order (correct order)
		for i := precision - 1; i >= 0; i-- {
			c.wrByte(dest, digits[i])
		}
	}
}
