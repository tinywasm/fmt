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
		case int:
			// Convert int to string and write - simple conversion for errors
			if v == 0 {
				c.WrString(BuffErr, "0")
			} else {
				// Simple int to string conversion for error messages
				var buf [20]byte // Enough for 64-bit int
				n := len(buf)
				negative := v < 0
				if negative {
					v = -v
				}
				for v > 0 {
					n--
					buf[n] = byte(v%10) + '0'
					v /= 10
				}
				if negative {
					n--
					buf[n] = '-'
				}
				c.WrString(BuffErr, string(buf[n:]))
			}
		default:
			// For other types, convert to string representation
			c.WrString(BuffErr, "<unsupported>")
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
