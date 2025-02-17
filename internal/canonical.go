package internal

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
)

const (
	// DefaultNamespace is the default namespace for errors defined in this package.
	DefaultNamespace = "go-stdlib/go-errors"
)

// ErrorKey returns a slug that should be unique for each error (namespace + code).
func ErrorKey(namespace Namespace, code Code) string {
	return fmt.Sprintf("%s/%s", namespace, code)
}

// Code represents a human-readable code that identifies a canonical error.
// A code value should be unique within a namespace.
type Code string

// Namespace represents a human-readable namespace that identifies a logical
// grouping of errors.
type Namespace string

// ErrUnknown indicates the wrapped error is not well-known and not previously
// defined. This commonly indicates it's coming from an external system/library.
var ErrUnknown = Canonical{
	Code:      "unknown",
	Flags:     FlagUnknown,
	Message:   "wrapped error is unknown",
	Namespace: DefaultNamespace,
}

// Canonical represents a known/defined application error.
type Canonical struct {
	// Code is a machine-readable representation of the error.
	Code Code `json:"code"`
	// Extras is an optional struct to store execution context
	// that is helpful for understanding the error.
	Extras Extras `json:"extras,omitempty"`
	// Flags is a bitmask that contains additional classification/context
	// for the error, e.g. indicating if the error can be retried.
	Flags Flags `json:"flags,omitempty"`
	// Message is a human-readable representation for the error.
	Message string `json:"message"`
	// Namespace is a machine-readable representation for a bucketing/grouping
	// concept of errors. This is commonly used to indicate the package/repository/service
	// an error originated from.
	Namespace Namespace `json:"namespace"`
	// Wrapped is a wrapped error if this was created from another via `Wrap`. This
	// is hidden from human consumers and only visible to machine/operators.
	Wrapped error `json:"-"`
}

// AsGroup returns a Grouper containing this error and all the
// wrapped errors it contains.
func (c Canonical) AsGroup() Grouper {
	g := NewGroup(c)

	err := c
	for err.Wrapped != nil {
		g.Append(err.Wrapped)

		var ce Canonical
		if !errors.As(err.Wrapped, &ce) {
			break
		}
		err = ce
	}

	return g
}

// Copy returns a full copy of the error, including copies
// of all wrapped errors within.
func (c Canonical) Copy() Error {
	if c.Wrapped != nil {
		var wrapped Canonical

		if errors.As(c.Wrapped, &wrapped) {
			return Canonical{
				Code:      c.Code,
				Extras:    c.Extras,
				Flags:     c.Flags,
				Message:   c.Message,
				Namespace: c.Namespace,
				Wrapped:   wrapped.Copy(),
			}
		}
	}
	return Canonical{
		Code:      c.Code,
		Extras:    c.Extras,
		Flags:     c.Flags,
		Message:   c.Message,
		Namespace: c.Namespace,
		Wrapped:   c.Wrapped,
	}
}

// Equal returns true if the two Errors are equal.
func (c Canonical) Equal(e Error) bool {
	var ce Canonical
	if !errors.As(e, &ce) {
		return false
	}
	return c.Code == ce.Code &&
		c.Message == ce.Message &&
		c.Namespace == ce.Namespace &&
		c.Flags == ce.Flags &&
		reflect.DeepEqual(c.Extras, ce.Extras)
}

// Key returns a value that uniquely identifies the type of error.
func (c Canonical) Key() string {
	return ErrorKey(c.Namespace, c.Code)
}

// IsZero returns true if the Canonical is an empty/zero value.
func (c Canonical) IsZero() bool {
	return reflect.DeepEqual(c, new(Canonical))
}

// IsRetryable returns true if the error indicates the failed operation
// is safe to retry.
func (c Canonical) IsRetryable() bool { return c.Flags.Has(FlagRetryable) }

// IsTimeout returns true if the error indicates an operation timeout.
func (c Canonical) IsTimeout() bool { return c.Flags.Has(FlagTimeout) }

// IsTransient returns true if the error indicates the operation failure
// is transient and a result might be different if tried at another time.
func (c Canonical) IsTransient() bool { return c.Flags.Has(FlagUnknown) }

// WithExtras returns a new copy of the error with the extras added.
func (c Canonical) WithExtras(extras Extras) Error {
	return Canonical{
		Code:      c.Code,
		Extras:    extras,
		Flags:     c.Flags,
		Message:   c.Message,
		Namespace: c.Namespace,
		Wrapped:   c.Wrapped,
	}
}

// WithFlags returns a new copy of the error with the given attributes applied.
func (c Canonical) WithFlags(flags Flags) Error {
	return Canonical{
		Code:      c.Code,
		Extras:    c.Extras,
		Flags:     c.Flags.Set(flags),
		Message:   c.Message,
		Namespace: c.Namespace,
		Wrapped:   c.Wrapped,
	}
}

// WithTags returns a new copy of the error with the additional tags added.
func (c Canonical) WithTags(tags ...string) Error {
	return Canonical{
		Code:      c.Code,
		Extras:    c.Extras.WithTags(tags...),
		Flags:     c.Flags,
		Message:   c.Message,
		Namespace: c.Namespace,
		Wrapped:   c.Wrapped,
	}
}

// String returns the Canonical string representation.
//
// Interface: fmt.Stringer.
func (c Canonical) String() string {
	return c.Error()
}

// Format returns a complex string representation of the Canonical
// for the given verbs.
//
// Interface: fmt.Formatter.
func (c Canonical) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if _, err := io.WriteString(s, c.AsGroup().Error()); err != nil {
				panic(err)
			}
			return
		}
		fallthrough
	case 's':
		if _, err := io.WriteString(s, c.Error()); err != nil {
			panic(err)
		}
	case 'q':
		if _, err := io.WriteString(s, c.Error()); err != nil {
			panic(err)
		}
	}
}

// Canonical returns the string representation of the Canonical.
//
// Interface: error.
func (c Canonical) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[%s:%s] %s", c.Namespace, c.Code, c.Message))
	if c.Wrapped != nil {
		sb.WriteString(fmt.Sprintf("\n-> %s", c.Wrapped.Error()))
	}
	return sb.String()
}

// Is implements error equality checking.
//
// Interface: HasIs.
func (c Canonical) Is(target error) bool {
	var err Canonical
	if !errors.As(target, &err) {
		return false
	}
	return c.Equal(err)
}

// Unwrap implements error unwrapping for nested errors.
//
// Interface: Unwrap.
func (c Canonical) Unwrap() error {
	return c.Wrapped
}

// Wrap returns a new Error with the given err wrapped.
//
// If the given err is also an Canonical and the current instance
// is a zero value, just return a copy of the given Canonical. This
// allows us to avoid checking this case at every call-site; we
// can just Wrap the error and handle it.
func (c Canonical) Wrap(err error) Error {
	if err == nil {
		return c
	}
	if c.IsZero() {
		var ce Canonical
		if errors.As(err, &ce) {
			return ce.Copy()
		}
	}
	return Canonical{
		Code:      c.Code,
		Extras:    c.Extras,
		Flags:     c.Flags,
		Message:   c.Message,
		Namespace: c.Namespace,
		Wrapped:   err,
	}
}

// Wrapf returns a new Canonical with an error created by the given format + args.
func (c Canonical) Wrapf(format string, a ...any) Error {
	return c.Wrap(fmt.Errorf(format, a...))
}
