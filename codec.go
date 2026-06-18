package fmt

// FieldWriter recibe los campos de un valor por llamadas TIPADAS.
// Implementaciones: jsvalue (escribe js.Value), json (escribe bytes), etc.
// Reglas: cero `any`, cero `map`, cero asignación en el heap Go (el writer reusa su buffer).
type FieldWriter interface {
	String(name, val string)
	Int(name string, val int64)
	Uint(name string, val uint64)
	Float(name string, val float64)
	Bool(name string, val bool)
	Bytes(name string, val []byte)
	Null(name string)
	// Anidado: objeto hijo que también es Encodable.
	Object(name string, val Encodable)
	// Arrays tipados sin []any: el writer abre el array y el valor empuja elementos.
	Array(name string, n int, each func(i int, a ArrayWriter))
}

// ArrayWriter empuja elementos tipados de un array (sin []any).
type ArrayWriter interface {
	String(val string)
	Int(val int64)
	Float(val float64)
	Bool(val bool)
	Bytes(val []byte)
	Object(val Encodable)
}

// Encodable: un valor que sabe escribir SUS campos (lo genera ormc).
type Encodable interface {
	EncodeFields(w FieldWriter)
}

// FieldReader entrega los campos por nombre, TIPADOS. El bool indica presencia.
// Lee por nombre directo (NO construye un map).
type FieldReader interface {
	String(name string) (string, bool)
	Int(name string) (int64, bool)
	Uint(name string) (uint64, bool)
	Float(name string) (float64, bool)
	Bool(name string) (bool, bool)
	Bytes(name string) ([]byte, bool)
	Object(name string, into Decodable) bool
	Array(name string) (ArrayReader, bool)
}

// ArrayReader recorre un array tipado.
type ArrayReader interface {
	Len() int
	String(i int) string
	Int(i int) int64
	Float(i int) float64
	Bool(i int) bool
	Bytes(i int) []byte
	Object(i int, into Decodable) bool
}

// Decodable: un valor que sabe leer SUS campos (lo genera ormc).
type Decodable interface {
	DecodeFields(r FieldReader) error
}
