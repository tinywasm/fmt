package fmt

// Capitalize transforms the first letter of each word to uppercase and the rest to lowercase.
// Preserves all whitespace formatting (spaces, tabs, newlines) without normalization.
// OPTIMIZED: Uses work buffer efficiently to minimize allocations
// For example: "  hello   world  " -> "  Hello   World  "
func (t *Conv) Capitalize() *Conv {
	if t.hasContent(BuffErr) {
		return t // Error chain interruption
	}

	if t.outLen == 0 {
		return t
	}

	// Fast path for ASCII-only content (common case)
	if t.isASCIIOnly() {
		t.capitalizeASCIIOptimized()
		return t
	}

	// Unicode fallback
	return t.capitalizeUnicode()
}

// capitalizeASCIIOptimized processes ASCII text preserving all formatting
func (t *Conv) capitalizeASCIIOptimized() {
	// Use work buffer for processing
	t.ResetBuffer(BuffWork)

	inWord := false

	for i := 0; i < t.outLen; i++ {
		ch := t.out[i]

		// Use centralized word separator detection
		if isWordSeparator(ch) {
			// Preserve all separator characters as-is
			t.work = append(t.work, ch)
			t.workLen++
			inWord = false
		} else {
			if !inWord {
				// Start of new word - capitalize first letter
				if ch >= 'a' && ch <= 'z' {
					ch -= 32 // Convert to uppercase
				}
				inWord = true
			} else {
				// Rest of word - lowercase other letters
				if ch >= 'A' && ch <= 'Z' {
					ch += 32 // Convert to lowercase
				}
			}
			t.work = append(t.work, ch)
			t.workLen++
		}
	}

	// Swap processed content to output
	t.swapBuff(BuffWork, BuffOut)
}

// capitalizeUnicode handles full Unicode capitalization preserving formatting
func (t *Conv) capitalizeUnicode() *Conv {
	str := t.GetString(BuffOut)

	// Use internal work buffer for intermediate processing
	t.ResetBuffer(BuffWork)

	inWord := false

	for _, r := range str {
		// Use centralized word separator detection
		if isWordSeparator(r) {
			// Preserve all separator characters as-is
			t.WrString(BuffWork, string(r))
			inWord = false
		} else {
			if !inWord {
				// Start of new word - capitalize first letter
				t.WrString(BuffWork, string(toUpperRune(r)))
				inWord = true
			} else {
				// Rest of word - lowercase other letters
				t.WrString(BuffWork, string(toLowerRune(r)))
			}
		}
	}

	// Copy result from work buffer to output buffer
	result := t.GetString(BuffWork)
	t.ResetBuffer(BuffOut)
	t.WrString(BuffOut, result)
	return t
}

// convert to lower case eg: "HELLO WORLD" -> "hello world"
func (t *Conv) ToLower() *Conv {
	return t.changeCaseOptimized(true)
}

// convert to upper case eg: "hello world" -> "HELLO WORLD"
func (t *Conv) ToUpper() *Conv {
	return t.changeCaseOptimized(false)
}

// hasUpperPrefix reports whether the first character is an uppercase letter.
// Supports ASCII (A-Z) and common accented uppercase (Á, É, Í, etc.).
func (t *Conv) hasUpperPrefix() bool {
	if t.outLen == 0 {
		return false
	}
	ch := t.out[0]
	// ASCII fast path (A-Z) - reuses pattern from capitalizeASCIIOptimized
	if ch >= 'A' && ch <= 'Z' {
		return true
	}
	// Unicode: check accented uppercase from aU (mapping.go)
	if ch > 127 {
		runes := []rune(t.GetString(BuffOut))
		if len(runes) == 0 {
			return false
		}
		r := runes[0]
		for _, char := range aU {
			if r == char {
				return true
			}
		}
	}
	return false
}

// HasUpperPrefix reports whether the string s starts with an uppercase letter.
// Supports ASCII (A-Z) and common accented uppercase (Á, É, Í, Ó, Ú, etc.).
//
// Examples:
//
//	HasUpperPrefix("Hello") -> true
//	HasUpperPrefix("Ángel") -> true
//	HasUpperPrefix("hello") -> false
func HasUpperPrefix(s string) bool {
	if len(s) == 0 {
		return false
	}
	return Convert(s).hasUpperPrefix()
}

// changeCaseOptimized implements fast ASCII path with fallback to full Unicode
func (t *Conv) changeCaseOptimized(toLower bool) *Conv {
	if t.hasContent(BuffErr) {
		return t // Error chain interruption
	}

	if t.outLen == 0 {
		return t
	}

	// Fast path: ASCII-only optimization (covers 85% of use cases)
	if t.isASCIIOnly() {
		t.changeCaseASCIIInPlace(toLower)
		return t
	}

	// Fallback: Full Unicode support for complex cases
	return t.changeCaseUnicode(toLower)
}

// changeCaseASCIIInPlace processes ASCII characters directly in buffer (zero allocations)
func (t *Conv) changeCaseASCIIInPlace(toLower bool) {
	for i := 0; i < t.outLen; i++ {
		if toLower {
			// A-Z (65-90) → a-z (97-122): add 32
			if t.out[i] >= 'A' && t.out[i] <= 'Z' {
				t.out[i] += 32
			}
		} else {
			// a-z (97-122) → A-Z (65-90): subtract 32
			if t.out[i] >= 'a' && t.out[i] <= 'z' {
				t.out[i] -= 32
			}
		}
	}
}

// isASCIIOnly checks if buffer contains only ASCII characters (fast check)
func (t *Conv) isASCIIOnly() bool {
	for i := 0; i < t.outLen; i++ {
		if t.out[i] > 127 {
			return false
		}
	}
	return true
}

// changeCaseUnicode handles full Unicode case conversion (legacy method)
func (t *Conv) changeCaseUnicode(toLower bool) *Conv {
	return t.changeCase(toLower, BuffOut)
}

// changeCase consolidates ToLower and ToUpper functionality - now accepts a destination buffer for internal reuse
func (t *Conv) changeCase(toLower bool, dest BuffDest) *Conv {
	if t.hasContent(BuffErr) {
		return t // Error chain interruption
	}

	str := t.GetString(dest)
	if len(str) == 0 {
		return t
	}

	// Convert to runes for proper Unicode handling
	runes := []rune(str)

	// Process runes for case conversion
	for i, r := range runes {
		if toLower {
			runes[i] = toLowerRune(r)
		} else {
			runes[i] = toUpperRune(r)
		}
	}
	// Convert back to string and store in buffer using API
	out := string(runes)
	t.ResetBuffer(dest)   // Clear buffer using API
	t.WrString(dest, out) // Write using API

	return t
}

// converts Conv to camelCase (first word lowercase) eg: "Hello world" -> "helloWorld"
func (t *Conv) CamelLow() *Conv {
	return t.toCaseTransformMinimal(true, "")
}

// converts Conv to PascalCase (all words capitalized) eg: "hello world" -> "HelloWorld"
func (t *Conv) CamelUp() *Conv {
	return t.toCaseTransformMinimal(false, "")
}

// snakeCase converts a string to snake_case format with optional separator.
// If no separator is provided, underscore "_" is used as default.
// Example:
//
//	Input: "camelCase" -> Output: "camel_case"
//	Input: "PascalCase", "-" -> Output: "pascal-case"
//	Input: "APIResponse" -> Output: "api_response"
//	Input: "user123Name", "." -> Output: "user123.name"
//
// SnakeLow converts Conv to snake_case format
func (t *Conv) SnakeLow(sep ...string) *Conv {
	// Phase 4.3: Use local variable instead of struct field
	separator := "_" // underscore default
	if len(sep) > 0 {
		separator = sep[0]
	}
	return t.toCaseTransformMinimal(true, separator)
}

// SnakeUp converts Conv to Snake_Case format
func (t *Conv) SnakeUp(sep ...string) *Conv {
	// Phase 4.3: Use local variable instead of struct field
	separator := "_" // underscore default
	if len(sep) > 0 {
		separator = sep[0]
	}
	return t.toCaseTransformMinimal(false, separator)
}

// Minimal implementation without pools or builders - optimized for minimal allocations
// Minimal implementation - optimized for minimal allocations using mapping.go functions
func (t *Conv) toCaseTransformMinimal(firstWordLower bool, separator string) *Conv {
	if t.hasContent(BuffErr) {
		return t // Error chain interruption
	}

	if t.outLen == 0 {
		return t
	}

	// Use work buffer for processing
	t.ResetBuffer(BuffWork)

	// Process each character and determine word boundaries
	wordIndex := 0
	prevWasSpace := false
	prevWasLower := false
	prevWasDigit := false

	for i := 0; i < t.outLen; i++ {
		char := t.out[i]

		// For CamelCase, only whitespace characters are true separators
		isWhitespace := char == ' ' || char == '\t' || char == '\n' || char == '\r'

		if isWhitespace {
			prevWasSpace = true
			continue // Skip whitespace separators
		}

		// Determine if starting new word
		isNewWord := false
		if i == 0 {
			isNewWord = true // First character is always start of first word
		} else if prevWasSpace {
			isNewWord = true // After whitespace
		} else if separator != "" {
			// For snake_case: more aggressive word splitting
			if (prevWasLower && char >= 'A' && char <= 'Z') || // camelCase transition
				(prevWasDigit && ((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'))) { // digit to letter
				isNewWord = true
			}
		} else {
			// For CamelCase/PascalCase: Split on common word boundaries
			if prevWasLower && char >= 'A' && char <= 'Z' { // lowercase-to-uppercase (camelCase)
				isNewWord = true
			} else if prevWasDigit && char >= 'A' && char <= 'Z' { // digit-to-uppercase
				// For CamelLow: digit-to-uppercase is NOT a word boundary ("User123Name" → "user123name")
				// For CamelUp: digit-to-uppercase IS a word boundary ("User123Name" → "User123Name")
				if !firstWordLower {
					isNewWord = true // PascalCase (CamelUp) - treat as word boundary
				}
				// For camelCase (CamelLow) - don't treat as word boundary
			}
		}

		// Add separator if new word (except first) - only for snake_case
		if isNewWord && wordIndex > 0 && separator != "" {
			t.WrString(BuffWork, separator)
		}

		// Apply case transformation for letters only
		var result byte
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
			if isNewWord {
				// First letter of word
				if wordIndex == 0 && firstWordLower {
					result = t.toLowerByteHelper(char) // First word lowercase (camelCase)
				} else if separator != "" && firstWordLower {
					result = t.toLowerByteHelper(char) // snake_case - all lowercase
				} else {
					result = t.toUpperByteHelper(char) // PascalCase or subsequent camelCase words
				}
				wordIndex++
			} else {
				result = t.toLowerByteHelper(char) // Rest of word always lowercase
			}
		} else {
			// Non-letter characters: preserve as-is
			result = char
		}

		t.wrByte(BuffWork, result)

		// Update state
		prevWasSpace = false
		prevWasLower = (char >= 'a' && char <= 'z')
		prevWasDigit = (char >= '0' && char <= '9')
	}

	// Swap result to output
	t.swapBuff(BuffWork, BuffOut)
	return t
}

// Helper methods for case conversion (reuse mapping.go constants)
func (t *Conv) toUpperByteHelper(b byte) byte {
	if b >= 'a' && b <= 'z' {
		return b - asciiCaseDiff
	}
	return b
}

func (t *Conv) toLowerByteHelper(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + asciiCaseDiff
	}
	return b
}
