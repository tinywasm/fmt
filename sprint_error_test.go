package fmt

import (
	"errors"
	"testing"
)

func TestSprintDevuelveElMensajeDelError(t *testing.T) {
	err := errors.New("boom")
	if got := Sprint(err); got != "boom" {
		t.Errorf("Sprint(err) = %q, se esperaba %q", got, "boom")
	}
}

func TestSprintConErrorEnvueltoDevuelveLaCadenaCompleta(t *testing.T) {
	wrapped := Errf("contexto: %s", errors.New("boom").Error())
	if got := Sprint(wrapped); got == "" {
		t.Fatal("Sprint(error envuelto) = \"\" — el mensaje se perdió")
	}
	if got, want := Sprint(wrapped), "contexto: boom"; got != want {
		t.Errorf("Sprint(error envuelto) = %q, want %q", got, want)
	}
}

func TestSprintNoRompeLosTiposQueYaFuncionaban(t *testing.T) {
	tests := []struct {
		name string
		v    any
		want string
	}{
		{"string", "hola", "hola"},
		{"int", 42, "42"},
		{"bool", true, "true"},
		{"float64", 1.5, "1.5"},
		{"nil", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sprint(tt.v); got != tt.want {
				t.Errorf("Sprint(%v) = %q, want %q", tt.v, got, tt.want)
			}
		})
	}
}

func TestSprintConErrorNilTipadoNoEntraPorLaRamaDeError(t *testing.T) {
	var e error = nil
	if got := Sprint(e); got != "" {
		t.Errorf("Sprint(error nil tipado) = %q, want \"\" (no debe entrar en la rama de error)", got)
	}
}

func TestPatronDeLogDelEcosistemaNoPierdeElError(t *testing.T) {
	sprintMessages := func(messages ...any) string {
		res := ""
		for i, m := range messages {
			if i > 0 {
				res += " "
			}
			res += Sprint(m)
		}
		return res
	}
	got := sprintMessages("assetmin flush error:", errors.New("disco lleno"))
	want := "assetmin flush error: disco lleno"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}