//go:build wasm

package fmt

// WASM is single-threaded: use simple allocation instead of sync.Pool
func GetConv() *Conv {
	return &Conv{
		out:  make([]byte, 0, 64),
		work: make([]byte, 0, 64),
		err:  make([]byte, 0, 64),
	}
}

func (c *Conv) PutConv() {
	// No-op in WASM — GC handles cleanup
}

func (c *Conv) putConv() {
	// No-op in WASM
}
