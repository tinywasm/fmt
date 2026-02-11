//go:build wasm

package fmt

func (c *Conv) toInt64Reflect(_ any) (int64, bool) {
	return 0, false
}
