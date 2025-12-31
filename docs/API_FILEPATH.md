# Filepath Package Equivalents

Replace common `filepath` package functions with fmt equivalents:

| Go Standard | fmt Equivalent |
|-------------|----------------------|
| `filepath.Base()` | `Convert(path).PathBase().String()` |
| `filepath.Join()` | `PathJoin("a", "b", "c").String()` — variadic function, zero heap allocation for <=8 elements |
| Shorten path  | `Convert(log).PathShort().String()` — Shorten absolute paths relative to WD or custom base |

## PathBase (fluent API)

Use `Convert(path).PathBase().String()` to get the last element of a path.

Examples:

```go
Convert("/a/b/c.txt").PathBase().String() // -> "c.txt"
Convert("folder/file.txt").PathBase().String()   // -> "file.txt"
Convert("").PathBase().String()           // -> "."
Convert(`c:\\file program\\app.exe`).PathBase().String() // -> "app.exe"
```

## PathJoin (cross-platform path joining)

Standalone function with variadic string arguments.
Returns *Conv for method chaining with transformations like ToLower().
Uses fixed array for zero heap allocation (≤8 elements).
Detects separator ("/" or "\\") automatically and avoids duplicates.

Examples:

```go
PathJoin("a", "b", "c").String()            // -> "a/b/c"
PathJoin("/root", "sub", "file").String()   // -> "/root/sub/file"
PathJoin(`C:\dir`, "file").String()         // -> `C:\dir\file`
PathJoin(`\\server`, "share", "file").String() // -> `\\server\share\file`

// Typical use: normalize path case with ToLower() in the same chain
PathJoin("A", "B", "C").ToLower().String() // -> "a/b/c"
```

## Path Extension

Get the file extension (including the leading dot) from a path. Use the
fluent API form `Convert(path).PathExt().String()` which reads the path
from the Conv buffer and returns only the extension (or empty string).

Examples:

```go
Convert("file.txt").PathExt().String()          // -> ".txt"
Convert("/path/to/archive.tar.gz").PathExt().String() // -> ".gz"
Convert(".bashrc").PathExt().String()           // -> ""  (hidden file, no ext)
Convert("noext").PathExt().String()             // -> ""
Convert(`C:\\dir\\app.exe`).PathExt().String() // -> ".exe"

// Typical use: normalize extension case in the same chain. For example,
// when the extension is uppercase you can lower-case it immediately:
Convert("file.TXT").PathExt().ToLower().String() // -> ".txt"
```

## PathShort (smart path shortening)

Use `PathShort()` to convert absolute paths into relative paths starting with `./`. It is particularly useful for log messages where you want to hide long system paths.

> [!NOTE] 
> This feature uses platform-specific logic to determine the base path: `os.Getwd()` in standard environments and `window.location.origin` in WebAssembly.

### Features
- **Auto-detection**: If no base path is set, it uses `os.Getwd()` automatically.
- **SetPathBase**: Configure a custom project root globally with `SetPathBase(path)`.
- **Smart Replacement**: Works with paths embedded within longer strings (like log messages).
- **Consistently relative**: Prepends `./` to any shortened path.

### Examples

```go
// Optional: set custom base path (if not set, uses current working directory)
SetPathBase("/home/user/project")

// 1. Direct path shortening
Convert("/home/user/project/web/public").PathShort().String() 
// -> "./web/public"

// 2. Embedded paths (Smart detection in logs)
log := "Compiling WASM due to /home/user/project/web/client.go change..."
Convert(log).PathShort().String()
// -> "Compiling WASM due to ./web/client.go change..."

// 3. Multiple paths in same string
msg := "Moving /home/user/project/a to /home/user/project/b"
Convert(msg).PathShort().String()
// -> "Moving ./a to ./b"

// 4. Custom root as "/"
SetPathBase("/")
Convert("/etc/passwd").PathShort().String()
// -> "./etc/passwd"
```

## GetPathBase

Standalone function that returns the default base path automatically used by `PathShort` when no custom base is set via `SetPathBase`.

| Environment | Behavior |
|-------------|----------|
| **Standard** | Returns current working directory via `os.Getwd()`. |
| **WASM**     | Returns the domain root via `window.location.origin`. |

Example:
```go
base := GetPathBase()
println("Base path is:", base)
```