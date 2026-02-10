//go:build wasm

package fmt

func (c *Conv) toFloat64Reflect(_ any) (float64, bool) {
	return 0, false
}
