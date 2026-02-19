package fmt

import (
	"testing"
)

func TestRegisterWords(t *testing.T) {
	RegisterWords([]DictEntry{
		{EN: "VALE", ES: "Vale"},
	})

	// Pass lang explicitly to avoid 2-char strings being misread as lang codes
	got := Translate(EN, "vale").String()
	if got != "VALE" {
		t.Errorf("EN: want %q got %q", "VALE", got)
	}
	got = Translate(ES, "vale").String()
	if got != "Vale" {
		t.Errorf("ES: want %q got %q", "Vale", got)
	}

	// Unknown word → pass-through
	got = Translate(EN, "xyz_not_a_word").String()
	if got != "xyz_not_a_word" {
		t.Errorf("pass-through: want %q got %q", "xyz_not_a_word", got)
	}
}

func TestRegisterWordsValidation(t *testing.T) {
	// Empty EN → silently skipped, not registered
	RegisterWords([]DictEntry{
		{EN: "", ES: "should not register"},
	})
	got := Translate(EN, "").String()
	if got != "" {
		t.Errorf("empty EN should not register, got %q", got)
	}

	// Lang codes as EN → silently skipped (would never be reachable via lookup)
	RegisterWords([]DictEntry{
		{EN: "EN", ES: "not reachable"},
		{EN: "es", ES: "not reachable"},
		{EN: "ZH", ES: "not reachable"},
	})
	// These must pass-through unchanged, not return any ES translation
	for _, key := range []string{"EN", "en", "es", "ZH", "zh"} {
		got := Translate(ES, key).String()
		if got == "not reachable" {
			t.Errorf("lang code %q should not be registered as a word", key)
		}
	}
}

func TestTranslateMixedArgs(t *testing.T) {
	RegisterWords([]DictEntry{
		{EN: "file", ES: "archivo"},
		{EN: "not", ES: "no"},
		{EN: "found", ES: "encontrado"},
	})

	OutLang(ES)
	got := Translate("file", "not", "found").String()
	want := "archivo no encontrado"
	if got != want {
		t.Errorf("want %q got %q", want, got)
	}
}

func TestErrMixedArgs(t *testing.T) {
	OutLang(EN)
	err := Err("xyz_test", "xyz_err").Error()
	want := "xyz_test xyz_err"
	if err != want {
		t.Errorf("want %q got %q", want, err)
	}
}
