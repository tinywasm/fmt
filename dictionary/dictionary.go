package dictionary

import fmt "github.com/tinywasm/fmt"

// Built-in dictionary with EN (English) and ES (Spanish) translations.
// EN is the lookup key — used case-insensitively (e.g. Translate("empty") or Translate("Empty")).
// To add more languages, call RegisterWords() in your own init():
//
//	fmt.RegisterWords([]fmt.DictEntry{
//		{EN: "Empty", FR: "Vide", DE: "Leer"},
//	})
func init() {
	fmt.RegisterWords([]fmt.DictEntry{
		// A
		{EN: "All", ES: "Todo"},
		{EN: "Allowed", ES: "Permitido"},
		{EN: "Arrow", ES: "Flecha"},
		{EN: "Argument", ES: "Argumento"},
		{EN: "Assign", ES: "Asignar"},
		{EN: "Assignable", ES: "Asignable"},

		// B
		{EN: "Backing up", ES: "Respaldando"},
		{EN: "be", ES: "ser"},
		{EN: "Begin", ES: "Comenzar"},
		{EN: "Binary", ES: "Binario"},

		// C
		{EN: "Call", ES: "Llamar"},
		{EN: "Can", ES: "Puede"},
		{EN: "Cannot", ES: "No puede"},
		{EN: "Cancel", ES: "Cancelar"},
		{EN: "Changed", ES: "Cambiado"},
		{EN: "Character", ES: "Caracter"},
		{EN: "Chars", ES: "Caracteres"},
		{EN: "Checker", ES: "Verificador"},
		{EN: "Coding", ES: "Codificación"},
		{EN: "Compilation", ES: "Compilación"},
		{EN: "Configuration", ES: "Configuración"},
		{EN: "Connection", ES: "Conexión"},
		{EN: "Content", ES: "Contenido"},
		{EN: "Create", ES: "Crear"},

		// D
		{EN: "Date", ES: "Fecha"},
		{EN: "Debugging", ES: "Depuración"},
		{EN: "Decimal", ES: "Decimal"},
		{EN: "Delimiter", ES: "Delimitador"},
		{EN: "Dictionary", ES: "Diccionario"},
		{EN: "Digit", ES: "Dígito"},
		{EN: "Down", ES: "Abajo"},

		// E
		{EN: "Edit", ES: "Editar"},
		{EN: "Element", ES: "Elemento"},
		{EN: "Email", ES: "Correo electrónico"},
		{EN: "Empty", ES: "Vacío"},
		{EN: "End", ES: "Fin"},
		{EN: "Example", ES: "Ejemplo"},
		{EN: "Exceeds", ES: "Excede"},
		{EN: "Execute", ES: "Ejecutar"},

		// F
		{EN: "Failed", ES: "Falló"},
		{EN: "Female", ES: "Femenino"},
		{EN: "Field", ES: "Campo"},
		{EN: "Fields", ES: "Campos"},
		{EN: "Files", ES: "Archivos"},
		{EN: "Format", ES: "Formato"},
		{EN: "Found", ES: "Encontrado"},

		// H
		{EN: "Handler", ES: "Manejador"},
		{EN: "Hour", ES: "Hora"},
		{EN: "Hyphen", ES: "Guion"},

		// I
		{EN: "Icons", ES: "Iconos"},
		{EN: "Implemented", ES: "Implementado"},
		{EN: "in", ES: "en"},
		{EN: "Index", ES: "Índice"},
		{EN: "Information", ES: "Información"},
		{EN: "Input", ES: "Entrada"},
		{EN: "Insert", ES: "Insertar"},
		{EN: "Install", ES: "Instalar"},
		{EN: "Installation", ES: "Instalación"},
		{EN: "Invalid", ES: "Inválido"},

		// K
		{EN: "Keyboard", ES: "Teclado"},

		// L
		{EN: "Language", ES: "Idioma"},
		{EN: "Left", ES: "Izquierda"},
		{EN: "Letters", ES: "Letras"},
		{EN: "Line", ES: "Línea"},

		// M
		{EN: "Male", ES: "Masculino"},
		{EN: "Maximum", ES: "Máximo"},
		{EN: "Method", ES: "Método"},
		{EN: "Missing", ES: "Falta"},
		{EN: "Mismatch", ES: "Desajuste"},
		{EN: "Mode", ES: "Modo"},
		{EN: "Modes", ES: "Modos"},
		{EN: "More", ES: "Más"},
		{EN: "Move", ES: "Mover"},
		{EN: "Must", ES: "Debe"},

		// N
		{EN: "Negative", ES: "Negativo"},
		{EN: "Nil", ES: "Nulo"},
		{EN: "Non-numeric", ES: "No numérico"},
		{EN: "Not", ES: "No"},
		{EN: "Not of type", ES: "No es del tipo"},
		{EN: "Number", ES: "Número"},
		{EN: "Numbers", ES: "Números"},

		// O
		{EN: "of", ES: "de"},
		{EN: "Options", ES: "Opciones"},
		{EN: "Out", ES: "Fuera"},
		{EN: "Overflow", ES: "Desbordamiento"},

		// P
		{EN: "Page", ES: "Página"},
		{EN: "Point", ES: "Punto"},
		{EN: "Pointer", ES: "Puntero"},
		{EN: "Preparing", ES: "Preparando"},
		{EN: "Production", ES: "Producción"},
		{EN: "Provided", ES: "Proporcionado"},

		// Q
		{EN: "Quit", ES: "Salir"},

		// R
		{EN: "Range", ES: "Rango"},
		{EN: "Read", ES: "Leer"},
		{EN: "Required", ES: "Requerido"},
		{EN: "Right", ES: "Derecha"},
		{EN: "Round", ES: "Redondear"},

		// S
		{EN: "Seconds", ES: "Segundos"},
		{EN: "Session", ES: "Sesión"},
		{EN: "Shortcuts", ES: "Atajos"},
		{EN: "Slice", ES: "Segmento"},
		{EN: "Space", ES: "Espacio"},
		{EN: "Status", ES: "Estado"},
		{EN: "String", ES: "Cadena"},
		{EN: "Supported", ES: "Soportado"},
		{EN: "Switch", ES: "Cambiar"},
		{EN: "Switching", ES: "Cambiando"},
		{EN: "Sync", ES: "Sincronización"},
		{EN: "System", ES: "Sistema"},

		// T
		{EN: "Tab", ES: "Pestaña"},
		{EN: "Test", ES: "Prueba"},
		{EN: "Testing", ES: "Probando"},
		{EN: "Text", ES: "Texto"},
		{EN: "Time", ES: "Tiempo"},
		{EN: "to", ES: "a"},
		{EN: "Type", ES: "Tipo"},

		// U
		{EN: "Unexported", ES: "No Exportado"},
		{EN: "Unknown", ES: "Desconocido"},
		{EN: "Unsigned", ES: "Sin Signo"},
		{EN: "Up", ES: "Arriba"},
		{EN: "Use", ES: "Usar"},

		// V
		{EN: "Valid", ES: "Válido"},
		{EN: "Validating", ES: "Validando"},
		{EN: "Value", ES: "Valor"},
		{EN: "Visible", ES: "Visible"},

		// W
		{EN: "Warning", ES: "Advertencia"},
		{EN: "With", ES: "Con"},

		// Z
		{EN: "Zero", ES: "Cero"},
	})
}
