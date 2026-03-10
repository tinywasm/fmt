package fmt

import (
	"testing"
)

func TestFieldTypeString(t *testing.T) {
	tests := []struct {
		ft   FieldType
		want string
	}{
		{FieldText, "text"},
		{FieldInt, "int"},
		{FieldFloat, "float"},
		{FieldBool, "bool"},
		{FieldBlob, "blob"},
		{FieldStruct, "struct"},
		{FieldType(-1), "unknown"},
		{FieldType(6), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.ft.String(); got != tt.want {
			t.Errorf("FieldType(%d).String() = %v, want %v", tt.ft, got, tt.want)
		}
	}
}

func TestFieldZeroValue(t *testing.T) {
	var f Field
	if f.Name != "" {
		t.Errorf("expected empty Name, got %v", f.Name)
	}
	if f.Type != FieldText {
		t.Errorf("expected FieldText type, got %v", f.Type)
	}
	if f.PK || f.Unique || f.NotNull || f.AutoInc {
		t.Errorf("expected all bools false, got PK=%v, Unique=%v, NotNull=%v, AutoInc=%v", f.PK, f.Unique, f.NotNull, f.AutoInc)
	}
	if f.Input != "" {
		t.Errorf("expected empty Input, got %v", f.Input)
	}
	if f.JSON != "" {
		t.Errorf("expected empty JSON, got %v", f.JSON)
	}
}

func TestFieldInputHint(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"email", "email"},
		{"exclude", "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Field{Input: tt.input}
			if f.Input != tt.input {
				t.Errorf("expected Input %q, got %q", tt.input, f.Input)
			}
		})
	}
}

func TestFieldJSON(t *testing.T) {
	tests := []struct {
		name string
		json string
	}{
		{"empty", ""},
		{"key", "email"},
		{"omitempty", "email,omitempty"},
		{"exclude", "-"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Field{JSON: tt.json}
			if f.JSON != tt.json {
				t.Errorf("expected JSON %q, got %q", tt.json, f.JSON)
			}
		})
	}
}

func TestFieldConstraints(t *testing.T) {
	f := Field{
		Name:    "id",
		Type:    FieldInt,
		PK:      true,
		Unique:  true,
		NotNull: true,
		AutoInc: true,
	}
	if f.Name != "id" {
		t.Errorf("expected Name 'id', got %v", f.Name)
	}
	if f.Type != FieldInt {
		t.Errorf("expected FieldInt type, got %v", f.Type)
	}
	if !f.PK {
		t.Error("expected PK true")
	}
	if !f.Unique {
		t.Error("expected Unique true")
	}
	if !f.NotNull {
		t.Error("expected NotNull true")
	}
	if !f.AutoInc {
		t.Error("expected AutoInc true")
	}
}

func TestFieldStructType(t *testing.T) {
	f := Field{
		Name: "profile",
		Type: FieldStruct,
	}
	if f.Type != FieldStruct {
		t.Errorf("expected FieldStruct type, got %v", f.Type)
	}
	if f.Type.String() != "struct" {
		t.Errorf("expected 'struct' string, got %v", f.Type.String())
	}
}

type mockUser struct {
	id   string
	name string
}

func (m *mockUser) Schema() []Field {
	return []Field{
		{Name: "id", Type: FieldText, PK: true},
		{Name: "name", Type: FieldText, NotNull: true},
	}
}
func (m *mockUser) Values() []any  { return []any{m.id, m.name} }
func (m *mockUser) Pointers() []any { return []any{&m.id, &m.name} }

func TestFielderInterface(t *testing.T) {
	m := &mockUser{id: "u1", name: "Alice"}
	var i any = m
	f, ok := i.(Fielder)
	if !ok {
		t.Fatal("mockUser does not implement Fielder")
	}

	schema := f.Schema()
	values := f.Values()
	pointers := f.Pointers()

	if len(schema) != 2 || len(values) != 2 || len(pointers) != 2 {
		t.Errorf("length mismatch: schema=%d, values=%d, pointers=%d", len(schema), len(values), len(pointers))
	}

	if schema[0].Name != "id" || schema[1].Name != "name" {
		t.Errorf("schema name mismatch: %v, %v", schema[0].Name, schema[1].Name)
	}

	if values[0] != "u1" || values[1] != "Alice" {
		t.Errorf("value mismatch: %v, %v", values[0], values[1])
	}

	// Test writing through pointers
	*(pointers[0].(*string)) = "u2"
	*(pointers[1].(*string)) = "Bob"

	if m.id != "u2" || m.name != "Bob" {
		t.Errorf("pointer update failed: id=%s, name=%s", m.id, m.name)
	}

	newValues := f.Values()
	if newValues[0] != "u2" || newValues[1] != "Bob" {
		t.Errorf("values after update mismatch: %v, %v", newValues[0], newValues[1])
	}
}
