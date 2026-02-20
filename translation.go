package fmt

// Translate creates a translated string with support for multilingual translations.
// EN words are lookup keys (case-insensitive). Pass-through occurs if missing from dictionary.
//
// Usage examples (prefer Noun + Adjective order for better Spanish grammar):
// Translate("format", "invalid")    // English: "Format Invalid", Spanish: "Formato Inválido"
// Translate(ES, "format", "invalid") // Spanish force: "Formato Inválido"
//
// MEMORY MANAGEMENT:
// The returned *Conv object is pooled.
// - Calling .String() or .Apply() automatically returns it to the pool.
// - If you use .Bytes() or other methods, you MUST call .PutConv() manually to avoid memory leaks.
func Translate(values ...any) *Conv {
	// UNIFIED PROCESSING: Use shared intermediate function
	return GetConv().SmartArgs(BuffOut, " ", true, false, values...)
}

// SmartArgs handles language detection, format string detection, and argument processing
// This unifies logic between Html (detectFormat=true, allowStringCode=false) and Translate (detectFormat=false, allowStringCode=true)
func (c *Conv) SmartArgs(dest BuffDest, separator string, allowStringCode bool, detectFormat bool, values ...any) *Conv {
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

	// PASO 2: Detección de formato (Opcional, usado por Html)
	if detectFormat {
		if format, ok := args[0].(string); ok {
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
				fmtArgs := args[1:]
				c.wrFormat(dest, currentLang, format, fmtArgs...)
				return c
			}
		}
	}

	// PASO 3: Procesamiento de argumentos traducidos
	c.processTranslatedArgs(dest, args, currentLang, 0, separator)
	return c
}

// =============================================================================
// SHARED LANGUAGE SYSTEM FUNCTIONS - REUSED BY ERROR.GO AND TRANSLATION.GO
// =============================================================================

// detectLanguage determines the current language and start index from variadic arguments
// UNIFIED FUNCTION: Handles language detection for both Translate() and wrErr()
// Returns: (language, startIndex) where startIndex skips the language argument if present
func detectLanguage(c *Conv, args []any, allowStringCode bool) (lang, int) {
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
			if l, ok := c.mapLangCode(strVal); ok {
				return l, 1
			}
		}
	}

	// No language specified, use default
	return getCurrentLang(), 0
}

// processTranslatedArgs processes arguments with language-aware translation
// UNIFIED FUNCTION: Handles argument processing for both Translate() and wrErr()
// Eliminates code duplication between Translate() and wrErr()
// REFACTORED: Uses WrString instead of direct buffer access
func (c *Conv) processTranslatedArgs(dest BuffDest, args []any, currentLang lang, startIndex int, separator string) {
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
			c.AnyToBuff(BuffWork, v)
			if c.hasContent(BuffWork) {
				workResult := c.GetString(BuffWork)
				c.WrString(dest, workResult)
				c.ResetBuffer(BuffWork)
			}
		}

		// Agregar separador después, excepto si es el último o el siguiente es separador
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
	// No agregar espacio si es el último argumento
	if currentIndex >= len(args)-1 {
		return false
	}

	// Si el argumento actual termina en newline, espacio, o ciertos separadores específicos, no agregar espacio
	if currentStr, ok := args[currentIndex].(string); ok {
		if len(currentStr) > 0 {
			lastChar := currentStr[len(currentStr)-1]
			// Solo ciertos separadores no necesitan espacio después (como '/')
			if lastChar == '\n' || lastChar == ' ' || lastChar == '/' {
				return false
			}
		}
	}

	// Si el siguiente argumento es un string separador, no agregar espacio
	if nextStr, ok := args[currentIndex+1].(string); ok {
		return !isWordSeparator(nextStr)
	}

	// Para otros tipos (LocStr, etc.) sí agregar espacio
	return true
}
