package errors

import (
	"errors"
	"github.com/go-stdlib/go-errors/internal"
	"io"
)

// Assertions
var (
	_ Error   = (*internal.Defined)(nil)
	_ Grouper = (*internal.Group)(nil)
)

// Const Aliases
const (
	// DefaultNamespace is the default namespace for errors defined in this package.
	DefaultNamespace = internal.DefaultNamespace
	// FlagUnknown is set to represent unknown/unregistered errors.
	FlagUnknown = internal.FlagUnknown
	// FlagRetryable is set to represent errors that can be retried.
	FlagRetryable = internal.FlagRetryable
	// FlagTimeout is set to represent errors indicating a timeout occurred.
	FlagTimeout = internal.FlagTimeout
)

// Func Aliases
var (
	// As is alias for `errors.As`.
	As = errors.As
	// Is is alias for `errors.Is`.
	Is = errors.Is
	// Unwrap is alias for `errors.Unwrap`.
	Unwrap = errors.Unwrap
)

// Type Aliases
type (
	Defined = internal.Defined
	Error   = internal.Error
	Extras  = internal.Extras
	Flags   = internal.Flags
	Group   = internal.Group
	Grouper = internal.Grouper
)

// Var Aliases
var (

	// ErrUndefined indicates the wrapped error is not well-known and not previously
	// defined. This commonly indicates it's coming from an external system/library.
	ErrUndefined = internal.ErrUndefined
)

// Defer captures errors from calls inside a `defer` using
// a pointer to an error which may or may not already be allocated.
func Defer(err *error, errs ...error) {
	if err == nil {
		*err = Defined{}
	}
	*err = Join(*err, errs...).ErrorOrNil()
}

// DeferCloser captures errors from an `io.Closer` inside a `defer` using
// a pointer to an error which may or may not already be allocated.
func DeferCloser(err *error, closer io.Closer) {
	Defer(err, closer.Close())
}

// DeferFn captures errors from a `Close` function inside a `defer` using
// a pointer to an error which may or may not already be allocated.
func DeferFn(err *error, fn func() error) {
	Defer(err, fn())
}

// Join one or more errors into a group.
//
// If err is not already a Grouper, then it will be turned into
// one. If any of the errs are Grouper, they will be flattened
// one level into err.
// Any nil errors within errs will be ignored. If err is nil, a new
// Grouper will be returned containing the given errs.
func Join(err error, errs ...error) Grouper {
	var g *internal.Group

	switch {
	case errors.As(err, &g):
		g.Append(errs...)
		return g
	default:
		g = internal.NewGroup(err)
		g.Append(errs...)
		return g
	}
}
