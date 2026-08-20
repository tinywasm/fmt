package fmt

// Sprint convierte v en texto para mostrarlo.
//
// El caso error se trata aquí y no en Convert: Convert mete el mensaje en el
// buffer de error a propósito (String() devuelve "" y el texto se recupera con
// StringErr()), un contrato que otros llamadores usan. Sprint tiene el trabajo
// contrario —imprimir lo que le den—, y lo que más le llega es un error.
//
// Example: Sprint(42) returns "42"
// Example: Sprint(true) returns "true"
// Example: Sprint("hello") returns "hello"
// Example: Sprint(errors.New("boom")) returns "boom"
func Sprint(v any) string {
	if err, ok := v.(error); ok && err != nil {
		return err.Error()
	}
	return Convert(v).String()
}