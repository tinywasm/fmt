//go:build !wasm

package fmt

import "reflect"

func (c *Conv) toInt64Reflect(arg any) (int64, bool) {
	rv := reflect.ValueOf(arg)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(rv.Uint()), true
	default:
		return 0, false
	}
}
