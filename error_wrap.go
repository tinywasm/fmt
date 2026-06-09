package fmt

// wrappedErr is a long-lived error that joins a message with an
// Unwrap() chain compatible with errors.Is / errors.As (Go 1.20+).
// It is a separate type from *Conv — it does not use the pool.
type wrappedErr struct {
	msg  string  // pre-built message (no extra allocation)
	errs []error // cause + sentinel(s)
}

func (e *wrappedErr) Error() string   { return e.msg }
func (e *wrappedErr) Unwrap() []error { return e.errs }

// ErrType creates an error that:
//   - shows: cause.Error() + ": " + sentinel.Error()
//   - allows: errors.Is(result, sentinel) == true
//   - allows: errors.Is(result, cause)    == true  (if cause supports it)
//
// The sentinel acts as the type identity of the error — analogous to the
// category, class or "type" to which the resulting error belongs.
//
// If cause is nil, it returns sentinel directly (no wrapping).
//
// Example:
//
//	return fmt.ErrType(dbErr, ErrSyncFailed)
func ErrType(cause, sentinel error) error {
	if cause == nil {
		return sentinel
	}
	if sentinel == nil {
		return cause
	}
	return &wrappedErr{
		msg:  cause.Error() + ": " + sentinel.Error(),
		errs: []error{cause, sentinel},
	}
}
