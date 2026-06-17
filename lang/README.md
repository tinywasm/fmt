# fmt/lang

The `lang` package provides multilingual support and translation services for the `tinywasm/fmt` library.

## Features

- **9 Supported Languages**: English (EN), Spanish (ES), Chinese (ZH), Hindi (HI), Arabic (AR), Portuguese (PT), French (FR), German (DE), Russian (RU).
- **Opt-in Design**: Importing this package automatically enables translation support in the root `fmt` package via a global hook.
- **Noun + Adjective Order**: Optimized for Spanish and English natural phrasing.
- **Manual Language Detection**: Detects language from environment variables (backend) or browser settings (Wasm).
- **Zero-Allocation**: Leverages the root `fmt` package's buffer pooling for efficient translation.

## Installation

```bash
import "github.com/tinywasm/fmt/lang"
```

Simply importing the package activates translation for `fmt.Err()`, `fmt.Html()`, and `%L` formatting in `fmt.Sprintf()`.

## API

### Global Language Configuration

- `OutLang(l ...any) string`: Sets or gets the current global language. Accepts `lang` constants (e.g., `lang.ES`) or string codes (e.g., `"fr"`). Returns the current language code.
- `OutLang()`: Auto-detects system/browser language and sets it as the global default.

### Translation

- `Translate(values ...any) *fmt.Conv`: Translates multiple terms using the dictionary and joins them with spaces.
- `RegisterWords(entries []DictEntry)`: Adds or merges new words into the global dictionary.

### Constants

- `EN, ES, ZH, HI, AR, PT, FR, DE, RU`: Language identifiers.

## Documentation

For a detailed guide on how to write translatable messages, see the [Translation Guide](../docs/TRANSLATE.md).
