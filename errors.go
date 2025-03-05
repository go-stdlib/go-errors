package errors

import (
	"errors"
	"github.com/go-stdlib/go-errors/internal"
)

// Assertions
var (
	_ Error   = (*internal.Canonical)(nil)
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

// Type Aliases
type (
	Canonical = internal.Canonical
	Error     = internal.Error
	Extras    = internal.Extras
	Flags     = internal.Flags
	Group     = internal.Group
	Grouper   = internal.Grouper
)

// Var Aliases
var (
	// ErrUnknown indicates the wrapped error is not well-known and not previously
	// defined. This commonly indicates it's coming from an external system/library.
	ErrUnknown = internal.ErrUnknown
)

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
