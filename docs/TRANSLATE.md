# 🌍 `fmt` Translation Engine (LLM Context)

Dependency-free translation engine for composable i18n messages (EN/ES built-in, Wasm-ready).

## Setup & Rules
1. **Import Dictionary**: `import _ "github.com/tinywasm/fmt/dictionary"` loads ~100 common EN/ES terms.
2. **Language**: `fmt.OutLang(fmt.ES)` sets global lang. `fmt.OutLang()` auto-detects system language.
3. **Word Order**: ALWAYS use **Noun + Adjective** (e.g., `"format", "invalid"` -> ES: *"Formato Inválido"*, EN: *"Format Invalid"*).

## API Usage
`Translate()` & `Err()` accept strings (case-insensitive keys), language constants, and other types. Unknown strings pass through.

```go
// 1. Normal usage (global lang)
msg := fmt.Translate("format", "invalid").String() 

// 2. Force specific language
msg = fmt.Translate(fmt.ES, "format", "invalid").String()

// 3. Error generation with mixed types 
err := fmt.Err("input", 42, "invalid") // -> "Input 42 Invalid"
```

## Memory Management (CRITICAL)
`Translate()` and `Err()` return a pooled `*fmt.Conv` object.
- **Auto-Release**: `.String()` or `.Apply()` auto-returns it to the pool.
- **Manual Release**: If using `.Bytes()`, you MUST call `.PutConv()` to prevent leaks.
```go
c := fmt.Translate("format")
b := c.Bytes() 
c.PutConv() // MUST release manually
```

## Custom Dictionaries
Add or extend words dynamically via `init()`. `EN` is the lookup key.
**Crucial Behavior**: If the `EN` key already exists in the dictionary (e.g., from the built-in list), `RegisterWords` will **merge (fuse)** the new translations with the existing ones. It does not overwrite or duplicate existing languages. This allows you to easily inject missing translations into the built-in dictionary without recreating it.
Supported: `EN, ES, ZH, HI, AR, PT, FR, DE, RU`.

```go
func init() {
    fmt.RegisterWords([]fmt.DictEntry{
        {EN: "User", ES: "Usuario", FR: "Utilisateur"}, // Adds a completely new word
        {EN: "Empty", FR: "Vide"},                      // Merges new FR translation into the existing "Empty" base word
    })
}
```
