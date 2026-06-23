package fmt

import (
	"bytes"
	"testing"
)

type mockFieldWriter struct {
	buf  *bytes.Buffer
	conv *Conv // reused across all fields — keeps the writer 0-alloc (GetConv allocates in WASM)
}

func newMockFieldWriter(buf *bytes.Buffer) *mockFieldWriter {
	return &mockFieldWriter{buf: buf, conv: GetConv()}
}

func (m *mockFieldWriter) String(name, val string) {
	m.buf.WriteString(name)
	m.buf.WriteByte('=')
	m.buf.WriteString(val)
	m.buf.WriteByte(';')
}

func (m *mockFieldWriter) Int(name string, val int64) {
	m.buf.WriteString(name)
	m.buf.WriteByte('=')
	m.conv.ResetBuffer(BuffWork)
	m.conv.wrIntBase(BuffWork, val, 10, true)
	m.buf.Write(m.conv.getBytes(BuffWork))
	m.buf.WriteByte(';')
}

func (m *mockFieldWriter) Uint(name string, val uint64) {
	m.buf.WriteString(name)
	m.buf.WriteByte('=')
	m.conv.ResetBuffer(BuffWork)
	m.conv.wrIntBase(BuffWork, int64(val), 10, false)
	m.buf.Write(m.conv.getBytes(BuffWork))
	m.buf.WriteByte(';')
}

func (m *mockFieldWriter) Float(name string, val float64) {
	m.buf.WriteString(name)
	m.buf.WriteByte('=')
	m.conv.ResetBuffer(BuffWork)
	m.conv.wrFloat64(BuffWork, val)
	m.buf.Write(m.conv.getBytes(BuffWork))
	m.buf.WriteByte(';')
}

func (m *mockFieldWriter) Bool(name string, val bool) {
	m.buf.WriteString(name)
	m.buf.WriteByte('=')
	if val {
		m.buf.WriteString("true")
	} else {
		m.buf.WriteString("false")
	}
	m.buf.WriteByte(';')
}

func (m *mockFieldWriter) Bytes(name string, val []byte) {
	m.buf.WriteString(name)
	m.buf.WriteByte('=')
	m.buf.Write(val)
	m.buf.WriteByte(';')
}

func (m *mockFieldWriter) Null(name string) {
	m.buf.WriteString(name)
	m.buf.WriteString("=null;")
}

func (m *mockFieldWriter) Raw(name, val string) {
	m.buf.WriteString(name)
	m.buf.WriteByte('=')
	m.buf.WriteString(val)
	m.buf.WriteByte(';')
}

func (m *mockFieldWriter) Object(name string, val Encodable) {
	m.buf.WriteString(name)
	m.buf.WriteString("={")
	val.EncodeFields(m)
	m.buf.WriteString("};")
}

func (m *mockFieldWriter) Array(name string, n int) ArrayWriter {
	m.buf.WriteString(name)
	m.buf.WriteString("=[")
	return &mockArrayWriter{m: m, first: true}
}

type mockArrayWriter struct {
	m     *mockFieldWriter
	first bool
}

func (a *mockArrayWriter) maybeComma() {
	if !a.first {
		a.m.buf.WriteByte(',')
	}
	a.first = false
}

func (a *mockArrayWriter) String(val string) {
	a.maybeComma()
	a.m.buf.WriteString(val)
}
func (a *mockArrayWriter) Int(val int64) {
	a.maybeComma()
	a.m.conv.ResetBuffer(BuffWork)
	a.m.conv.wrIntBase(BuffWork, val, 10, true)
	a.m.buf.Write(a.m.conv.getBytes(BuffWork))
}
func (a *mockArrayWriter) Float(val float64) {
	a.maybeComma()
	a.m.conv.ResetBuffer(BuffWork)
	a.m.conv.wrFloat64(BuffWork, val)
	a.m.buf.Write(a.m.conv.getBytes(BuffWork))
}
func (a *mockArrayWriter) Bool(val bool) {
	a.maybeComma()
	if val {
		a.m.buf.WriteString("true")
	} else {
		a.m.buf.WriteString("false")
	}
}
func (a *mockArrayWriter) Bytes(val []byte) {
	a.maybeComma()
	a.m.buf.Write(val)
}
func (a *mockArrayWriter) Object(val Encodable) {
	a.maybeComma()
	a.m.buf.WriteByte('{')
	val.EncodeFields(a.m)
	a.m.buf.WriteByte('}')
}
func (a *mockArrayWriter) Close() {
	a.m.buf.WriteString("];")
}

type mockFieldReader struct {
	fields [][]byte // Each element is "name=value"
}

func newMockFieldReader(data []byte) *mockFieldReader {
	return &mockFieldReader{
		fields: bytes.Split(data, []byte(";")),
	}
}

func (r *mockFieldReader) find(name string) ([]byte, bool) {
	prefix := []byte(name + "=")
	for _, f := range r.fields {
		if bytes.HasPrefix(f, prefix) {
			return f[len(prefix):], true
		}
	}
	return nil, false
}

func (r *mockFieldReader) String(name string) (string, bool) {
	val, ok := r.find(name)
	return string(val), ok
}

func (r *mockFieldReader) Int(name string) (int64, bool) {
	val, ok := r.find(name)
	if !ok { return 0, false }
	c := GetConv()
	c.LoadBytes(val)
	v, _ := c.Int64()
	c.PutConv()
	return v, true
}

func (r *mockFieldReader) Uint(name string) (uint64, bool) {
	val, ok := r.find(name)
	if !ok { return 0, false }
	c := GetConv()
	c.LoadBytes(val)
	v, _ := c.Uint64()
	c.PutConv()
	return v, true
}

func (r *mockFieldReader) Float(name string) (float64, bool) {
	val, ok := r.find(name)
	if !ok { return 0, false }
	c := GetConv()
	c.LoadBytes(val)
	v, _ := c.Float64()
	c.PutConv()
	return v, true
}

func (r *mockFieldReader) Bool(name string) (bool, bool) {
	val, ok := r.find(name)
	if !ok { return false, false }
	return string(val) == "true", true
}

func (r *mockFieldReader) Bytes(name string) ([]byte, bool) {
	return r.find(name)
}

func (r *mockFieldReader) Object(name string, into Decodable) bool {
	val, ok := r.find(name)
	if !ok || len(val) < 2 || val[0] != '{' || val[len(val)-1] != '}' {
		return false
	}
	innerReader := newMockFieldReader(val[1 : len(val)-1])
	return into.DecodeFields(innerReader) == nil
}

func (r *mockFieldReader) Array(name string) (ArrayReader, bool) {
	val, ok := r.find(name)
	if !ok || len(val) < 2 || val[0] != '[' || val[len(val)-1] != ']' {
		return nil, false
	}
	return &mockArrayReader{data: val[1 : len(val)-1]}, true
}

func (r *mockFieldReader) Raw(name string) (string, bool) {
	return r.String(name)
}

type mockArrayReader struct {
	data []byte
}

func (r *mockArrayReader) elements() [][]byte {
	if len(r.data) == 0 { return nil }
	return bytes.Split(r.data, []byte(","))
}

func (r *mockArrayReader) Len() int {
	return len(r.elements())
}

func (r *mockArrayReader) String(i int) string {
	return string(r.elements()[i])
}

func (r *mockArrayReader) Int(i int) int64 {
	c := GetConv()
	c.LoadBytes(r.elements()[i])
	v, _ := c.Int64()
	c.PutConv()
	return v
}

func (r *mockArrayReader) Float(i int) float64 {
	c := GetConv()
	c.LoadBytes(r.elements()[i])
	v, _ := c.Float64()
	c.PutConv()
	return v
}

func (r *mockArrayReader) Bool(i int) bool {
	return string(r.elements()[i]) == "true"
}

func (r *mockArrayReader) Bytes(i int) []byte {
	return r.elements()[i]
}

func (r *mockArrayReader) Object(i int, into Decodable) bool {
	val := r.elements()[i]
	if len(val) < 2 || val[0] != '{' || val[len(val)-1] != '}' {
		return false
	}
	innerReader := newMockFieldReader(val[1 : len(val)-1])
	return into.DecodeFields(innerReader) == nil
}

type sampleUser struct {
	Name string
	Age  int
	Tags []string
}

func (u *sampleUser) EncodeFields(w FieldWriter) {
	w.String("name", u.Name)
	w.Int("age", int64(u.Age))
	if len(u.Tags) > 0 {
		aw := w.Array("tags", len(u.Tags))
		for i := 0; i < len(u.Tags); i++ {
			aw.String(u.Tags[i])
		}
		aw.Close()
	}
}

func (u *sampleUser) DecodeFields(r FieldReader) error {
	if v, ok := r.String("name"); ok {
		u.Name = v
	}
	if v, ok := r.Int("age"); ok {
		u.Age = int(v)
	}
	if ar, ok := r.Array("tags"); ok {
		u.Tags = make([]string, ar.Len())
		for i := 0; i < ar.Len(); i++ {
			u.Tags[i] = ar.String(i)
		}
	}
	return nil
}

func TestCodecRoundTrip(t *testing.T) {
	user := &sampleUser{Name: "Alice", Age: 30, Tags: []string{"go", "wasm"}}
	buf := bytes.NewBuffer(nil)
	writer := newMockFieldWriter(buf)

	user.EncodeFields(writer)

	encoded := buf.Bytes()
	// Alice=name;30=age;[go,wasm]=tags; -- wait, I wrote name=Alice;age=30;
	// Let's re-verify mock writer output format: name=val;
	expected := "name=Alice;age=30;tags=[go,wasm];"
	if string(encoded) != expected {
		t.Errorf("expected %q, got %q", expected, string(encoded))
	}

	reader := newMockFieldReader(encoded)
	decodedUser := &sampleUser{}
	err := decodedUser.DecodeFields(reader)
	if err != nil {
		t.Fatal(err)
	}

	if decodedUser.Name != user.Name || decodedUser.Age != user.Age || len(decodedUser.Tags) != len(user.Tags) {
		t.Errorf("decoded user mismatch: %+v", decodedUser)
	}
}

func TestCodecAllocations(t *testing.T) {
	user := &sampleUser{Name: "Alice", Age: 30}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	writer := newMockFieldWriter(buf)

	allocs := testing.AllocsPerRun(100, func() {
		buf.Reset()
		user.EncodeFields(writer)
	})

	if allocs > 0 {
		t.Errorf("EncodeFields allocated %f times, want 0", allocs)
	}
}
