package fmt

// Index finds the first occurrence of substr in s, returns -1 if not found.
// This is the base primitive that other functions will reuse.
//
// Examples:
//
//	Index("hello world", "world")  // returns 6
//	Index("hello world", "lo")     // returns 3
//	Index("hello world", "xyz")    // returns -1 (not found)
//	Index("hello world", "")       // returns 0 (empty string)
//	Index("data\x00more", "\x00")  // returns 4 (null byte)
func Index(s, substr string) int {
	n := len(substr)
	if n == 0 {
		return 0 // Standard behavior: empty string is found at position 0
	}
	if n == 1 {
		// Optimized single byte search
		for i := 0; i < len(s); i++ {
			if s[i] == substr[0] {
				return i
			}
		}
		return -1
	}

	// Brute force for longer strings
	for i := 0; i <= len(s)-n; i++ {
		if s[i:i+n] == substr {
			return i
		}
	}
	return -1
}

// Count checks how many times the string 'search' is present in 'Conv'.
// Uses Index internally for consistency and maintainability.
//
// Examples:
//
//	Count("abracadabra", "abra")    // returns 2
//	Count("hello world", "l")       // returns 3
//	Count("golang", "go")           // returns 1
//	Count("test", "xyz")            // returns 0 (not found)
//	Count("anything", "")           // returns 0 (empty search)
//	Count("a\x00b\x00c", "\x00")    // returns 2 (null bytes)
func Count(Conv, search string) int {
	if len(search) == 0 {
		return 0
	}

	count := 0
	s := Conv
	for {
		i := Index(s, search)
		if i == -1 {
			break
		}
		count++
		s = s[i+len(search):] // Skip past this match
	}
	return count
}

// Contains checks if the string 'search' is present in 'Conv'.
// Uses Index internally for efficient single-pass detection.
//
// Examples:
//
//	Contains("hello world", "world")  // returns true
//	Contains("hello world", "xyz")    // returns false
//	Contains("", "test")              // returns false (empty string)
//	Contains("test", "")              // returns false (empty search)
//	Contains("data\x00more", "\x00")  // returns true (null byte)
//	Contains("Case", "case")          // returns false (case sensitive)
func Contains(Conv, search string) bool {
	if len(search) == 0 {
		return false // Cadena vacía no se considera contenida
	}
	return Index(Conv, search) != -1
}

// HasPrefix reports whether the string 'conv' begins with 'prefix'.
// Implemented using Index for consistency with other helpers in this package.
//
// Examples:
//
//	HasPrefix("hello", "he")   // returns true
//	HasPrefix("hello", "hello") // returns true
//	HasPrefix("hello", "")      // returns false (empty prefix)
//	HasPrefix("a", "abc")       // returns false (prefix longer than string)
func HasPrefix(conv, prefix string) bool {
	if len(prefix) == 0 {
		return true // Follow stdlib semantics: empty string is a prefix of any string
	}
	if len(conv) < len(prefix) {
		return false
	}
	// Search only in the slice that should match the prefix; Index must return 0.
	return Index(conv[:len(prefix)], prefix) == 0
}

// HasSuffix reports whether the string 'Conv' ends with 'suffix'.
// Implemented using Index by checking the tail of the string.
//
// Examples:
//
//	HasSuffix("testing", "ing")  // returns true
//	HasSuffix("file.txt", ".txt") // returns true
//	HasSuffix("hello", "")        // returns false (empty suffix)
//	HasSuffix("go", "golang")     // returns false
func HasSuffix(conv, suffix string) bool {
	if len(suffix) == 0 {
		return true // Follow stdlib semantics: empty string is a suffix of any string
	}
	if len(conv) < len(suffix) {
		return false
	}
	// The suffix should be at the start of the tail slice
	return Index(conv[len(conv)-len(suffix):], suffix) == 0
}

// Matches reports whether content contains ALL given terms (AND semantics).
// content and each term are lowercased before comparison.
// Returns false if no terms are given or any term is empty.
func Matches(content string, terms ...string) bool {
	return matches(false, content, terms...)
}

// MatchesAny reports whether content contains AT LEAST ONE of the given terms (OR semantics).
// content and each term are lowercased before comparison.
// Returns false if no terms are given or any term is empty.
func MatchesAny(content string, terms ...string) bool {
	return matches(true, content, terms...)
}

func matches(or bool, content string, terms ...string) bool {
	if len(terms) == 0 {
		return false
	}
	conv := GetConv()
	defer conv.PutConv()

	lower := conv.Reset().Write(content).ToLower().GetString(BuffOut)
	for _, term := range terms {
		if term == "" {
			return false
		}
		termLower := conv.Reset().Write(term).ToLower().GetString(BuffOut)
		found := Contains(lower, termLower)
		if or && found {
			return true // OR: basta con uno
		}
		if !or && !found {
			return false // AND: falla en el primero ausente
		}
	}
	return !or // AND sin fallos → true; OR sin éxito → false
}
