package fmt

// Replace replaces up to n occurrences of old with new in the Conv content
// If n < 0, there is no limit on the number of replacements
// eg: "hello world" with old "world" and new "universe" will return "hello universe"
// Old and new can be any type, they will be converted to string using Convert
func (c *Conv) Replace(oldAny, newAny any, n ...int) *Conv {
	if c.hasContent(BuffErr) {
		return c // Error chain interruption
	}

	if c.outLen == 0 {
		return c // OPTIMIZED: Direct length check instead of GetString
	}

	// Preserve original state before temporary conversions
	originalDataPtr := c.dataPtr
	originalKind := c.kind

	// Use internal work buffer instead of GetConv() for zero-allocation
	c.ResetBuffer(BuffWork)                     // Clear work buffer
	c.AnyToBuff(BuffWork, oldAny)               // Convert oldAny to work buffer
	oldBytesTemp := c.GetBytes(BuffWork)        // Get old bytes from work buffer
	oldBytes := make([]byte, len(oldBytesTemp)) // Copy to prevent corruption
	copy(oldBytes, oldBytesTemp)
	old := string(oldBytes) // Only create string when needed for compatibility

	c.ResetBuffer(BuffWork)                     // Clear work buffer for next conversion
	c.AnyToBuff(BuffWork, newAny)               // Convert newAny to work buffer
	newBytesTemp := c.GetBytes(BuffWork)        // Get new bytes from work buffer
	newBytes := make([]byte, len(newBytesTemp)) // Copy to prevent corruption
	copy(newBytes, newBytesTemp)
	newStr := string(newBytes) // Only create string when needed for compatibility

	// Restore original state after temporary conversions
	c.dataPtr = originalDataPtr
	c.kind = originalKind

	// Check early return condition
	if len(old) == 0 {
		return c
	}

	// Estimate buffer capacity based on replacement patterns
	estimatedCap := c.outLen
	if len(newStr) > len(old) {
		// If new string is longer, estimate extra space needed
		estimatedCap += (len(newStr) - len(old)) * 5 // Assume up to 5 replacements
	}
	// Inline makeBuf logic
	bufCap := estimatedCap
	if bufCap < defaultBufCap {
		bufCap = defaultBufCap
	}
	out := make([]byte, 0, bufCap)
	// Default behavior: replace all occurrences (n = -1)
	maxReps := -1
	if len(n) > 0 {
		maxReps = n[0]
	}

	rep := 0
	// OPTIMIZED: Process buffer directly for ASCII content or use string fallback for Unicode
	isASCII := true
	for i := 0; i < c.outLen; i++ {
		if c.out[i] > 127 {
			isASCII = false
			break
		}
	}

	if isASCII && len(oldBytes) > 0 && oldBytes[0] <= 127 {
		// Fast path: ASCII-only content using direct byte comparison

		for i := 0; i < c.outLen; i++ {
			// Check for occurrence of old in the buffer
			if i+len(oldBytes) <= c.outLen && (maxReps < 0 || rep < maxReps) {
				match := true
				for j := 0; j < len(oldBytes); j++ {
					if c.out[i+j] != oldBytes[j] {
						match = false
						break
					}
				}
				if match {
					// Add the new bytes to the out
					out = append(out, newBytes...)
					// Skip the length of the old bytes in the original buffer
					i += len(oldBytes) - 1
					// Increment replacement counter
					rep++
					continue
				}
			}
			// Add the current byte to the out
			out = append(out, c.out[i])
		}
	} else {
		// Unicode fallback: use string processing
		str := c.GetString(BuffOut)
		for i := 0; i < len(str); i++ {
			// Check for occurrence of old in the string and if we haven't reached the maximum rep
			if i+len(old) <= len(str) && str[i:i+len(old)] == old && (maxReps < 0 || rep < maxReps) {
				// Add the new word to the out
				out = append(out, newStr...)
				// Skip the length of the old word in the original string
				i += len(old) - 1
				// Increment replacement counter
				rep++
			} else {
				// Add the current character to the out
				out = append(out, str[i])
			}
		}
	}

	// ✅ Update buffer using API instead of direct manipulation
	c.ResetBuffer(BuffOut)  // Clear buffer using API
	c.wrBytes(BuffOut, out) // Write using API
	return c
}

// TrimSuffix removes the specified suffix from the Conv content if it exists
// eg: "hello.txt" with suffix ".txt" will return "hello"
func (c *Conv) TrimSuffix(suffix string) *Conv {
	if c.hasContent(BuffErr) {
		return c // Error chain interruption
	}

	// OPTIMIZED: Direct length check and buffer processing
	if c.outLen < len(suffix) {
		return c
	}

	// Check if suffix matches (ASCII optimization)
	suffixBytes := []byte(suffix)
	match := true
	startPos := c.outLen - len(suffixBytes)
	for i := 0; i < len(suffixBytes); i++ {
		if c.out[startPos+i] != suffixBytes[i] {
			match = false
			break
		}
	}

	if !match {
		return c
	}

	// ✅ Update buffer using API instead of direct manipulation
	newLen := c.outLen - len(suffix)
	c.out = c.out[:newLen]
	c.outLen = newLen
	return c
}

// TrimPrefix removes the specified prefix from the Conv content if it exists
// eg: "prefix-hello" with prefix "prefix-" will return "hello"
func (c *Conv) TrimPrefix(prefix string) *Conv {
	if c.hasContent(BuffErr) {
		return c // Error chain interruption
	}

	// OPTIMIZED: Direct length check and buffer processing
	if c.outLen < len(prefix) {
		return c
	}

	// Check if prefix matches (ASCII optimization)
	prefixBytes := []byte(prefix)
	match := true
	for i := 0; i < len(prefixBytes); i++ {
		if c.out[i] != prefixBytes[i] {
			match = false
			break
		}
	}

	if !match {
		return c
	}

	// ✅ Update buffer using API - shift remaining content
	prefixLen := len(prefix)
	copy(c.out, c.out[prefixLen:c.outLen])
	c.outLen -= prefixLen
	c.out = c.out[:c.outLen]
	return c
}

// TrimSpace removes spaces at the beginning and end of the Conv content
// eg: "  hello world  " will return "hello world"
func (c *Conv) TrimSpace() *Conv {
	if c.hasContent(BuffErr) {
		return c // Error chain interruption
	}

	// OPTIMIZED: Direct buffer processing
	if c.outLen == 0 {
		return c
	}

	// Remove spaces at the beginning
	start := 0
	for start < c.outLen && (c.out[start] == ' ' || c.out[start] == '\t' || c.out[start] == '\n' || c.out[start] == '\r') {
		start++
	}

	// Remove spaces at the end
	end := c.outLen - 1
	for end >= 0 && (c.out[end] == ' ' || c.out[end] == '\t' || c.out[end] == '\n' || c.out[end] == '\r') {
		end--
	}

	// Special case: empty string (all whitespace)
	if start > end {
		// Clear buffer
		c.outLen = 0
		c.out = c.out[:0]
		// Also clear dataPtr to prevent fallback
		c.dataPtr = nil
		c.kind = K.String
		return c
	}

	// ✅ Update buffer using direct manipulation for efficiency
	newLen := end - start + 1
	if start > 0 {
		copy(c.out, c.out[start:end+1])
	}
	c.outLen = newLen
	c.out = c.out[:newLen]
	return c
}
