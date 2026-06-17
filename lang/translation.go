package lang

import "github.com/tinywasm/fmt"

// Translate creates a translated string with support for multilingual translations.
// EN words are lookup keys (case-insensitive). Pass-through occurs if missing from dictionary.
func Translate(values ...any) *fmt.Conv {
	return SmartArgs(fmt.GetConv(), fmt.BuffOut, " ", true, false, values...)
}

// SmartArgs handles language detection, format string detection, and argument processing
func SmartArgs(c *fmt.Conv, dest fmt.BuffDest, separator string, allowStringCode bool, detectFormat bool, values ...any) *fmt.Conv {
	if len(values) == 0 {
		return c
	}

	// PASO 1: Detección de idioma
	currentLang, startIdx := detectLanguage(c, values, allowStringCode)

	// Adjust values based on startIdx
	args := values[startIdx:]
	if len(args) == 0 {
		return c
	}

	// PASO 2: Detección de formato
	if detectFormat {
		if format, ok := args[0].(string); ok {
			// Check if it's a format string
			if isFormatString(format) {
				// Use public API
				c.WrString(dest, fmt.Sprintf(format, args[1:]...))
				return c
			}
		}
	}

	// PASO 3: Procesamiento de argumentos traducidos
	processTranslatedArgs(c, dest, args, currentLang, 0, separator)
	return c
}

func isFormatString(format string) bool {
	for i := 0; i < len(format)-1; i++ {
		if format[i] == '%' {
			if format[i+1] != '%' {
				return true
			}
			i++
		}
	}
	return false
}

// detectLanguage determines the current language and start index from variadic arguments
func detectLanguage(c *fmt.Conv, args []any, allowStringCode bool) (lang, int) {
	if len(args) == 0 {
		return getCurrentLang(), 0
	}

	// Check if first argument is a language specifier
	if langVal, ok := args[0].(lang); ok {
		return langVal, 1 // Skip the language argument in processing
	}

	// If first argument is a string of length 2, treat as language code only if recognized
	if allowStringCode {
		if strVal, ok := args[0].(string); ok && len(strVal) == 2 {
			if l, ok := mapLangCode(strVal); ok {
				return l, 1
			}
		}
	}

	// No language specified, use default
	return getCurrentLang(), 0
}

// processTranslatedArgs processes arguments with language-aware translation
func processTranslatedArgs(c *fmt.Conv, dest fmt.BuffDest, args []any, currentLang lang, startIndex int, separator string) {
	for i := startIndex; i < len(args); i++ {
		arg := args[i]
		switch v := arg.(type) {
		case string:
			if translated, ok := lookupWord(v, currentLang); ok {
				c.WrString(dest, translated)
			} else {
				c.WrString(dest, v) // pass-through
			}
		default:
			// Use public API to convert other types
			c.WrString(dest, fmt.Convert(v).String())
		}

		// Add separator
		if i < len(args)-1 {
			if separator == " " {
				if shouldAddSpace(args, i) {
					c.WrString(dest, separator)
				}
			} else {
				c.WrString(dest, separator)
			}
		}
	}
}

// shouldAddSpace determina si se debe agregar espacio después del argumento actual
func shouldAddSpace(args []any, currentIndex int) bool {
	if currentIndex >= len(args)-1 {
		return false
	}

	if currentStr, ok := args[currentIndex].(string); ok {
		if len(currentStr) > 0 {
			lastChar := currentStr[len(currentStr)-1]
			// Use public IsWordSeparator
			if lastChar == '\n' || lastChar == ' ' || lastChar == '/' {
				return false
			}
		}
	}

	if nextStr, ok := args[currentIndex+1].(string); ok {
		// Use public IsWordSeparator
		return !fmt.IsWordSeparator(nextStr)
	}

	return true
}
