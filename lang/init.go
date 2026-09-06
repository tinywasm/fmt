package lang

import "webtyp.com/fmt"

func init() {
	fmt.SetTranslator(func(word string) (string, bool) {
		return lookupWord(word, getCurrentLang())
	})
}
