package fmt

import (
	"unsafe"
)

type Conv struct {
	// Buffers with initial capacity 64, grow as needed (no truncation)
	out     []byte // Buffer principal - make([]byte, 0, 64)
	outLen  int    // Longitud actual en out
	work    []byte // Buffer temporal - make([]byte, 0, 64)
	workLen int    // Longitud actual en work
	err     []byte // Buffer de errores - make([]byte, 0, 64)
	errLen  int    // Longitud actual en err
	// Type indicator - most frequently accessed	// Type indicator - most frequently accessed
	kind Kind // Hot path: type checking (private)

	// ✅ OPTIMIZED MEMORY ARCHITECTURE - unsafe.Pointer for complex types
	dataPtr unsafe.Pointer // Direct unsafe pointer to data (replaces ptrValue)
}

// Convert initializes a new Conv struct with optional value for string,bool and number manipulation.
// REFACTORED: Now accepts variadic parameters - Convert() or Convert(value)
// Phase 7: Uses object pool internally for memory optimization (transparent to user)
func Convert(v ...any) *Conv {
	c := GetConv()
	c.resetAllBuffers() // Asegurar que el objeto Conv esté completamente limpio
	// Validation: Only accept 0 or 1 parameter
	if len(v) > 1 {
		return c.wrErr("invalid", "number", "of", "argument")
	}
	// Initialize with value if provided, empty otherwise
	if len(v) == 1 {
		val := v[0]
		if val == nil {
			return c.wrErr("string", "empty")
		}

		// Special case: error type should return immediately with error state
		if _, isError := val.(error); isError {
			return c.wrErr(val.(error).Error())
		}

		// Use AnyToBuff for ALL other conversions - eliminates all duplication
		c.ResetBuffer(BuffOut)
		c.AnyToBuff(BuffOut, val)

		// AnyToBuff handles everything:
		// - Setting c.Kind and c.dataPtr for all types
		// - String pointer handling (*string)
		// - Complex types ([]string, map, etc.) with deferred conversion
		// - All numeric and boolean type conversions
		// - Error handling for unsupported types
	}
	// If no value provided, Conv is ready for builder pattern
	return c
}

// =============================================================================
// UNIVERSAL CONVERSION FUNCTION - REUSES EXISTING IMPLEMENTATIONS

// =============================================================================

// AnyToBuff converts any supported type to buffer using existing conversion logic
// REUSES: floatToOut, wrStringToOut, wrStringToErr
// Supports: string, int variants, uint variants, float variants, bool, []byte, LocStr
func (c *Conv) AnyToBuff(dest BuffDest, value any) {
	// Limpiar buffer de error antes de cualquier conversión inmediata
	c.ResetBuffer(BuffErr)

	switch v := value.(type) {
	// IMMEDIATE CONVERSION - Simple Types (ordered as in Kind.go)

	// K.Bool
	case bool:
		c.kind = K.Bool
		c.wrBool(dest, v)

	// K.Float32
	case float32:
		c.kind = K.Float32
		c.wrFloat32(dest, v)

	// K.Float64
	case float64:
		c.kind = K.Float64
		c.wrFloat64(dest, v)

	// K.Int
	case int:
		c.kind = K.Int
		c.wrIntBase(dest, int64(v), 10, true)

	// K.Int8
	case int8:
		c.kind = K.Int8
		c.wrIntBase(dest, int64(v), 10, true)

	// K.Int16
	case int16:
		c.kind = K.Int16
		c.wrIntBase(dest, int64(v), 10, true)

	// K.Int32
	case int32:
		c.kind = K.Int32
		c.wrIntBase(dest, int64(v), 10, true)

	// K.Int64
	case int64:
		c.kind = K.Int64
		c.wrIntBase(dest, v, 10, true)

	// K.Pointer - Only *string pointer supported
	case *string:
		// String pointer - verify not nil before dereferencing
		if v == nil {
			c.wrErr("string", "empty")
			return
		}
		// Store content relationship
		c.kind = K.Pointer            // Correctly set Kind to K.Pointer for *string
		c.dataPtr = unsafe.Pointer(v) // Store the pointer itself for Apply()
		c.WrString(dest, *v)

	// K.String
	case string:
		c.kind = K.String
		c.WrString(dest, v)

	// K.SliceStr - Special case for []string
	case []string:
		c.dataPtr = unsafe.Pointer(&v)
		c.kind = K.Slice

	// K.Uint
	case uint:
		c.kind = K.Uint
		c.wrIntBase(dest, int64(v), 10, false)

	// K.Uint8
	case uint8:
		c.kind = K.Uint8
		c.wrIntBase(dest, int64(v), 10, false)

	// K.Uint16
	case uint16:
		c.kind = K.Uint16
		c.wrIntBase(dest, int64(v), 10, false)

	// K.Uint32
	case uint32:
		c.kind = K.Uint32
		c.wrIntBase(dest, int64(v), 10, false)

	// K.Uint64
	case uint64:
		c.kind = K.Uint64
		c.wrIntBase(dest, int64(v), 10, false)

	// Special cases
	case error:
		c.WrString(dest, v.Error())

	default:
		c.anyToBuffFallback(dest, value)
	}
}

// GetKind returns the Kind of the value stored in the Conv
// This allows external packages to reuse tinystring's type detection logic
func (c *Conv) GetKind() Kind {
	return c.kind
}

// Apply updates the original string pointer with the current content and auto-releases to pool.
// This method should be used when you want to modify the original string directly
// without additional allocations.
func (t *Conv) Apply() {
	if t.kind == K.Pointer && t.dataPtr != nil {
		// Type assert to *string for Apply() functionality using unsafe pointer
		if strPtr := (*string)(t.dataPtr); strPtr != nil {
			*strPtr = t.GetString(BuffOut)
		}
	}
	// Auto-release back to pool for memory efficiency
	t.putConv()
}

// String method to return the content of the Conv and automatically returns object to pool
// Phase 7: Auto-release makes pool usage completely transparent to user
func (c *Conv) String() string {
	// If there's an error, return empty string (error available via StringErr())
	if c.hasContent(BuffErr) {
		c.putConv() // Auto-release back to pool for memory efficiency
		return ""
	}

	out := c.GetString(BuffOut)
	// Auto-release back to pool for memory efficiency
	c.putConv()
	return out
}

// Bytes returns the content of the Conv as a byte slice
func (c *Conv) Bytes() []byte {
	return c.getBytes(BuffOut)
}
