package fmt

func (c *Conv) parseIntString(s string, base int, signed bool) int64 {
	// Handle decimal point for float-like input (e.g., "3.14")
	for i := 0; i < len(s); i++ {
		if s[i] == '.' {
			// Try to parse as float, then truncate
			// Use BuffWork to avoid conflicts with BuffOut
			c.ResetBuffer(BuffWork)
			c.WrString(BuffWork, s)

			// Swap buffers temporarily to use parseFloatBase
			c.swapBuff(BuffOut, BuffErr)  // Save current BuffOut to BuffErr
			c.swapBuff(BuffWork, BuffOut) // Move string to BuffOut for parsing

			f := c.parseFloatBase()
			hasError := c.hasContent(BuffErr)

			// Restore original BuffOut
			c.swapBuff(BuffOut, BuffWork) // Move parsed content to BuffWork
			c.swapBuff(BuffErr, BuffOut)  // Restore original BuffOut

			if hasError {
				return 0
			}
			return int64(f)
		}
	}
	if base < 2 || base > 36 {
		c.wrErr("Base", "invalid")
		return 0
	}
	var neg bool
	i := 0
	if len(s) > 0 && s[0] == '-' {
		if !signed {
			c.wrErr("number", "negative", "not", "allowed")
			return 0
		}
		neg = true
		i = 1
		if len(s) == 1 {
			c.wrErr("format", "invalid")
			return 0
		}
	} else if len(s) > 0 && s[0] == '+' {
		i = 1
		if len(s) == 1 {
			c.wrErr("format", "invalid")
			return 0
		}
	}
	var n int64
	for ; i < len(s); i++ {
		ch := s[i]
		var v byte
		switch {
		case '0' <= ch && ch <= '9':
			v = ch - '0'
		case 'a' <= ch && ch <= 'z':
			v = ch - 'a' + 10
		case 'A' <= ch && ch <= 'Z':
			v = ch - 'A' + 10
		default:
			c.wrErr("format", "invalid")
			return 0
		}
		if int(v) >= base {
			c.wrErr("format", "invalid")
			return 0
		}
		n = n*int64(base) + int64(v)
	}
	if neg {
		n = -n
	}
	return n
}

// Int converts the value to an integer with optional base specification.
// If no base is provided, base 10 is used. Supports bases 2-36.
// Returns the converted integer and any error that occurred during conversion.
func (c *Conv) Int(base ...int) (int, error) {
	val := c.parseIntBase(base...)
	if val < -2147483648 || val > 2147483647 {
		return 0, c.wrErr("number", "overflow")
	}
	if c.hasContent(BuffErr) {
		return 0, c
	}
	return int(val), nil
}

// getInt32 extrae el valor del buffer de salida y lo convierte a int32.
// Int32 extrae el valor del buffer de salida y lo convierte a int32.
func (c *Conv) Int32(base ...int) (int32, error) {
	val := c.parseIntBase(base...)
	if val < -2147483648 || val > 2147483647 {
		return 0, c.wrErr("number", "overflow")
	}
	if c.hasContent(BuffErr) {
		return 0, c
	}
	return int32(val), nil
}

// getInt64 extrae el valor del buffer de salida y lo convierte a int64.
// Int64 extrae el valor del buffer de salida y lo convierte a int64.
func (c *Conv) Int64(base ...int) (int64, error) {
	val := c.parseIntBase(base...)
	if c.hasContent(BuffErr) {
		return 0, c
	}
	return val, nil
}

// toInt64 converts various integer types to int64
func (c *Conv) toInt64(arg any) (int64, bool) {
	switch v := arg.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		return int64(v), true
	default:
		// Try reflection for custom types (e.g., type customInt int)
		return c.toInt64Reflect(arg)
	}
}

// wrIntBase writes an integer in the given base to the buffer, with optional uppercase digits
func (c *Conv) wrIntBase(dest BuffDest, val int64, base int, signed bool, upper ...bool) {
	if base < 2 || base > 36 {
		c.wrErr("Base", "invalid")
		return
	}
	if val == 0 {
		c.WrString(dest, "0")
		return
	}
	negative := signed && val < 0
	uval := val
	if negative {
		uval = -val
	}
	useUpper := false
	if len(upper) > 0 && upper[0] {
		useUpper = true
	}
	var digits string
	if useUpper {
		digits = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	} else {
		digits = "0123456789abcdef"
	}
	var out [64]byte
	idx := len(out)
	for uval > 0 {
		idx--
		out[idx] = digits[uval%int64(base)]
		uval /= int64(base)
	}
	if negative {
		idx--
		out[idx] = '-'
	}
	c.wrBytes(dest, out[idx:])
}

// parseIntBase reutiliza la lógica de conversión de string a int64, soportando signo y base, y reporta error usando la API interna.
// parseIntBase auto-detects signed/unsigned mode using c.Kind and parses the string accordingly.
// It does not take a signed parameter; instead, it checks c.Kind (K.Int = signed, K.Uint = unsigned).
func (c *Conv) parseIntBase(base ...int) int64 {

	s := c.GetString(BuffOut)
	baseVal := 10
	if len(base) > 0 {
		baseVal = base[0]
	}
	isSigned := c.kind == K.Int
	// Solo permitir negativos en base 10
	if len(s) > 0 && s[0] == '-' {
		if baseVal == 10 {
			isSigned = true
		} else {
			isSigned = false
		}
	}
	return c.parseIntString(s, baseVal, isSigned)
}
