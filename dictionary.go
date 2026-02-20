package fmt

// DictEntry describes one translatable word.
// EN serves as lookup key (case-insensitive) and English display value.
// Other language fields fall back to EN when empty.
type DictEntry struct {
	EN string
	ES string
	FR string
	DE string
	ZH string
	HI string
	AR string
	PT string
	RU string
}

const langCount = 9

type entry struct {
	translations [langCount]string
}

var dictEntries []entry

// RegisterWords adds entries to the lookup engine. Safe to call from init().
func RegisterWords(entries []DictEntry) {
	for _, de := range entries {
		if de.EN == "" {
			continue
		}
		// Skip entries whose EN is a known language code (len==2):
		// detectLanguage would consume it before lookupWord ever sees it.
		if len(de.EN) == 2 {
			c1, c2 := de.EN[0]|32, de.EN[1]|32
			switch [2]byte{c1, c2} {
			case [2]byte{'e', 'n'}, [2]byte{'e', 's'}, [2]byte{'z', 'h'},
				[2]byte{'h', 'i'}, [2]byte{'a', 'r'}, [2]byte{'p', 't'},
				[2]byte{'f', 'r'}, [2]byte{'d', 'e'}, [2]byte{'r', 'u'}:
				continue
			}
		}

		// Check if the word already exists to merge translations
		idx := -1
		for j, exist := range dictEntries {
			if compareCaseInsensitive(exist.translations[EN], de.EN) == 0 {
				idx = j
				break
			}
		}

		if idx >= 0 {
			// Merge existing entry
			if de.ES != "" {
				dictEntries[idx].translations[ES] = de.ES
			}
			if de.ZH != "" {
				dictEntries[idx].translations[ZH] = de.ZH
			}
			if de.HI != "" {
				dictEntries[idx].translations[HI] = de.HI
			}
			if de.AR != "" {
				dictEntries[idx].translations[AR] = de.AR
			}
			if de.PT != "" {
				dictEntries[idx].translations[PT] = de.PT
			}
			if de.FR != "" {
				dictEntries[idx].translations[FR] = de.FR
			}
			if de.DE != "" {
				dictEntries[idx].translations[DE] = de.DE
			}
			if de.RU != "" {
				dictEntries[idx].translations[RU] = de.RU
			}
		} else {
			// Add new entry
			var e entry
			e.translations[EN] = de.EN
			e.translations[ES] = de.ES
			e.translations[ZH] = de.ZH
			e.translations[HI] = de.HI
			e.translations[AR] = de.AR
			e.translations[PT] = de.PT
			e.translations[FR] = de.FR
			e.translations[DE] = de.DE
			e.translations[RU] = de.RU
			for i := 1; i < langCount; i++ {
				if e.translations[i] == "" {
					e.translations[i] = de.EN
				}
			}
			dictEntries = append(dictEntries, e)
		}
	}
	sortDict()
}

func lookupWord(word string, l lang) (string, bool) {
	if len(dictEntries) == 0 || word == "" {
		return "", false
	}
	low, high := 0, len(dictEntries)-1
	for low <= high {
		mid := low + (high-low)/2
		cmp := compareCaseInsensitive(word, dictEntries[mid].translations[EN])
		if cmp == 0 {
			return dictEntries[mid].translations[int(l)], true
		}
		if cmp < 0 {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}
	return "", false
}

func sortDict() {
	if len(dictEntries) < 2 {
		return
	}
	quicksort(dictEntries, 0, len(dictEntries)-1)
}

func quicksort(data []entry, low, high int) {
	if low < high {
		p := partition(data, low, high)
		quicksort(data, low, p)
		quicksort(data, p+1, high)
	}
}

func partition(data []entry, low, high int) int {
	pivot := data[(low+high)/2].translations[EN]
	i, j := low-1, high+1
	for {
		i++
		for compareCaseInsensitive(data[i].translations[EN], pivot) < 0 {
			i++
		}
		j--
		for compareCaseInsensitive(data[j].translations[EN], pivot) > 0 {
			j--
		}
		if i >= j {
			return j
		}
		data[i], data[j] = data[j], data[i]
	}
}

// compareCaseInsensitive compares two ASCII strings case-insensitively.
func compareCaseInsensitive(s1, s2 string) int {
	n1, n2 := len(s1), len(s2)
	n := n1
	if n2 < n {
		n = n2
	}
	for i := 0; i < n; i++ {
		c1, c2 := s1[i], s2[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 += 32
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 += 32
		}
		if c1 < c2 {
			return -1
		}
		if c1 > c2 {
			return 1
		}
	}
	if n1 < n2 {
		return -1
	}
	if n1 > n2 {
		return 1
	}
	return 0
}
