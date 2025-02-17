package internal

import (
	"fmt"
	"sort"
)

// Error represents an interface to a known/defined application error.
type Error interface {
	error
	fmt.Stringer
	fmt.Formatter
	HasIs
	HasUnwrap

	// AsGroup returns a Grouper containing this error and all the
	// wrapped errors it contains.
	AsGroup() Grouper
	// Copy returns a full copy of the error, including copies
	// of all wrapped errors within.
	Copy() Error
	// Equal returns true if the two Errors are equal.
	Equal(Error) bool
	// Key returns a value that uniquely identifies the error.
	Key() string
	// WithExtras returns a new copy of the error with the extras added.
	WithExtras(Extras) Error
	// WithFlags returns a new copy of the error with the given attributes applied.
	WithFlags(Flags) Error
	// WithTags returns a new copy of the error with the given tags applied.
	WithTags(...string) Error
}

// Causer defines types that return the underlying cause of an error.
type Causer interface {
	Cause() error
}

// HasAs defines types necessary for stdlib `errors.As` support.
type HasAs interface {
	As(target any) bool
}

// HasIs defines types necessary for stdlib `errors.Is` support.
type HasIs interface {
	Is(target error) bool
}

// HasUnwrap defines types necessary for stdlib `errors.Unwrap` support.
type HasUnwrap interface {
	Unwrap() error
}

// Grouper represents an interface for grouping errors together.
type Grouper interface {
	error
	fmt.Stringer
	sort.Interface

	// Append one or more errors to the group.
	Append(...error)
	// Empty returns true if the group contains no errors.
	Empty() bool
	// ErrorOrNil returns an error if the group contains one or more or nil when empty.
	ErrorOrNil() error
}
