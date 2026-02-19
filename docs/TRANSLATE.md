# 🌍 fmt: Multilingual Message System

**fmt** is a lightweight, dependency-free translation engine for generating composable messages. It uses a flexible string-based lookup mechanism with a built-in dictionary and support for custom extensions.

## 🚀 Features

- ✅ **EN + ES** built-in translations (~100 words) — other languages fall back to EN
- 🧱 Composable messages from string keys
- 🌐 Auto-detects system/browser language
- 🧩 Custom dictionaries via `RegisterWords` — add more languages or words easily
- 🔒 Zero external dependencies
- ⚙️ Compatibility: Go + TinyGo (WASM ready)

---

## 🌍 Basic Usage

By default, the `fmt` package contains only the translation engine. To use the built-in dictionary (~100 common terms), import the sub-package:

```go
import _ "github.com/tinywasm/fmt/dictionary"
```

### Setting Language
```go
// Set global language to Spanish
code := OutLang(ES) // "ES"
code = OutLang()    // auto-detects system/browser language
```

### Translating Messages
`Translate` accepts string keys, language constants, and other types. Unknown strings are passed through as-is.

```go
// Direct string (natural lowercase keys)
msg := Translate("format", "invalid").String()
// → "format invalid" (EN)

// Force to Spanish (ES) for a single call
msg = Translate(ES, "format", "invalid").String()
// → "formato inválido"

// Composing with other types
err := Err("input", 42, "invalid")
// → "input 42 invalid"
```

---

## 🧩 Custom Words & Language Extension

`RegisterWords` lets you add new words **or extend existing ones with more languages**. Call it from `init()`.

### Add a new domain word
```go
func init() {
    fmt.RegisterWords([]fmt.DictEntry{
        {Key: "user", EN: "User", ES: "Usuario", FR: "Utilisateur"},
        {Key: "email", EN: "Email", ES: "Correo"},
    })
}
```

### Extend built-in words with more languages
The built-in dictionary only ships EN and ES. To add FR, DE, ZH, etc., register the same keys again — they will merge/override:

```go
import _ "github.com/tinywasm/fmt/dictionary" // load EN+ES

func init() {
    // Add French and German to built-in words
    fmt.RegisterWords([]fmt.DictEntry{
        {Key: "empty",   FR: "Vide",   DE: "Leer"},
        {Key: "invalid", FR: "Invalide", DE: "Ungültig"},
        {Key: "format",  FR: "Format",  DE: "Format"},
        // ... add as many as needed
    })
}
```

> **Note:** If a language field is empty, it falls back to EN automatically.


## ⚡ Performance & Memory

`Translate` and `Err` return a pooled `*Conv` object.

- **Automatic Release**: Calling `.String()` or `.Apply()` returns the object to the pool.
- **Manual Release**: If using `.Bytes()`, call `.PutConv()` manually.

```go
// ✅ Recommended usage
msg := Translate("format").String()

// ⚠️ Manual release
c := Translate("format")
b := c.Bytes()
c.PutConv()
```

### Zero-Allocation Optimization
Passing string literals is efficient (1 alloc/op for boxing in interface). For hot paths, you can register words and use them directly.

---

## ✅ Validation Example

```go
func validate(input string) error {
    if input == "" {
        return Err("string", "empty")
    }
    if _, err := Convert(input).Int(); err != nil {
        return Err("invalid", "number")
    }
    return nil
}
```

---

## 🔍 Language Constants
The following constants are supported: `EN`, `ES`, `ZH`, `HI`, `AR`, `PT`, `FR`, `DE`, `RU`.
