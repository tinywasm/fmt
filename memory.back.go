//go:build !wasm

package fmt

import "sync"

// Reuse Conv objects to eliminate the 53.67% allocation hotspot from newConv()
var convPool = sync.Pool{
	New: func() any {
		return &Conv{
			out:  make([]byte, 0, 64),
			work: make([]byte, 0, 64),
			err:  make([]byte, 0, 64),
			// TODO: Add bufFmt when struct is updated
		}
	},
}

// GetConv gets a reusable Conv from the pool
// FIXED: Ensures object is completely clean to prevent race conditions
func GetConv() *Conv {
	c := convPool.Get().(*Conv)
	// Defensive cleanup: ensure object is completely clean
	c.resetAllBuffers()
	c.out = c.out[:0]
	c.work = c.work[:0]
	c.err = c.err[:0]
	c.dataPtr = nil
	c.kind = K.String
	return c
}

// PutConv returns a Conv to the pool after resetting it
func (c *Conv) PutConv() {
	// Reset all buffer positions using centralized method
	c.resetAllBuffers()
	// Clear buffer contents (keep capacity for reuse)
	c.out = c.out[:0]
	c.work = c.work[:0]
	c.err = c.err[:0]

	// Reset other fields to default state - only keep dataPtr and Kind
	c.dataPtr = nil
	c.kind = K.String

	convPool.Put(c)
}

// putConv returns a Conv to the pool after resetting it (internal)
func (c *Conv) putConv() {
	c.PutConv()
}
