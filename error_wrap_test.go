package fmt

import (
	"errors"
	"testing"
)

func TestErrType(t *testing.T) {
	sentinel := Err("sentinel error")
	cause := Err("original cause")
	other := Err("other error")

	t.Run("Wrapping", func(t *testing.T) {
		err := ErrType(cause, sentinel)

		// Check message
		expectedMsg := "original cause: sentinel error"
		if err.Error() != expectedMsg {
			t.Errorf("expected %q, got %q", expectedMsg, err.Error())
		}

		// Check errors.Is for sentinel
		if !errors.Is(err, sentinel) {
			t.Error("errors.Is(err, sentinel) should be true")
		}

		// Check errors.Is for cause
		if !errors.Is(err, cause) {
			t.Error("errors.Is(err, cause) should be true")
		}

		// Check errors.Is for other
		if errors.Is(err, other) {
			t.Error("errors.Is(err, other) should be false")
		}
	})

	t.Run("NilCause", func(t *testing.T) {
		err := ErrType(nil, sentinel)
		if err != sentinel {
			t.Error("ErrType(nil, sentinel) should return sentinel directly")
		}
	})

	t.Run("NilSentinel", func(t *testing.T) {
		err := ErrType(cause, nil)
		if err != cause {
			t.Error("ErrType(cause, nil) should return cause directly")
		}
	})

	t.Run("NestedWrapping", func(t *testing.T) {
		innerSentinel := Err("inner sentinel")
		outerSentinel := Err("outer sentinel")
		baseCause := Err("base cause")

		innerWrap := ErrType(baseCause, innerSentinel)
		outerWrap := ErrType(innerWrap, outerSentinel)

		if !errors.Is(outerWrap, outerSentinel) {
			t.Error("should detect outer sentinel")
		}
		if !errors.Is(outerWrap, innerSentinel) {
			t.Error("should detect inner sentinel")
		}
		if !errors.Is(outerWrap, baseCause) {
			t.Error("should detect base cause")
		}

		expectedMsg := "base cause: inner sentinel: outer sentinel"
		if outerWrap.Error() != expectedMsg {
			t.Errorf("expected %q, got %q", expectedMsg, outerWrap.Error())
		}
	})
}
