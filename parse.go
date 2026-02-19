package fmt

// KeyValue represents a key-value pair extracted from a string.
type KeyValue struct {
	Key   string // Key
	Value string // Value
}

// TagPairs searches for a key in a Go struct tag and parses its value as multiple key-value pairs.
// Example: Convert(`options:"key1:text1,key2:text2"`).TagPairs("options") => []KeyValue{{Key: "key1", Value: "text1"}, {Key: "key2", Value: "text2"}}
func (c *Conv) TagPairs(key string) []KeyValue {
	val, ok := c.TagValue(key)
	if !ok || val == "" {
		return nil
	}

	// Split by comma to get pairs
	pairs := c.splitStr(val, ",")
	res := make([]KeyValue, 0, len(pairs))

	for _, pair := range pairs {
		k, v, found := c.splitByDelimiterWithBuffer(pair, ":")
		if found {
			res = append(res, KeyValue{Key: k, Value: v})
		}
	}
	return res
}

// TagValue searches for the value of a key in a Go struct tag-like string.
// Example: Convert(`json:"name" Label:"Nombre"`).TagValue("Label") => "Nombre", true
func (c *Conv) TagValue(key string) (string, bool) {
	src := c.GetString(BuffOut)

	// Reutilizar splitStr para dividir por espacios
	parts := c.splitStr(src)

	for _, part := range parts {
		// Split by ':' using existing function
		k, v, found := c.splitByDelimiterWithBuffer(part, ":")
		if !found {
			continue
		}

		if k == key {
			// Remove quotes if present
			if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
				v = v[1 : len(v)-1]
			}
			return v, true
		}
	}
	return "", false
}

// ExtractValue extracts the value after the first delimiter. If not found, returns an error.
// Usage: Convert("key:value").ExtractValue(":") => "value", nil
// If no delimiter is provided, uses ":" by default.
func (c *Conv) ExtractValue(delimiters ...string) (string, error) {
	src := c.String()
	d := ":"
	if len(delimiters) > 0 && delimiters[0] != "" {
		d = delimiters[0]
	}
	if src == d {
		return "", nil
	}
	_, after, found := c.splitByDelimiterWithBuffer(src, d)
	if !found {
		return "", c.wrErr("format", "invalid", "delimiter", "not", "found")
	}
	return after, nil
}
