package lang

import "github.com/tinywasm/fmt"

func init() {
	fmt.SetTranslator(func(word string) (string, bool) {
		return lookupWord(word, getCurrentLang())
	})
}
