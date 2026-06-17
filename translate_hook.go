package fmt

// translateWord is the global hook for word translation.
// It is nil by default, meaning no translation is performed.
var translateWord func(word string) (string, bool)

// SetTranslator installs the global translator.
// It is typically called by the fmt/lang package during its initialization.
func SetTranslator(fn func(word string) (string, bool)) {
	translateWord = fn
}

// tr is an internal helper that translates a word using the global hook.
// If the hook is nil or the word is not found, it returns the original word.
func tr(word string) string {
	if translateWord != nil {
		if t, ok := translateWord(word); ok {
			return t
		}
	}
	return word
}
