//go:build !wasm

package fmt

import "reflect"

func (c *Conv) toFloat64Reflect(arg any) (float64, bool) {
	rv := reflect.ValueOf(arg)
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return rv.Float(), true
	default:
		return 0, false
	}
}
