# fmt Dictionary Integration - Technical Design Document

## Executive Summary

This document outlines the integration of multilingual error message support into the fmt library, maintaining its core philosophy of minimal binary size while providing optional translation capabilities. The implementation follows a hybrid approach with zero dependencies and full TinyGo/WebAssembly compatibility.

## Problem Statement

Modern applications require internationalized error messages from the beginning of development. Currently, fmt provides excellent error handling but only in English. Users need:

1. Multilingual error messages for better user experience
2. Minimal impact on binary size (critical for WebAssembly)
3. Zero external dependencies
4. Full TinyGo compatibility
5. Backward compatibility with existing APIs

## Solution Architecture

### Core Design Principles

1. **Zero Dependencies**: No external imports, no reflection usage
2. **Minimal Binary Impact**: Target <15KB for complete translation system
3. **Backward Compatibility**: Existing API remains unchanged
4. **Optional Usage**: Translations only included when explicitly used
5. **Performance First**: Efficient array-based lookups

### Type System

#### Language Enumeration
```go
type lang uint8
const (
    EN lang = iota // 0 - English (default)
    ES             // 1 - Spanish
    PT             // 2 - Portuguese  
    FR             // 3 - French
    RU             // 4 - Russian
    DE             // 5 - German
    IT             // 6 - Italian
    HI             // 7 - Hindi
    BN             // 8 - Bengali
    ID             // 9 - Indonesian
    AR             // 10 - Arabic
    UR             // 11 - Urdu
    ZH             // 12 - Chinese
)
```

#### Translation Container
```go
// LocStr represents localized string Language translations using fixed array for efficiency
type LocStr [13]string // Fixed array: EN, ES, PT, FR, RU, DE, IT, HI, BN, ID, AR, UR, ZH

// get returns translation for specified language with English fallback
func (o LocStr) get(l lang) string {
    if int(l) < len(o) && o[l] != "" {
        return o[l]
    }
    return o[EN] // Fallback to English
}
```

### Dictionary Structure

Based on analysis of current fmt error messages, the dictionary contains reusable words in alphabetical order that can be combined to form any error message:

```go
type dictionary struct {
    // Basic words sorted alphabetically for maximum reusability
    Argument    LocStr // "argument" 
    Base        LocStr // "base"
    Boolean     LocStr // "boolean"
    Cannot      LocStr // "cannot"
    Empty       LocStr // "empty"
    End         LocStr // "end"
    Float       LocStr // "float"
    Fmt      LocStr // "format"
    Integer     LocStr // "integer"
    Invalid     LocStr // "invalid"
    Missing     LocStr // "missing"
    Negative    LocStr // "negative"
    NonNumeric  LocStr // "non-numeric"
    Not         LocStr // "not"
    Number      LocStr // "number"
    Numbers     LocStr // "numbers"
    Overflow    LocStr // "overflow"
    Round       LocStr // "round"
    Specifier   LocStr // "specifier"
    String      LocStr // "string"
    Supported   LocStr // "supported"
    Type        LocStr // "type"
    Unsigned    LocStr // "unsigned"
    Unsupported LocStr // "unsupported"
    Value       LocStr // "value"
    Wrong       LocStr // "wrong"
    
    // Additional common terms for user extensions
    At          LocStr // "at"
    For         LocStr // "for"
    In          LocStr // "in"
    Of          LocStr // "of"
    Range       LocStr // "range"
    Required    LocStr // "required"
    Text        LocStr // "text"
    Unknown     LocStr // "unknown"
}

// Global dictionary instance
var D dictionary
```

#### Error Message Composition Examples
With this word-based approach, current error messages can be composed as:

**⚠️ CRITICAL: Spanish-First Word Order Convention**
- **ERROR MESSAGE COMPOSITION MUST ALWAYS BE THOUGHT IN SPANISH FIRST**
- The word order should be semantically correct in Spanish, so it sounds natural in both Spanish and English
- **Rule**: Compose error messages thinking "How would this sound in Spanish?" first
- **Example**: `D.Format, D.Invalid` → "formato inválido" (natural Spanish) vs `D.Invalid, D.Format` → "inválido formato" (unnatural Spanish)
- **Practice**: Always prioritize Spanish semantic correctness while ensuring English remains understandable
- **Why**: Spanish adjective placement often differs from English, but following Spanish order usually makes both languages sound natural

```go
// errEmptyString = "empty string" → "cadena vacía" (Spanish order: noun + adjective)
Err(D.String, D.Empty)

// errNegativeUnsigned = "negative numbers are not supported for unsigned integers"
// "números negativos no soportados para enteros sin signo" (Spanish order)  
Err(D.Numbers, D.Negative, D.Not, D.Supported, D.To, D.Integer, D.Unsigned)

// errInvalidBase = "invalid base" → "base inválida" (Spanish order: noun + adjective)
Err("Base", D.Invalid)

// errOverflow = "number overflow" → "desbordamiento de número" (Spanish order)
Err(D.Overflow, D.Of, D.Number)

// errInvalidFormat = "invalid format" → "formato inválido" (Spanish order: noun + adjective)
Err(D.Format, D.Invalid)

// errFormatMissingArg = "missing argument" → "argumento faltante" (Spanish order: noun + adjective)
Err(D.Argument, D.Missing)

// errFormatWrongType = "wrong argument type" → "tipo de argumento incorrecto" (Spanish order)
Err(D.Type, D.Of, D.Argument, D.Wrong)

// errCannotRound = "cannot round non-numeric value" → "no puede redondear valor no numérico" (Spanish order)
Err(D.Cannot, D.Round, D.Value, D.NonNumeric)
```

// errFormatWrongType = "wrong argument type" → "tipo de argumento incorrecto" (Spanish order)
Err(D.Type, D.Of, D.Argument, D.Wrong)

// errCannotRound = "cannot round non-numeric value" → "no puede redondear valor no numérico" (Spanish order)
Err(D.Cannot, D.Round, D.Value, D.NonNumeric)
```

### Language Configuration

#### Default Language Management
```go
// Private global configuration
var defLang lang = EN

// OutLang sets the default output language
// OutLang() without parameters auto-detects system language
// OutLang(ES) sets Spanish as default
func OutLang(l ...lang) {
    if len(l) == 0 {
        defLang = getSystemLang()
    } else {
        defLang = l[0]
    }
}
```

#### System Language Detection
```go
// env.back.go - Backend (non-WASM) implementation
//go:build !wasm

func getSystemLang() lang {
    // Environment variable parsing logic
    // Returns lang enum based on LANG, LANGUAGE, LC_ALL, LC_MESSAGES
    // Fallback to EN if detection fails
}

// env.front.go - Frontend (WASM) implementation  
//go:build wasm

func getSystemLang() lang {
    // Browser language detection logic
    // Uses navigator.language via syscall/js
    // Fallback to EN if detection fails
}
```

### API Integration

#### Existing API Preservation
```go
// These functions remain unchanged - 100% backward compatibility
func Err(args ...any) *Conv
func Errorf(format string, args ...any) *Conv
func (c *Conv) NewErr(values ...any) *Conv
```

#### Translation Detection
The key innovation is automatic detection of LocStr types in existing functions:

```go
func (c *Conv) NewErr(values ...any) *Conv {
    var sep, out string
    c.tmpStr = ""
    c.err = ""
    
    // Determine target language
    targetLang := defLang
    
    for _, v := range values {
        switch val := v.(type) {
        case lang:
            // Language specified inline
            targetLang = val
            continue        case LocStr:
            // Translation type detected - use translation
            if val.get(targetLang) != "" {
                out += sep + val.get(targetLang)
            } else {
                out += sep + val.get(EN)
            }
        default:
            // Handle other types normally
            c.any2s(v)
            if c.err != "" {
                out += sep + c.err
            }
            if c.tmpStr != "" {
                out += sep + c.tmpStr
            }
        }
        
        if c.err != "" || c.tmpStr != "" || (v != nil) {
            sep = " "
        }
    }
    
    c.err = out
    c.Kind = K.Err
    return c
}
```

### Usage Examples

#### Basic Usage (Backward Compatible)
```go
// Existing code works unchanged
err := tinystring.Err("invalid input").Error()
// Output: "invalid input"
```

#### Translated Errors
```go
import . "webtyp.com/fmt"

// Configure default language
OutLang(ES) // Spanish

// Use word combinations to create error messages
err := Err(D.Invalid, D.Format, "value").Error()
// Output: "inválido formato value"

// Complex error message composition
err := Err(D.Negative, D.Numbers, D.Not, D.Supported).Error()
// Output: "negativo números no soportado"

// Mix languages inline
err := Err(FR, D.Empty, D.String).Error()  
// Output: "vide chaîne" (French)

// Auto-detect system language
OutLang() // Detects browser/OS language
err := Err(D.Cannot, D.Round, D.NonNumeric, D.Value).Error()
// Output in user's system language
```

#### User Extensions
```go
// Users can extend with their own words and combine with dictionary
type MyDict struct {
    User     LocStr
    Email    LocStr
    Password LocStr
    Login    LocStr
}

var MD = MyDict{
    // Language order: EN, ES, PT, FR, RU, DE, IT, HI, BN, ID, AR, UR, ZH
    User:     LocStr{"user", "usuario", "usuário", "utilisateur", "пользователь", "Benutzer", "utente", "उपयोगकर्ता", "ব্যবহারকারী", "pengguna", "مستخدم", "صارف", "用户"},
    Email:    LocStr{"email", "correo", "email", "courriel", "электронная почта", "E-Mail", "email", "ईमेल", "ইমেইল", "email", "بريد إلكتروني", "ای میل", "邮箱"},
    Password: LocStr{"password", "contraseña", "senha", "mot de passe", "пароль", "Passwort", "password", "पासवर्ड", "পাসওয়ার্ড", "kata sandi", "كلمة مرور", "پاس ورڈ", "密码"},
    Login:    LocStr{"login", "iniciar sesión", "login", "connexion", "вход", "Anmeldung", "accesso", "लॉगिन", "লগইন", "masuk", "تسجيل الدخول", "لاگ ان", "登录"},
}

// Combine system dictionary with user extensions
err := Err(MD.User, D.Not, D.Found).Error()
// Output: "usuario no encontrado" (Spanish)

err := Err(D.Invalid, MD.Email, D.Format).Error()  
// Output: "inválido correo formato" (Spanish)
```

## Implementation Plan

### Phase 1: Core Infrastructure

#### File Structure
```
tinystring/
├── dictionary.go      # NEW: LocStr type, lang enum, dictionary struct and D instance
├── env.back.go       # NEW: System language detection (non-WASM)
├── env.front.go      # NEW: Browser language detection (WASM)
├── error.go          # MODIFIED: Remove errorType, update NewErr to handle LocStr
├── convert.go        # MODIFIED: Change err field from errorType to string
└── ...existing files
```

#### Implementation Steps
1. **Create `dictionary.go`**:
   - Define `lang` enum with 13 languages
   - Define `LocStr` type as `[13]string` array
   - Implement `(o LocStr) get(l lang) string` method
   - Define `dictionary` struct with all error terms
   - Initialize global `D` dictionary instance with translations
   - Implement `OutLang(l ...lang)` function
   - Add private `defLang` variable

2. **Create `env.back.go` and `env.front.go`**:
   - Backend: Parse environment variables (LANG, LANGUAGE, etc.)
   - Frontend: Use `syscall/js` to access `navigator.language`
   - Both return `lang` enum, fallback to `EN`

3. **Update `error.go`**:
   - Remove `errorType` type completely
   - Change all `errorType` references to `string`
   - Update `NewErr()` to detect and handle `LocStr` and `lang` types
   - Maintain backward compatibility with existing error constants

4. **Update `convert.go`**:
   - Change `err` field from `errorType` to `string` (line 45)

### Phase 2: Dictionary Population

#### Translation Strategy

**Dictionary Fmt Convention:**
- All dictionary definitions in `dictionary.go` must use horizontal format for maximum compactness
- All translations for a word are placed on a single line
- Language order follows the enum: EN, ES, PT, FR, RU, DE, IT, HI, BN, ID, AR, UR, ZH
- Language codes are specified in comments above the translations for reference
- This format maximizes readability and makes translation updates easier

```go
// Dictionary location: dictionary.go
var D = dictionary{
    // Basic words sorted alphabetically - full translations for maximum reusability
    // Language order: EN, ES, PT, FR, RU, DE, IT, HI, BN, ID, AR, UR, ZH
    Argument: LocStr{"argument", "argumento", "argumento", "argument", "аргумент", "Argument", "argomento", "तर्क", "যুক্তি", "argumen", "وسيط", "دلیل", "参数"},    
    Base: LocStr{"base", "base", "base", "base", "основание", "Basis", "base", "आधार", "ভিত্তি", "basis", "قاعدة", "بنیاد", "进制"},
    
    Empty: LocStr{"empty", "vacío", "vazio", "vide", "пустой", "leer", "vuoto", "खाली", "খালি", "kosong", "فارغ", "خالی", "空"},
    
    Invalid: LocStr{"invalid", "inválido", "inválido", "invalide", "недопустимый", "ungültig", "non valido", "अमान्य", "অবৈধ", "tidak valid", "غير صالح", "غیر درست", "无效"},
    
    // ... continue for all ~30 words
}
```

#### Error Mapping Strategy
Instead of direct mapping, errors are now composed from words with **Spanish-first word order**:

**Current Constants** → **Word Composition (Spanish Order)**
1. `errEmptyString` → `D.String + D.Empty` ("cadena vacía")
2. `errNegativeUnsigned` → `D.Numbers + D.Negative + D.Not + D.Supported + D.To + D.Integer + D.Unsigned` ("números negativos no soportados para enteros sin signo")
3. `errInvalidBase` → `"Base" + D.Invalid` ("base inválida")
4. `errOverflow` → `D.Overflow + D.Of + D.Number` ("desbordamiento de número")
5. `errInvalidFormat` → `D.Format + D.Invalid` ("formato inválido")
6. `errFormatMissingArg` → `D.Argument + D.Missing` ("argumento faltante")
7. `errFormatWrongType` → `D.Type + D.Of + D.Argument + D.Wrong` ("tipo de argumento incorrecto")
8. `errFormatUnsupported` → `D.Specifier + D.Of + D.Format + D.Unsupported` ("especificador de formato no soportado")
9. `errIncompleteFormat` → `D.Specifier + D.Of + D.Format + D.Invalid + D.At + D.End` ("especificador de formato inválido al final")
10. `errCannotRound` → `D.Cannot + D.Round + D.Value + D.NonNumeric` ("no puede redondear valor no numérico")
11. `errCannotFormat` → `D.Cannot + D.Format + D.Value + D.NonNumeric` ("no puede formatear valor no numérico")
12. `errInvalidFloat` → `D.String + D.Of + D.Float + D.Invalid` ("cadena de flotante inválida")
13. `errInvalidBool` → `D.Value + "Bool" + D.Invalid` ("valor booleano inválido")

### Phase 3: Testing & Validation

#### Test Files Structure
```
tinystring/
├── dictionary_test.go    # NEW: Dictionary functionality tests
├── env_test.go          # NEW: Language detection tests  
├── error_test.go        # MODIFIED: Add concurrency tests for translations
└── integration_test.go  # NEW: Full integration tests
```

#### Test Coverage
1. **Dictionary Tests** (`dictionary_test.go`):
   - Test all 13 languages for each dictionary entry
   - Verify fallback to English when translation missing
   - Test `OutLang()` with and without parameters
   - Benchmark translation performance vs string literals

2. **Environment Tests** (`env_test.go`):
   - Mock environment variables for backend detection
   - Mock browser language for frontend detection
   - Test fallback behavior when detection fails

3. **Error Concurrency Tests** (`error_test.go`):
   - Test concurrent access to `defLang` global variable
   - Test race conditions when changing default language
   - Verify thread safety of dictionary access

4. **Integration Tests** (`integration_test.go`):
   - Test backward compatibility with existing error usage
   - Test mixed usage (LocStr types + regular strings)
   - Test inline language specification
   - Validate binary size impact (<15KB)
   - TinyGo compilation tests

### Phase 4: Documentation
1. Update README with translation examples
2. Create migration guide for existing users
3. Document performance characteristics
4. Add troubleshooting guide

## Detailed File Specifications

### dictionary.go
```go
package tinystring

// Language enumeration for supported languages
type lang uint8

const (
    EN lang = iota // 0 - English (default)
    ES             // 1 - Spanish
    PT             // 2 - Portuguese  
    FR             // 3 - French
    RU             // 4 - Russian
    DE             // 5 - German
    IT             // 6 - Italian
    HI             // 7 - Hindi
    BN             // 8 - Bengali
    ID             // 9 - Indonesian
    AR             // 10 - Arabic
    UR             // 11 - Urdu
    ZH             // 12 - Chinese
)

// LocStr represents Output Language translations using fixed array for efficiency
type LocStr [13]string

// get returns translation for specified language with English fallback
func (o LocStr) get(l lang) string {
    if int(l) < len(o) && o[l] != "" {
        return o[l]
    }
    return o[EN] // Fallback to English
}

// Dictionary structure containing all translatable terms
type dictionary struct {
    // Basic words sorted alphabetically for maximum reusability
    Argument    LocStr // "argument" 
    At          LocStr // "at"
    Base        LocStr // "base"
    Boolean     LocStr // "boolean"
    Cannot      LocStr // "cannot"
    Empty       LocStr // "empty"
    End         LocStr // "end"
    Float       LocStr // "float"
    For         LocStr // "for"
    Fmt      LocStr // "format"
    Found       LocStr // "found"
    In          LocStr // "in"
    Integer     LocStr // "integer"
    Invalid     LocStr // "invalid"
    Missing     LocStr // "missing"
    Negative    LocStr // "negative"
    NonNumeric  LocStr // "non-numeric"
    Not         LocStr // "not"
    Number      LocStr // "number"
    Numbers     LocStr // "numbers"
    Of          LocStr // "of"
    Overflow    LocStr // "overflow"
    Range       LocStr // "range"
    Required    LocStr // "required"
    Round       LocStr // "round"
    Specifier   LocStr // "specifier"
    String      LocStr // "string"
    Supported   LocStr // "supported"
    Text        LocStr // "text"
    Type        LocStr // "type"
    Unknown     LocStr // "unknown"
    Unsigned    LocStr // "unsigned"
    Unsupported LocStr // "unsupported"
    Value       LocStr // "value"
    Wrong       LocStr // "wrong"
}

// Global dictionary instance - populated with all translations using horizontal format
var D = dictionary{
    // Language order: EN, ES, PT, FR, RU, DE, IT, HI, BN, ID, AR, UR, ZH
    Argument:    LocStr{"argument", "argumento", "argumento", "argument", "аргумент", "Argument", "argomento", "तर्क", "যুক্তি", "argumen", "وسيط", "دلیل", "参数"},
    At:          LocStr{"at", "en", "em", "à", "в", "bei", "a", "पर", "এ", "di", "في", "میں", "在"},
    Base:        LocStr{"base", "base", "base", "base", "основание", "Basis", "base", "आधार", "ভিত্তি", "basis", "قاعدة", "بنیاد", "进制"},
    Boolean:     LocStr{"boolean", "booleano", "booleano", "booléen", "логический", "boolescher", "booleano", "बूलियन", "বুলিয়ান", "boolean", "منطقي", "بولین", "布尔"},
    Cannot:      LocStr{"cannot", "no puede", "não pode", "ne peut pas", "не может", "kann nicht", "non può", "नहीं कर सकते", "পারে না", "tidak bisa", "لا يمكن", "نہیں کر سکتے", "不能"},
    Empty:       LocStr{"empty", "vacío", "vazio", "vide", "пустой", "leer", "vuoto", "खाली", "খালি", "kosong", "فارغ", "خالی", "空"},
    End:         LocStr{"end", "fin", "fim", "fin", "конец", "Ende", "fine", "अंत", "শেষ", "akhir", "نهاية", "اختتام", "结束"},
    Float:       LocStr{"float", "flotante", "flutuante", "flottant", "число с плавающей точкой", "Gleitkomma", "virgola mobile", "फ्लोट", "ফ্লোট", "float", "عائم", "فلوٹ", "浮点"},
    For:         LocStr{"for", "para", "para", "pour", "для", "für", "per", "के लिए", "জন্য", "untuk", "لـ", "کے لیے", "为"},
    Fmt:      LocStr{"format", "formato", "formato", "format", "формат", "Fmt", "formato", "प्रारूप", "বিন্যাস", "format", "تنسيق", "فارمیٹ", "格式"},
    Found:       LocStr{"found", "encontrado", "encontrado", "trouvé", "найден", "gefunden", "trovato", "मिला", "পাওয়া", "ditemukan", "موجود", "ملا", "找到"},
    In:          LocStr{"in", "en", "em", "dans", "в", "in", "in", "में", "এ", "dalam", "في", "میں", "在"},
    Integer:     LocStr{"integer", "entero", "inteiro", "entier", "целое число", "ganze Zahl", "intero", "पूर्णांक", "পূর্ণসংখ্যা", "integer", "عدد صحيح", "انٹیجر", "整数"},
    Invalid:     LocStr{"invalid", "inválido", "inválido", "invalide", "недопустимый", "ungültig", "non valido", "अमान्य", "অবৈধ", "tidak valid", "غير صالح", "غیر درست", "无效"},
    Missing:     LocStr{"missing", "falta", "ausente", "manquant", "отсутствует", "fehlend", "mancante", "गुम", "অনুপস্থিত", "hilang", "مفقود", "غائب", "缺少"},
    Negative:    LocStr{"negative", "negativo", "negativo", "négatif", "отрицательный", "negativ", "negativo", "नकारात्मक", "নেগেটিভ", "negatif", "سالب", "منفی", "负"},
    NonNumeric:  LocStr{"non-numeric", "no numérico", "não numérico", "non numérique", "нечисловой", "nicht numerisch", "non numerico", "गैर-संख्यात्मक", "অ-সংখ্যাসূচক", "non-numerik", "غير رقمي", "غیر عددی", "非数字"},
    Not:         LocStr{"not", "no", "não", "pas", "не", "nicht", "non", "नहीं", "না", "tidak", "ليس", "نہیں", "不"},
    Number:      LocStr{"number", "número", "número", "nombre", "число", "Zahl", "numero", "संख्या", "সংখ্যা", "angka", "رقم", "نمبر", "数字"},
    Numbers:     LocStr{"numbers", "números", "números", "nombres", "числа", "Zahlen", "numeri", "संख्याएं", "সংখ্যা", "angka", "أرقام", "نمبرز", "数字"},
    Of:          LocStr{"of", "de", "de", "de", "из", "von", "di", "का", "এর", "dari", "من", "کا", "的"},
    Overflow:    LocStr{"overflow", "desbordamiento", "estouro", "débordement", "переполнение", "Überlauf", "overflow", "ओवरफ्लो", "ওভারফ্লো", "overflow", "فيض", "اوور فلو", "溢出"},
    Range:       LocStr{"range", "rango", "intervalo", "plage", "диапазон", "Bereich", "intervallo", "रेंज", "পরিসর", "rentang", "نطاق", "رینج", "范围"},
    Required:    LocStr{"required", "requerido", "necessário", "requis", "обязательный", "erforderlich", "richiesto", "आवश्यक", "প্রয়োজনীয়", "diperlukan", "مطلوب", "ضروری", "必需"},
    Round:       LocStr{"round", "redondear", "arredondar", "arrondir", "округлить", "runden", "arrotondare", "गोल", "গোল", "bulatkan", "جولة", "گول", "圆"},
    Specifier:   LocStr{"specifier", "especificador", "especificador", "spécificateur", "спецификатор", "Spezifizierer", "specificatore", "निर्दिष्टकर्ता", "নির্দিষ্টকারী", "penentu", "محدد", "تعین کنندہ", "说明符"},
    String:      LocStr{"string", "cadena", "string", "chaîne", "строка", "Zeichenkette", "stringa", "स्ट्रिंग", "স্ট্রিং", "string", "سلسلة", "سٹرنگ", "字符串"},
    Supported:   LocStr{"supported", "soportado", "suportado", "pris en charge", "поддерживается", "unterstützt", "supportato", "समर्थित", "সমর্থিত", "didukung", "مدعوم", "معاون", "支持"},
    Text:        LocStr{"text", "texto", "texto", "texte", "текст", "Text", "testo", "पाठ", "পাঠ", "teks", "نص", "متن", "文本"},
    Type:        LocStr{"type", "tipo", "tipo", "type", "тип", "Typ", "tipo", "प्रकार", "টাইপ", "tipe", "نوع", "قسم", "类型"},
    Unknown:     LocStr{"unknown", "desconocido", "desconhecido", "inconnu", "неизвестный", "unbekannt", "sconosciuto", "अज्ञात", "অজানা", "tidak diketahui", "غير معروف", "نامعلوم", "未知"},
    Unsigned:    LocStr{"unsigned", "sin signo", "sem sinal", "non signé", "беззнаковый", "vorzeichenlos", "senza segno", "अहस्ताक्षरित", "স্বাক্ষরহীন", "tidak bertanda", "غير موقع", "غیر دستخط شدہ", "无符号"},
    Unsupported: LocStr{"unsupported", "no soportado", "não suportado", "non pris en charge", "не поддерживается", "nicht unterstützt", "non supportato", "असमर्थित", "অসমর্থিত", "tidak didukung", "غير مدعوم", "غیر معاون", "不支持"},
    Value:       LocStr{"value", "valor", "valor", "valeur", "значение", "Wert", "valore", "मूल्य", "মান", "nilai", "قيمة", "قیمت", "值"},
    Wrong:       LocStr{"wrong", "incorrecto", "errado", "mauvais", "неправильный", "falsch", "sbagliato", "गलत", "ভুল", "salah", "خطأ", "غلط", "错误"},
}

// Private global configuration
var defLang lang = EN

// OutLang sets the default output language
// OutLang() without parameters auto-detects system language
// OutLang(ES) sets Spanish as default
func OutLang(l ...lang) {
    if len(l) == 0 {
        defLang = getSystemLang()
    } else {
        defLang = l[0]
    }
}
```

### env.back.go
```go
//go:build !wasm

package tinystring

import (
    "os"
    "strings"
)

// getSystemLang detects system language from environment variables
func getSystemLang() lang {
    // Common environment variables that contain language information
    langVars := []string{"LANG", "LANGUAGE", "LC_ALL", "LC_MESSAGES"}
    
    for _, envVar := range langVars {
        if envValue := os.Getenv(envVar); envValue != "" {
            // Parse language code from environment variable
            code := strings.Split(envValue, ".")[0] // Remove encoding part
            code = strings.Split(code, "_")[0]      // Get language part
            code = strings.Split(code, "-")[0]      // Handle dash format
            code = strings.ToLower(code)
            
            // Map to lang enum
            switch code {
            case "es": return ES
            case "pt": return PT
            case "fr": return FR
            case "ru": return RU
            case "de": return DE
            case "it": return IT
            case "hi": return HI
            case "bn": return BN
            case "id": return ID
            case "ar": return AR
            case "ur": return UR
            case "zh": return ZH
            default: return EN
            }
        }
    }
    
    return EN // Default fallback
}
```

### env.front.go
```go
//go:build wasm

package tinystring

import (
    "strings"
    "syscall/js"
)

// getSystemLang detects browser language from navigator.language
func getSystemLang() lang {
    // Get browser language
    navigator := js.Global().Get("navigator")
    if navigator.IsUndefined() {
        return EN
    }
    
    language := navigator.Get("language")
    if language.IsUndefined() {
        return EN
    }
    
    langCode := language.String()
    if langCode == "" {
        return EN
    }
    
    // Parse language code (e.g., "es-ES" -> "es")
    code := strings.Split(langCode, "-")[0]
    code = strings.ToLower(code)
    
    // Map to lang enum
    switch code {
    case "es": return ES
    case "pt": return PT
    case "fr": return FR
    case "ru": return RU
    case "de": return DE
    case "it": return IT
    case "hi": return HI
    case "bn": return BN
    case "id": return ID
    case "ar": return AR
    case "ur": return UR
    case "zh": return ZH
    default: return EN
    }
}
```

### Modified error.go Structure
```go
package tinystring

// Remove: type errorType string
// Remove: All errorType constants

// Error message constants - keep for backward compatibility
const (
    errNone              = ""
    errEmptyString       = "empty string"
    errNegativeUnsigned  = "negative numbers are not supported for unsigned integers"
    // ... rest remain as string constants
)

// Modified NewErr to handle LocStr types
func (c *Conv) NewErr(values ...any) *Conv {
    var sep, out string
    c.tmpStr = ""
    c.err = ""
    
    // Determine target language
    targetLang := defLang
    
    for _, v := range values {
        switch val := v.(type) {
        case lang:
            // Language specified inline
            targetLang = val
            continue
        case LocStr:
            // Translation type detected - use translation
            out += sep + val.get(targetLang)
        default:
            // Handle other types normally (unchanged logic)
            c.any2s(v)
            if c.err != "" {
                out += sep + c.err
            }
            if c.tmpStr != "" {
                out += sep + c.tmpStr
            }
        }
        
        if c.err != "" || c.tmpStr != "" || (v != nil) {
            sep = " "
        }
    }
    
    c.err = out
    c.Kind = K.Err
    return c
}
```

### Modified convert.go Structure
```go
type Conv struct {
    // ...existing fields...
    err string // Changed from: err errorType
    // ...rest unchanged...
}
```

### Memory Layout
- **LocStr type**: 13 * 8 bytes (string headers) = 104 bytes per entry
- **Dictionary**: ~35 words * 104 bytes = ~3.6KB base overhead
- **Translation strings**: ~8KB for all languages (short words are efficient)
- **Total impact**: ~12KB additional binary size (within 15KB limit)

### Performance Characteristics
- **Translation lookup**: O(1) array access
- **Language detection**: Cached result, ~μs overhead
- **Memory allocation**: Zero additional allocations
- **Concurrency**: Thread-safe reads, atomic writes for config

### Build Targets
- **Standard Go**: Full compatibility
- **TinyGo WASM**: Primary target, full support
- **TinyGo Embedded**: Conditional compilation support
- **All platforms**: Windows, Linux, macOS, WASM

## Risk Assessment

### ToLower Risk
- Binary size increase (well within 15KB limit)
- Performance impact (negligible with array access)
- Backward compatibility (100% preserved)

### Medium Risk
- Translation accuracy (mitigated by native speaker review)
- System language detection (fallback to English)

### High Risk
- None identified

## Success Metrics

1. **Binary Size**: <15KB additional size
2. **Performance**: <1% overhead on error creation
3. **Compatibility**: 100% backward compatibility
4. **Coverage**: All existing error messages translated
5. **Usability**: Simple API requiring minimal code changes

## Migration Strategy

### For Existing Users
- No action required - code continues to work
- Optional: Add `OutLang()` calls for internationalization
- Optional: Replace string literals with `D.*` references

### For New Users
- Use `import . "webtyp.com/fmt"` idiom
- Configure language with `OutLang()`
- Use dictionary entries for consistent translations

## Conclusion

This design provides a robust, efficient, and user-friendly internationalization system for fmt while maintaining its core principles of minimal size and zero dependencies. The hybrid approach allows users to opt into translation features without impacting those who don't need them.

The implementation preserves 100% backward compatibility while providing powerful new capabilities for multilingual applications. The fixed-array approach ensures predictable performance and minimal memory overhead, making it suitable for resource-constrained environments including WebAssembly and embedded systems.

---

**Document Version**: 1.0  
**Date**: June 17, 2025  
**Author**: fmt Development Team  
**Status**: Implementation Ready

## Technical Specifications

### File Organization
- **`dictionary.go`**: Core dictionary, types, and configuration
- **`env.back.go`**: Backend language detection (build tag: `!wasm`)
- **`env.front.go`**: Frontend language detection (build tag: `wasm`)
- **`error.go`**: Modified to support LocStr types, remove errorType
- **`convert.go`**: Change err field type to string
