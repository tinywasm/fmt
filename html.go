package fmt

// EscapeAttr returns a string safe to place inside an HTML attribute value.
func (c *Conv) EscapeAttr() string {
	return c.Replace("&", "&amp;").
		Replace("\"", "&quot;").
		Replace("'", "&#39;").
		Replace("<", "&lt;").
		Replace(">", "&gt;").
		String()
}

// EscapeHTML returns a string safe for inclusion into HTML content.
func (c *Conv) EscapeHTML() string {
	return c.Replace("&", "&amp;").
		Replace("<", "&lt;").
		Replace(">", "&gt;").
		Replace("\"", "&quot;").
		Replace("'", "&#39;").
		String()
}

// Html creates a string for HTML content, similar to Translate but without automatic spacing.
// It supports two modes:
// 1. Format mode: If the first argument is a string containing '%', it behaves like Fmt.
// 2. Concatenation mode: Otherwise, it concatenates arguments (translating with hook) without spaces.
func Html(values ...any) *Conv {
	c := GetConv()
	if len(values) == 0 {
		return c
	}

	// PASO 1: Detección de formato
	if format, ok := values[0].(string); ok {
		// Simple check for % to detect format string
		hasFormat := false
		for i := 0; i < len(format)-1; i++ {
			if format[i] == '%' {
				if c.isValidWriteFormatChar(rune(format[i+1])) {
					hasFormat = true
					break
				}
			}
		}

		if hasFormat {
			// Use Fmt logic
			return c.wrFormat(BuffOut, format, values[1:]...)
		}
	}

	// PASO 2: Concatenación sin espacios
	for _, val := range values {
		switch v := val.(type) {
		case string:
			c.WrString(BuffOut, tr(v))
		default:
			c.AnyToBuff(BuffWork, v)
			if c.hasContent(BuffWork) {
				c.WrString(BuffOut, c.GetString(BuffWork))
				c.ResetBuffer(BuffWork)
			}
		}
	}

	return c
}
