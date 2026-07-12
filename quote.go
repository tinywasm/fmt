package fmt

// Quote wraps a string in double quotes and escapes any special characters
// Example: Quote("hello \"world\"") returns "\"hello \\\"world\\\"\""
func (c *Conv) Quote() *Conv {
	if c.hasContent(BuffErr) {
		return c // Error chain interruption
	}
	if c.outLen == 0 {
		c.ResetBuffer(BuffOut)
		c.WrString(BuffOut, quoteStr)
		return c
	}

	// Use work buffer to build quoted string, then swap to output
	c.ResetBuffer(BuffWork)
	c.wrByte(BuffWork, '"')

	// Process buffer directly without string allocation (like capitalizeASCIIOptimized)
	for i := 0; i < c.outLen; i++ {
		char := c.out[i]
		switch char {
		case '"':
			c.wrByte(BuffWork, '\\')
			c.wrByte(BuffWork, '"')
		case '\\':
			c.wrByte(BuffWork, '\\')
			c.wrByte(BuffWork, '\\')
		case '\n':
			c.wrByte(BuffWork, '\\')
			c.wrByte(BuffWork, 'n')
		case '\r':
			c.wrByte(BuffWork, '\\')
			c.wrByte(BuffWork, 'r')
		case '\t':
			c.wrByte(BuffWork, '\\')
			c.wrByte(BuffWork, 't')
		default:
			c.wrByte(BuffWork, char)
		}
	}

	c.wrByte(BuffWork, '"')
	c.swapBuff(BuffWork, BuffOut)
	return c
}

// JSONEscape writes s to b with JSON string escaping (without surrounding quotes).
// Escapes: " → \", \ → \\, newline → \n, carriage return → \r, tab → \t,
// control chars (< 0x20) → \u00XX.
//
// The caller is responsible for writing the surrounding double quotes.
// This design allows the caller to compose JSON strings without extra allocations.
func JSONEscape(s string, b *Builder) {
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 0x20 && c != '"' && c != '\\' {
			continue // safe byte — part of the current bulk run
		}
		if i > start {
			b.WriteString(s[start:i])
		}
		switch c {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteString(`\u00`)
			_ = b.WriteByte("0123456789abcdef"[c>>4])
			_ = b.WriteByte("0123456789abcdef"[c&0xf])
		}
		start = i + 1
	}
	if start < len(s) {
		b.WriteString(s[start:])
	}
}
