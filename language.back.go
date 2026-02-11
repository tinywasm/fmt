//go:build !wasm

package fmt

import "sync"

var defLangMu sync.RWMutex

func setDefaultLang(l lang) {
	defLangMu.Lock()
	defLang = l
	defLangMu.Unlock()
}

func getCurrentLang() lang {
	defLangMu.RLock()
	defer defLangMu.RUnlock()
	return defLang
}
