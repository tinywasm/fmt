package fmt

// Permitted validates strings character-by-character against a configurable whitelist.
//
// Zero value = nothing permitted (strictest). Enable flags to allow character classes.
// Moved from form/input to fmt for cross-layer reuse.
type Permitted struct {
	Letters    bool       // a-z, A-Z, ñ, Ñ
	Tilde      bool       // á, é, í, ó, ú (and uppercase) — uses aL/aU from mapping.go
	Numbers    bool       // 0-9
	Spaces     bool       // ' '
	BreakLine  bool       // '\n'
	Tab        bool       // '\t'
	Extra      []rune     // additional allowed characters (e.g., '@', '.', '-')
	NotAllowed []string   // forbidden substrings
	Minimum    int        // min length (0 = no limit)
	Maximum    int        // max length (0 = no limit)
	StartWith  *Permitted // rules for first character (nil = same as main rules)
}

// Validate checks that text conforms to the permitted rules.
// Order: length → forbidden substrings → start-with → characters.
func (p Permitted) Validate(field, text string) error {
	// Length checks (using range to count runes without importing unicode/utf8)
	var count int
	if p.Minimum != 0 || p.Maximum != 0 {
		for range text {
			count++
		}
	}
	if p.Minimum != 0 && count < p.Minimum {
		return Err(field, "minimum", p.Minimum, "chars")
	}
	if p.Maximum != 0 && count > p.Maximum {
		return Err(field, "maximum", p.Maximum, "chars")
	}

	// Forbidden substrings
	for _, na := range p.NotAllowed {
		if Contains(text, na) {
			return Err(field, "text not allowed", na)
		}
	}

	// StartWith check (first rune only)
	if p.StartWith != nil && len(text) > 0 {
		var firstRune rune
		for _, r := range text {
			firstRune = r
			break
		}
		if err := p.StartWith.Validate(field, string(firstRune)); err != nil {
			return Err(field, "start", err)
		}
	}

	// Character-by-character validation — NO MAPS, only range checks
	for _, r := range text {
		if p.isAllowed(r) {
			continue
		}
		return errCharNotAllowed(field, r)
	}

	return nil
}

// isAllowed checks if a rune is permitted using ASCII ranges and slice lookups.
// Follows the same pattern as mapping.go (toUpperRune, isWordSeparatorChar).
func (p Permitted) isAllowed(r rune) bool {
	// ASCII letters: a-z, A-Z (fast path)
	if p.Letters {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
		if r == 'ñ' || r == 'Ñ' {
			return true
		}
	}

	// Numbers: 0-9 (ASCII range)
	if p.Numbers && r >= '0' && r <= '9' {
		return true
	}

	// Tildes: reuse aL/aU slices from mapping.go
	if p.Tilde {
		for _, a := range aL {
			if r == a {
				return true
			}
		}
		for _, a := range aU {
			if r == a {
				return true
			}
		}
	}

	// Whitespace (individual checks)
	if p.Spaces && r == ' ' {
		return true
	}
	if p.BreakLine && r == '\n' {
		return true
	}
	if p.Tab && r == '\t' {
		return true
	}

	// Extra allowed characters (linear scan, typically 0-5 items)
	for _, c := range p.Extra {
		if r == c {
			return true
		}
	}

	return false
}

func errCharNotAllowed(field string, r rune) error {
	switch r {
	case ' ':
		return Err(field, "space", "not", "allowed")
	case '\t':
		return Err(field, "tab", "not", "allowed")
	case '\n':
		return Err(field, "new", "line", "not", "allowed")
	default:
		return Err(field, "character", "not", "allowed", string(r))
	}
}
