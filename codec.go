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
	// Raw emite el valor directamente sin escapar.
	Raw(name, val string)
	// Anidado: objeto hijo que también es Encodable.
	Object(name string, val Encodable)
	// Arrays tipados sin []any: retorna ArrayWriter directo para evitar closures.
	Array(name string, n int) ArrayWriter
}

// ArrayWriter empuja elementos tipados de un array (sin []any).
type ArrayWriter interface {
	String(val string)
	Int(val int64)
	Float(val float64)
	Bool(val bool)
	Bytes(val []byte)
	Object(val Encodable)
	// Close finaliza el array.
	Close()
}

// Encodable: un valor que sabe escribir SUS campos (lo genera ormc).
type Encodable interface {
	EncodeFields(w FieldWriter)
	IsNil() bool
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
	// Raw devuelve el valor crudo asociado al nombre.
	Raw(name string) (string, bool)
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
	DecodeFields(r FieldReader)
	IsNil() bool
}

// IsNil comprueba de forma segura y portable si un valor es nil o
// si es un puntero nil dentro de una interfaz Encodable o Decodable.
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	if enc, ok := v.(Encodable); ok {
		return enc.IsNil()
	}
	if dec, ok := v.(Decodable); ok {
		return dec.IsNil()
	}
	return false
}
