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
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.ft.String(); got != tt.want {
				t.Errorf("FieldType(%d).String() = %v, want %v", tt.ft, got, tt.want)
			}
		})
	}
}

type stubInput struct{ kind string }

func (s stubInput) Type() string                          { return s.kind }
func (s stubInput) Clone(parentID, name string) Widget     { return s }
func (s stubInput) Validate(v string) error {
	if v == "invalid" {
		return Err(s.kind, "invalid value")
	}
	return nil
}

func TestFieldValidate_WithWidget_Valid(t *testing.T) {
	f := Field{
		Name:   "email",
		Widget: stubInput{kind: "email"},
	}
	if err := f.Validate("valid@email.com"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestFieldValidate_WithWidget_Invalid(t *testing.T) {
	f := Field{
		Name:   "email",
		Widget: stubInput{kind: "email"},
	}
	if err := f.Validate("invalid"); err == nil {
		t.Error("expected error from widget, got nil")
	}
}

func TestFieldValidate_WithWidgetAndPermitted(t *testing.T) {
	f := Field{
		Name:      "email",
		Widget:    stubInput{kind: "email"},
		Permitted: Permitted{Minimum: 10, Letters: true, Extra: []rune{'@', '.'}},
	}
	// Widget passes, but Permitted fails (length)
	if err := f.Validate("abc"); err == nil {
		t.Error("expected error from Permitted.Minimum, got nil")
	}
	// Both pass
	if err := f.Validate("valid@email.com"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestFieldValidate_NilWidget(t *testing.T) {
	f := Field{
		Name:      "name",
		Widget:    nil,
		Permitted: Permitted{Numbers: true},
	}
	if err := f.Validate("123"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if err := f.Validate("abc"); err == nil {
		t.Error("expected error from Permitted.Numbers, got nil")
	}
}

func TestFieldValidate_NotNull_EmptyValue(t *testing.T) {
	var called bool
	stub := &stubInputWithCallback{
		stubInput: stubInput{kind: "text"},
		callback:  func() { called = true },
	}
	f := Field{
		Name:    "name",
		NotNull: true,
		Widget:  stub,
	}

	if err := f.Validate(""); err == nil {
		t.Error("expected error for NotNull, got nil")
	}

	if called {
		t.Error("Widget.Validate should NOT have been called for empty value when NotNull fails")
	}
}

type stubInputWithCallback struct {
	stubInput
	callback func()
}

func (s *stubInputWithCallback) Validate(v string) error {
	s.callback()
	return s.stubInput.Validate(v)
}

func TestWidgetType(t *testing.T) {
	s := stubInput{kind: "email"}
	if s.Type() != "email" {
		t.Errorf("expected Type() = \"email\", got %q", s.Type())
	}
}

func TestWidgetClone(t *testing.T) {
	s := stubInput{kind: "textarea"}
	c := s.Clone("parent", "field")
	if c == nil {
		t.Fatal("Clone() returned nil")
	}
	if c.Type() != "textarea" {
		t.Errorf("Clone().Type() = %q, want \"textarea\"", c.Type())
	}
	// Clone must be independent — mutating original should not affect clone
	// (value receiver, so already independent by definition)
}

func TestFieldValidate_WidgetRunsBeforePermitted(t *testing.T) {
	// Widget fails → Permitted must NOT be evaluated (order: Widget first)
	permittedCalled := false
	// Use a Permitted that would succeed — if it's called, we detect it via a custom Permitted
	// Instead: use Widget that fails + Permitted that would also fail, then check error source
	f := Field{
		Name:      "field",
		Widget:    stubInput{kind: "text"}, // fails on "invalid"
		Permitted: Permitted{Minimum: 100}, // would also fail (too short)
	}
	err := f.Validate("invalid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Error must come from Widget (contains "text"), not from Permitted (contains "minimum")
	msg := err.Error()
	if !containsAny(msg, "invalid value") {
		t.Errorf("expected Widget error, got: %q", msg)
	}
	_ = permittedCalled
}

func containsAny(s string, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestFieldZeroValue(t *testing.T) {
	var f Field
	if f.Name != "" {
		t.Errorf("expected empty Name, got %v", f.Name)
	}
	if f.Type != FieldText {
		t.Errorf("expected FieldText type, got %v", f.Type)
	}
	if f.IsPK() || f.IsUnique() || f.NotNull || f.IsAutoInc() {
		t.Errorf("expected all bools false, got PK=%v, Unique=%v, NotNull=%v, AutoInc=%v", f.IsPK(), f.IsUnique(), f.NotNull, f.IsAutoInc())
	}
}

func TestFieldOmitEmpty(t *testing.T) {
	f := Field{OmitEmpty: true}
	if !f.OmitEmpty {
		t.Error("expected OmitEmpty true")
	}
}

func TestFieldConstraints(t *testing.T) {
	f := Field{
		Name:    "id",
		Type:    FieldInt,
		NotNull: true,
		DB: &FieldDB{
			PK:      true,
			Unique:  true,
			AutoInc: true,
		},
	}
	if f.Name != "id" {
		t.Errorf("expected Name 'id', got %v", f.Name)
	}
	if f.Type != FieldInt {
		t.Errorf("expected FieldInt type, got %v", f.Type)
	}
	if !f.IsPK() {
		t.Error("expected PK true")
	}
	if !f.IsUnique() {
		t.Error("expected Unique true")
	}
	if !f.NotNull {
		t.Error("expected NotNull true")
	}
	if !f.IsAutoInc() {
		t.Error("expected AutoInc true")
	}
}

func TestFieldValidate(t *testing.T) {
	tests := []struct {
		name    string
		field   Field
		value   string
		wantErr bool
	}{
		{"NotNull empty", Field{NotNull: true}, "", true},
		{"NotNull not empty", Field{NotNull: true}, "foo", false},
		{"Nullable empty", Field{NotNull: false}, "", false},
		{"With rules pass", Field{Permitted: Permitted{Numbers: true}}, "123", false},
		{"With rules fail", Field{Permitted: Permitted{Numbers: true}}, "abc", true},
		{"No rules pass", Field{Name: "any"}, "any value", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.field.Validate(tt.value); (err != nil) != tt.wantErr {
				t.Errorf("Field.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

type fullMock struct {
	Text   string
	Int    int
	Int32  int32
	Int64  int64
	Uint   uint
	Uint32 uint32
	Uint64 uint64
	Float  float64
	Float32 float32
	Bool   bool
	Blob   []byte
	Nested *mockUser
}

func (m *fullMock) Schema() []Field {
	return []Field{
		{Name: "text", Type: FieldText, NotNull: true},
		{Name: "int", Type: FieldInt, NotNull: true},
		{Name: "int32", Type: FieldInt},
		{Name: "int64", Type: FieldInt},
		{Name: "uint", Type: FieldInt},
		{Name: "uint32", Type: FieldInt},
		{Name: "uint64", Type: FieldInt},
		{Name: "float", Type: FieldFloat, NotNull: true},
		{Name: "float32", Type: FieldFloat},
		{Name: "bool", Type: FieldBool, NotNull: true},
		{Name: "blob", Type: FieldBlob, NotNull: true},
		{Name: "nested", Type: FieldStruct, NotNull: true},
	}
}

func (m *fullMock) Pointers() []any {
	return []any{&m.Text, &m.Int, &m.Int32, &m.Int64, &m.Uint, &m.Uint32, &m.Uint64, &m.Float, &m.Float32, &m.Bool, &m.Blob, m.Nested}
}

type mockUser struct {
	id   string
	name string
}

func (m *mockUser) Schema() []Field {
	return []Field{
		{Name: "id", Type: FieldText, DB: &FieldDB{PK: true}},
		{Name: "name", Type: FieldText, NotNull: true},
	}
}
func (m *mockUser) Pointers() []any { return []any{&m.id, &m.name} }
func (m *mockUser) Validate(action byte) error { return ValidateFields(action, m) }

func TestValidateFieldsRecursive(t *testing.T) {
	m := &fullMock{
		Text: "hello",
		Int: 1,
		Float: 1.1,
		Bool: true,
		Blob: []byte{1},
		Nested: &mockUser{id: "u1", name: "Alice"},
	}

	if err := ValidateFields('u', m); err != nil {
		t.Errorf("expected success, got %v", err)
	}

	// Fail nested
	m.Nested.name = ""
	if err := ValidateFields('u', m); err == nil {
		t.Error("expected failure in nested struct")
	}

	// Fail other types
	m.Nested.name = "Alice"
	m.Int = 0 // NotNull
	if err := ValidateFields('u', m); err == nil {
		t.Error("expected failure for int zero")
	}
}

func TestReadValuesAllTypes(t *testing.T) {
	m := &fullMock{
		Text: "text",
		Int: 10,
		Int32: 32,
		Int64: 64,
		Uint: 100,
		Uint32: 320,
		Uint64: 640,
		Float: 1.1,
		Float32: 2.2,
		Bool: true,
		Blob: []byte{0x01},
		Nested: &mockUser{id: "u1", name: "Alice"},
	}
	schema := m.Schema()
	ptrs := m.Pointers()
	vals := ReadValues(schema, ptrs)

	expected := []any{"text", 10, int32(32), int64(64), uint(100), uint32(320), uint64(640), 1.1, float32(2.2), true, []byte{0x01}, m.Nested}
	for i, v := range expected {
		if i == 10 { // Blob check
			b1 := v.([]byte)
			b2 := vals[i].([]byte)
			if len(b1) != len(b2) || b1[0] != b2[0] {
				t.Errorf("mismatch at index %d: got %v, want %v", i, vals[i], v)
			}
			continue
		}
		if vals[i] != v {
			t.Errorf("mismatch at index %d: got %v, want %v", i, vals[i], v)
		}
	}

	// Test nil pointers
	ptrsCopy := make([]any, len(ptrs))
	copy(ptrsCopy, ptrs)
	ptrsCopy[0] = nil
	valsNil := ReadValues(schema, ptrsCopy)
	if valsNil[0] != nil {
		t.Error("expected nil for nil pointer")
	}
}

func TestIsZeroPtrAllTypes(t *testing.T) {
	var i int = 0
	if !isZeroPtr(&i, FieldInt) { t.Error("int 0 should be zero") }
	i = 1
	if isZeroPtr(&i, FieldInt) { t.Error("int 1 should not be zero") }

	var i32 int32 = 0
	if !isZeroPtr(&i32, FieldInt) { t.Error("int32 0 should be zero") }
	var i64 int64 = 0
	if !isZeroPtr(&i64, FieldInt) { t.Error("int64 0 should be zero") }
	var u uint = 0
	if !isZeroPtr(&u, FieldInt) { t.Error("uint 0 should be zero") }
	var u32 uint32 = 0
	if !isZeroPtr(&u32, FieldInt) { t.Error("uint32 0 should be zero") }
	var u64 uint64 = 0
	if !isZeroPtr(&u64, FieldInt) { t.Error("uint64 0 should be zero") }

	var f64 float64 = 0
	if !isZeroPtr(&f64, FieldFloat) { t.Error("float64 0 should be zero") }
	f64 = 0.1
	if isZeroPtr(&f64, FieldFloat) { t.Error("float64 0.1 should not be zero") }
	var f32 float32 = 0
	if !isZeroPtr(&f32, FieldFloat) { t.Error("float32 0 should be zero") }

	var b bool = false
	if !isZeroPtr(&b, FieldBool) { t.Error("bool false should be zero") }
	b = true
	if isZeroPtr(&b, FieldBool) { t.Error("bool true should not be zero") }

	var bl []byte
	if !isZeroPtr(&bl, FieldBlob) { t.Error("nil blob should be zero") }
	bl = []byte{}
	if !isZeroPtr(&bl, FieldBlob) { t.Error("empty blob should be zero") }
	bl = []byte{1}
	if isZeroPtr(&bl, FieldBlob) { t.Error("non-empty blob should not be zero") }
}

func TestReadStringPtr(t *testing.T) {
	s := "hello"
	val, ok := ReadStringPtr(&s)
	if !ok || val != "hello" {
		t.Errorf("ReadStringPtr failed: got (%q, %v), want (\"hello\", true)", val, ok)
	}

	val, ok = ReadStringPtr("not a pointer")
	if ok {
		t.Error("ReadStringPtr should have failed for non-pointer")
	}

	var ns *string
	val, ok = ReadStringPtr(ns)
	if ok {
		t.Error("ReadStringPtr should have failed for nil pointer")
	}
}

type fielderOnlyMock struct {
	id string
}
func (m *fielderOnlyMock) Schema() []Field { return []Field{{Name: "id", Type: FieldText}} }
func (m *fielderOnlyMock) Pointers() []any { return []any{&m.id} }

func TestValidateFieldsWithOnlyFielder(t *testing.T) {
	// Nested struct that only implements Fielder, not Validator.
	sub := &fielderOnlyMock{id: "ok"}
	schema := []Field{{Name: "sub", Type: FieldStruct}}
	ptrs := []any{sub}

	// Validate it through the manualFielder helper
	if err := ValidateFields('u', &manualFielder{schema, ptrs}); err != nil {
		t.Errorf("expected success, got %v", err)
	}
}

func TestValidateFieldsActions(t *testing.T) {
	type userActionMock struct {
		ID      int
		Name    string
		Email   string
		Version int
	}

	schema := []Field{
		{Name: "id", Type: FieldInt, DB: &FieldDB{PK: true, AutoInc: true}},
		{Name: "name", Type: FieldText, NotNull: true},
		{Name: "email", Type: FieldText, Permitted: Permitted{Letters: true, Extra: []rune{'@', '.'}}},
		{Name: "version", Type: FieldInt, NotNull: true},
	}

	// Helper to create manual Fielder
	getFielder := func(m *userActionMock) Fielder {
		return &manualFielder{
			schema: schema,
			ptrs:   []any{&m.ID, &m.Name, &m.Email, &m.Version},
		}
	}

	t.Run("Create 'c'", func(t *testing.T) {
		m := &userActionMock{Name: "Alice", Email: "a@b.com", Version: 1}
		f := getFielder(m)

		// PK+AutoInc should be skipped in 'c'
		if err := ValidateFields('c', f); err != nil {
			t.Errorf("expected success in 'c' with zero ID, got %v", err)
		}

		// NotNull still applies
		m.Name = ""
		if err := ValidateFields('c', f); err == nil {
			t.Error("expected failure in 'c' with empty Name")
		}
	})

	t.Run("Update 'u'", func(t *testing.T) {
		m := &userActionMock{ID: 1, Name: "Alice", Email: "a@b.com", Version: 1}
		f := getFielder(m)

		if err := ValidateFields('u', f); err != nil {
			t.Errorf("expected success in 'u', got %v", err)
		}

		// PK is required in 'u'
		m.ID = 0
		if err := ValidateFields('u', f); err == nil {
			t.Error("expected failure in 'u' with zero ID")
		}
	})

	t.Run("Delete 'd'", func(t *testing.T) {
		m := &userActionMock{ID: 1, Name: "", Email: "invalid!!", Version: 0}
		f := getFielder(m)

		// Only PK matters in 'd'
		if err := ValidateFields('d', f); err != nil {
			t.Errorf("expected success in 'd' with only PK, got %v", err)
		}

		// PK missing
		m.ID = 0
		if err := ValidateFields('d', f); err == nil {
			t.Error("expected failure in 'd' with missing PK")
		}
	})

	t.Run("Unknown 'x'", func(t *testing.T) {
		m := &userActionMock{ID: 1, Name: "Alice", Email: "a@b.com", Version: 1}
		f := getFielder(m)

		// Should behave like 'u'
		if err := ValidateFields('x', f); err != nil {
			t.Errorf("expected success in 'x', got %v", err)
		}

		m.ID = 0
		if err := ValidateFields('x', f); err == nil {
			t.Error("expected failure in 'x' with missing PK")
		}
	})
}

type manualFielder struct {
	schema []Field
	ptrs   []any
}
func (f *manualFielder) Schema() []Field { return f.schema }
func (f *manualFielder) Pointers() []any { return f.ptrs }
