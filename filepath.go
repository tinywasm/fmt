package fmt

// PathJoin joins path elements using the appropriate separator.
// Accepts variadic string arguments and returns a Conv instance for method chaining.
// Detects Windows paths (backslash) or Unix paths (forward slash).
// Empty elements are ignored.
//
// Usage patterns:
//   - PathJoin("a", "b", "c").String()           // -> "a/b/c"
//   - PathJoin("a", "B", "c").ToLower().String() // -> "a/b/c"
//
// Examples:
//
//	PathJoin("a", "b", "c").String()           // -> "a/b/c"
//	PathJoin("/root", "sub", "file").String()   // -> "/root/sub/file"
//	PathJoin(`C:\dir`, "file").String()        // -> "C:\dir\file"
//	PathJoin("a", "", "b").String()            // -> "a/b"
//	PathJoin("a", "B", "c").ToLower().String() // -> "a/b/c"
func PathJoin(elem ...string) *Conv {
	c := GetConv()

	if len(elem) == 0 {
		return c
	}

	sep := "/"
	// detect separator from first element with a separator
	for _, e := range elem {
		if Index(e, "\\") != -1 {
			sep = "\\"
			break
		}
	}

	for i, e := range elem {
		if e == "" {
			continue
		}

		curr := c.GetString(BuffOut)

		// trim leading separators only if not the first element
		if i > 0 && len(curr) > 0 {
			for len(e) > 0 && (e[0] == '/' || e[0] == '\\') {
				e = e[1:]
			}
		}

		// add separator if needed
		if len(curr) > 0 && !HasSuffix(curr, sep) && e != "" {
			c.WrString(BuffOut, sep)
		}
		c.WrString(BuffOut, e)
	}

	return c
}

// pathClean normalizes a path by detecting the separator and handling special cases.
// Returns the cleaned path and the detected separator.
// This is a helper function used by PathBase and PathExt to avoid code duplication.
func pathClean(path string) (string, byte) {
	if path == "" {
		return ".", '/'
	}

	// prefer backslash if present
	sep := byte('/')
	if Index(path, "\\") != -1 {
		sep = '\\'
	}

	// windows drive root like "C:\" or with only extra separators -> return "\\"
	if sep == '\\' && len(path) >= 2 && path[1] == ':' {
		onlySep := true
		for i := 2; i < len(path); i++ {
			if path[i] != '\\' && path[i] != '/' {
				onlySep = false
				break
			}
		}
		if onlySep {
			return "\\", sep
		}
	}

	// trim trailing separators
	for len(path) > 1 && path[len(path)-1] == sep {
		path = path[:len(path)-1]
	}

	// if path reduced to a single root separator, return it
	if len(path) == 1 && (path[0] == '/' || path[0] == '\\') {
		return path, sep
	}

	return path, sep
}

// extractBase returns the base filename from a cleaned path if prefix is empty.
// If prefix is set, it attempts to return the relative path from that prefix.
func extractBase(cleaned string, sep byte, prefix string) string {
	// If prefix is set, try to strip it
	if prefix != "" {
		if HasPrefix(cleaned, prefix) {
			rel := cleaned[len(prefix):]
			// check if it's a full component match
			isRoot := len(prefix) == 1 && (prefix[0] == '/' || prefix[0] == '\\')
			if isRoot || len(rel) == 0 || rel[0] == '/' || rel[0] == '\\' {
				if len(rel) > 0 && (rel[0] == '/' || rel[0] == '\\') {
					return rel[1:]
				}
				return rel
			}
		}
		return cleaned
	}

	// Default behavior: extract filename after last separator
	// Special cases
	if cleaned == "." || cleaned == "\\" || cleaned == "/" {
		return ""
	}

	// search from end for last separator
	for i := len(cleaned) - 1; i >= 0; i-- {
		if cleaned[i] == sep {
			return cleaned[i+1:]
		}
	}
	// no separator found - whole cleaned path is the base
	return cleaned
}

// PathBase returns the last element of path, similar to
// filepath.Base from the Go standard library. It treats
// trailing slashes specially ("/a/b/" -> "b") and preserves
// a single root slash ("/" -> "/"). An empty path returns ".".
//
// The implementation uses tinystring helpers (HasSuffix and Index)
// to avoid importing the standard library and keep the function
// minimal and TinyGo-friendly.
//
// Examples:
//
//		PathBase("/a/b/c.txt") // -> "c.txt"
//		PathBase("folder/file.txt")   // -> "file.txt"
//		PathBase("")           // -> "."
//	 PathBase("c:\file program\app.exe") // -> "app.exe"
//
// PathBase writes the last element of the path into the Conv output buffer.
// Use it as: Convert(path).PathBase().String() and it behaves similarly to
// filepath.Base. Examples:
//
// Convert("/a/b/c.txt").PathBase().String() // -> "c.txt"
// Convert("folder/file.txt").PathBase().String()   // -> "file.txt"
// Convert("").PathBase().String()           // -> "."
// Convert(`c:\file program\app.exe`).PathBase().String() // -> "app.exe"
func (c *Conv) PathBase() *Conv {
	// read source path from buffer
	src := c.GetString(BuffOut)

	cleaned, sep := pathClean(src)

	// clear output buffer - PathBase will write the resulting base
	c.ResetBuffer(BuffOut)

	base := extractBase(cleaned, sep, "")
	if base == "" {
		// Special case: write the cleaned value (., /, or \)
		c.WrString(BuffOut, cleaned)
	} else {
		c.WrString(BuffOut, base)
	}

	return c
}

// PathExt extracts the file extension from a path and writes it to the Conv buffer.
// Returns the Conv instance for method chaining.
// An empty extension returns an empty string.
//
// Examples:
//
//	Convert("/a/b/c.txt").PathExt().String() // -> ".txt"
//	Convert("file.tar.gz").PathExt().String() // -> ".gz"
//	Convert("noext").PathExt().String()       // -> ""
//	Convert("C:\\dir\\app.EXE").PathExt().ToLower().String() // -> ".exe"
func (c *Conv) PathExt() *Conv {
	// Read current path from output buffer
	src := c.GetString(BuffOut)

	cleaned, sep := pathClean(src)

	// clear output buffer - PathExt returns only the extension
	c.ResetBuffer(BuffOut)

	// get the base filename using helper
	base := extractBase(cleaned, sep, "")
	if base == "" {
		// Special cases like ".", "/", "\\" have no extension
		return c
	}

	// special cases: "." and ".." have no extension
	if base == "." || base == ".." {
		return c
	}

	// search for last dot in base filename
	for i := len(base) - 1; i >= 0; i-- {
		if base[i] == '.' {
			// don't count leading dot (hidden files like .bashrc)
			if i == 0 {
				return c
			}
			c.WrString(BuffOut, base[i:])
			return c
		}
	}

	return c
}

// pathBase stores the base path for shortening operations.
var pathBase string

// SetPathBase sets the base path for PathShort operations.
// Optional: if not called, PathShort auto-detects using GetPathBase (os.Getwd or syscall/js).
func SetPathBase(base string) {
	pathBase, _ = pathClean(base)
}

// PathShort shortens absolute paths relative to base path.
// It can handle paths embedded in larger strings (e.g. log messages).
// Auto-detects base path via GetPathBase() if SetPathBase was not called.
// Returns relative path with "./" prefix for minimal output.
// Example: "Compiling /home/user/project/src/file.go ..." -> "Compiling ./src/file.go ..."
func (c *Conv) PathShort() *Conv {
	if pathBase == "" {
		pathBase = GetPathBase()
	}

	if pathBase == "" {
		return c
	}

	src := c.GetStringZeroCopy(BuffOut)
	if src == "" {
		return c
	}

	// We'll build the result in the work buffer to avoid multiple allocations
	c.ResetBuffer(BuffWork)

	start := 0
	for {
		idx := Index(src[start:], pathBase)
		if idx == -1 {
			c.WrString(BuffWork, src[start:])
			break
		}

		matchIdx := start + idx
		c.WrString(BuffWork, src[start:matchIdx])

		// Validate match boundary
		endIdx := matchIdx + len(pathBase)
		isRoot := len(pathBase) == 1 && (pathBase[0] == '/' || pathBase[0] == '\\')

		valid := false
		if isRoot {
			// Root is valid if it's the start of a component
			if matchIdx == 0 {
				valid = true
			} else {
				prevChar := src[matchIdx-1]
				if prevChar == ' ' || prevChar == '\t' || prevChar == '\n' || prevChar == '\r' || prevChar == '"' || prevChar == '\'' || prevChar == '(' {
					valid = true
				}
			}
			// Root followed by another separator is not a valid single root match (e.g. //)
			if valid && endIdx < len(src) && (src[endIdx] == '/' || src[endIdx] == '\\') {
				valid = false
			}
		} else {
			if endIdx == len(src) {
				valid = true
			} else {
				nextChar := src[endIdx]
				if nextChar == '/' || nextChar == '\\' {
					valid = true
				}
			}
		}

		if valid {
			if isRoot {
				if endIdx == len(src) {
					c.WrString(BuffWork, ".")
				} else {
					c.WrString(BuffWork, "./")
				}
				start = endIdx
			} else {
				c.WrString(BuffWork, ".")

				// If followed by a separator, consume it and write "/" to normalize
				if endIdx < len(src) && (src[endIdx] == '/' || src[endIdx] == '\\') {
					c.WrString(BuffWork, "/")
					start = endIdx + 1
				} else {
					start = endIdx
				}
			}
		} else {
			// Not a valid path boundary, just copy the match and continue
			c.WrString(BuffWork, pathBase)
			start = endIdx
		}
	}

	// Swap BuffWork to BuffOut
	c.swapBuff(BuffWork, BuffOut)

	return c
}
