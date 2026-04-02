package fmt

// Custom error messages to avoid importing standard library packages like "errors" or "fmt"
// This keeps the binary size minimal for embedded systems and WebAssembly

// Err creates a new error message with support for multilingual translations
// Supports LocStr types for translations and lang types for language specification
// eg:
// fmt.Err("invalid format") returns "invalid format"
// fmt.Err(D.Format, D.Invalid) returns "invalid format"
// fmt.Err(ES,D.Format, D.Invalid) returns "formato inválido"

func Err(msgs ...any) *Conv {
	// UNIFIED PROCESSING: Use same intermediate function as Translate() but write to BuffErr
	return GetConv().SmartArgs(BuffErr, " ", true, false, msgs...)
}

// Errf creates a new Conv instance with error formatting similar to fmt.Errf
// Example: fmt.Errf("invalid value: %s", value).Error()
func Errf(format string, args ...any) *Conv {
	return GetConv().wrFormat(BuffErr, getCurrentLang(), format, args...)
}

// StringErr returns the content of the Conv along with any error and auto-releases to pool
func (c *Conv) StringErr() (out string, err error) {
	// If there's an error, return empty string and the error object (do NOT release to pool)
	if c.hasContent(BuffErr) {
		return "", c
	}

	// Otherwise return the string content and no error (safe to release to pool)
	out = c.GetString(BuffOut)
	c.putConv()
	return out, nil
}

// wrErr writes error messages with support for int, string and LocStr
// ENHANCED: Now supports int, string and LocStr parameters
// Used internally by AnyToBuff for type error messages
func (c *Conv) wrErr(msgs ...any) *Conv {
	// Write messages using default language (no detection needed)
	for i, msg := range msgs {
		if i > 0 {
			// Add space between words
			c.WrString(BuffErr, " ")
		}
		// fmt.Printf("wrErr: Processing message part: %v\n", msg) // Depuración

		switch v := msg.(type) {
		case string:
			if translated, ok := lookupWord(v, getCurrentLang()); ok {
				c.WrString(BuffErr, translated)
			} else {
				c.WrString(BuffErr, v)
			}
		case int, int8, int16, int32, int64:
			c.ResetBuffer(BuffWork)
			var val int64
			switch i := v.(type) {
			case int: val = int64(i)
			case int8: val = int64(i)
			case int16: val = int64(i)
			case int32: val = int64(i)
			case int64: val = i
			}
			c.wrIntBase(BuffWork, val, 10, true, false)
			c.WrString(BuffErr, c.GetString(BuffWork))
		case uint, uint8, uint16, uint32, uint64:
			c.ResetBuffer(BuffWork)
			var val uint64
			switch i := v.(type) {
			case uint: val = uint64(i)
			case uint8: val = uint64(i)
			case uint16: val = uint64(i)
			case uint32: val = uint64(i)
			case uint64: val = i
			}
			c.wrUintBase(BuffWork, val, 10)
			c.WrString(BuffErr, c.GetString(BuffWork))
		case bool:
			if v {
				c.WrString(BuffErr, "true")
			} else {
				c.WrString(BuffErr, "false")
			}
		case error:
			c.WrString(BuffErr, v.Error())
		default:
			// For other types, try AnyToBuff with BuffWork to avoid recursion or side effects
			c.ResetBuffer(BuffWork)
			c.AnyToBuff(BuffWork, v)
			if c.hasContent(BuffWork) {
				c.WrString(BuffErr, c.GetString(BuffWork))
			} else {
				c.WrString(BuffErr, "<unsupported>")
			}
		}
	}
	// fmt.Printf("wrErr: Final error buffer content: %q, errLen: %d\n", c.GetString(BuffErr), c.errLen) // Depuración
	return c
}

func (c *Conv) getError() string {
	if !c.hasContent(BuffErr) { // ✅ Use API method instead of len(c.err)
		return ""
	}
	return c.GetString(BuffErr) // ✅ Use API method instead of direct string(c.err)
}

func (c *Conv) Error() string {
	return c.getError()
}
