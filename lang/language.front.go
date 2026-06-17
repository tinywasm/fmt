//go:build wasm

package lang

// WASM is single-threaded: no mutex needed
func setDefaultLang(l lang) {
	defLang = l
}

func getCurrentLang() lang {
	return defLang
}
