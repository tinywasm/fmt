package fmt

// Private global configuration with mutex protection
var (
	defLang lang = EN
)

// Language enumeration for supported languages
type lang uint8

// String devuelve el nombre del lenguaje como string (ej: EN => "EN")
func (l lang) String() string {
	switch l {
	case EN:
		return "EN"
	case ES:
		return "ES"
	case ZH:
		return "ZH"
	case HI:
		return "HI"
	case AR:
		return "AR"
	case PT:
		return "PT"
	case FR:
		return "FR"
	case DE:
		return "DE"
	case RU:
		return "RU"
	default:
		return "EN" // fallback
	}
}

const (
	// Group 1: Core Essential Languages (Maximum Global Reach)
	EN lang = iota // 0 - English (default)
	ES             // 1 - Spanish
	ZH             // 2 - Chinese
	HI             // 3 - Hindi
	AR             // 4 - Arabic

	// Group 2: Extended Reach Languages (Europe & Americas)
	PT // 5 - Portuguese
	FR // 6 - French
	DE // 7 - German
	RU // 8 - Russian

	// Group 3: Regional Languages (Commented out to reduce binary size)
	// IT             // Italian
	// ID             // Indonesian
	// BN             // Bengali
	// UR             // Urdu
)

// LocStr represents a string with translations for multiple languages.
//
// It is a fixed-size array where each index corresponds to a language constant
// (EN, ES, PT, etc.). This design ensures type safety and efficiency, as the
// compiler can verify that all translations are provided.
//
// The order of translations must match the order of the language constants.
//
// Example of creating a new translatable term for "File":
//
//	var MyDictionary = struct {
//		File LocStr
//	}{
//		File: LocStr{
//			EN: "file",
//			ES: "archivo",
//			ZH: "文件",
//			HI: "फ़ाइल",
//			AR: "ملف",
//			PT: "arquivo",
//			FR: "fichier",
//			DE: "Datei",
//			RU: "файл",
//		},
//	}
//
// Usage in code:
//
//	err := Err(MyDictionary.File, D.Not, D.Found) // -> "file not found", "archivo no encontrado", etc.
type LocStr [9]string

// OutLang sets and returns the current output language as a string.
//
// OutLang()                // Auto-detects system/browser language, returns code (e.g. "EN")
// OutLang(ES)              // Set Spanish as default (using lang constant), returns "ES"
// OutLang("ES")            // Set Spanish as default (using string code), returns "ES"
// OutLang("fr")            // Set French as default (case-insensitive), returns "FR"
// OutLang("en-US")         // Accepts locale strings, parses to EN, returns "EN"
//
// If a string is passed, it is automatically parsed using supported codes.
// If a lang value is passed, it is assigned directly.
// If another type is passed, nothing happens.
// Always returns the current language code as string (e.g. "EN", "ES", etc).
func OutLang(l ...any) string {
	c := GetConv()
	if len(l) == 0 {
		systemLang := c.getSystemLang() // Get system lang without holding lock
		setDefaultLang(systemLang)
		return systemLang.String()
	}

	var newLang lang
	switch v := l[0].(type) {
	case lang:
		newLang = v
	case string:
		newLang = c.langParser(v)
	default:
		// Return current language without changes
		return getCurrentLang().String()
	}

	setDefaultLang(newLang)
	return newLang.String()
}

// langParser processes a list of language strings (e.g., from env vars or browser settings)
// and returns the first valid language found. It centralizes the parsing logic for both
// frontend and backend environments.
func (c *Conv) langParser(langStrings ...string) lang {

	for _, langStr := range langStrings {
		if langStr == "" {
			continue
		}

		// Parse language code from the string, handling common formats using internal splitStr.
		code := c.splitStr(langStr, ".")[0] // Removes encoding, e.g., ".UTF-8"
		code = c.splitStr(code, "_")[0]     // Handles locale format, e.g., "en_US"
		code = c.splitStr(code, "-")[0]     // Handles standard format, e.g., "en-US"

		if code == "" {
			continue
		}

		// Inline mapLangCode logic
		return c.mapLangCode(code)
	}

	// c.putConv()

	return EN // Default fallback if no valid language string is found.
}

func (c *Conv) mapLangCode(strVal string) lang {

	// Convert to lowercase and map to internal lang typec
	c.ResetBuffer(BuffWork) // Clear work buffer before use
	c.WrString(BuffWork, strVal)
	// use changeCase
	c.changeCase(true, BuffWork)

	code := c.GetString(BuffWork) // Get lowercase string

	switch code {
	// Group 1
	case "en":
		return EN
	case "es":
		return ES
	case "zh":
		return ZH
	case "hi":
		return HI
	case "ar":
		return AR
	// Group 2
	case "pt":
		return PT
	case "fr":
		return FR
	case "de":
		return DE
	case "ru":
		return RU
	}
	return EN // Default fallback
}
