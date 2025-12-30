//go:build !wasm

package fmt

import (
	"os"
)

var pathBase string

// SetPathBase sets the base path for PathShort operations.
// Optional: if not called, PathShort auto-detects using os.Getwd().
func SetPathBase(base string) {
	pathBase, _ = pathClean(base)
}

// PathShort shortens absolute paths relative to base path.
// It can handle paths embedded in larger strings (e.g. log messages).
// Auto-detects base path via os.Getwd() if SetPathBase was not called.
// Returns relative path with "./" prefix for minimal output.
// Example: "Compiling /home/user/project/src/file.go ..." -> "Compiling ./src/file.go ..."
func (c *Conv) PathShort() *Conv {
	if pathBase == "" {
		if wd, err := os.Getwd(); err == nil {
			pathBase, _ = pathClean(wd)
		}
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
