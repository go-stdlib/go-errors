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

// ErrUndefined indicates the wrapped error is not well-known and not previously
// defined. This commonly indicates it's coming from an external system/library.
var ErrUndefined = Defined{
	Code:      "unknown",
	Flags:     FlagUnknown,
	Message:   "wrapped error is unknown",
	Namespace: DefaultNamespace,
}

// Defined represents a known/defined application error.
type Defined struct {
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
func (s Defined) AsGroup() Grouper {
	g := NewGroup(s)

	err := s
	for err.Wrapped != nil {
		g.Append(err.Wrapped)

		var ce Defined
		if !errors.As(err.Wrapped, &ce) {
			break
		}
		err = ce
	}

	return g
}

// Copy returns a full copy of the error, including copies
// of all wrapped errors within.
func (s Defined) Copy() Error {
	if s.Wrapped != nil {
		var wrapped Defined

		if errors.As(s.Wrapped, &wrapped) {
			return Defined{
				Code:      s.Code,
				Extras:    s.Extras,
				Flags:     s.Flags,
				Message:   s.Message,
				Namespace: s.Namespace,
				Wrapped:   wrapped.Copy(),
			}
		}
	}
	return Defined{
		Code:      s.Code,
		Extras:    s.Extras,
		Flags:     s.Flags,
		Message:   s.Message,
		Namespace: s.Namespace,
		Wrapped:   s.Wrapped,
	}
}

// Equal returns true if the two Errors are equal.
func (s Defined) Equal(e Error) bool {
	var ce Defined
	if !errors.As(e, &ce) {
		return false
	}
	return s.Code == ce.Code &&
		s.Message == ce.Message &&
		s.Namespace == ce.Namespace &&
		s.Flags == ce.Flags &&
		reflect.DeepEqual(s.Extras, ce.Extras)
}

// Key returns a value that uniquely identifies the type of error.
func (s Defined) Key() string {
	return ErrorKey(s.Namespace, s.Code)
}

// IsZero returns true if the Defined is an empty/zero value.
func (s Defined) IsZero() bool {
	return reflect.DeepEqual(s, new(Defined))
}

// IsRetryable returns true if the error indicates the failed operation
// is safe to retry.
func (s Defined) IsRetryable() bool { return s.Flags.Has(FlagRetryable) }

// IsTimeout returns true if the error indicates an operation timeout.
func (s Defined) IsTimeout() bool { return s.Flags.Has(FlagTimeout) }

// IsTransient returns true if the error indicates the operation failure
// is transient and a result might be different if tried at another time.
func (s Defined) IsTransient() bool { return s.Flags.Has(FlagUnknown) }

// WithExtras returns a new copy of the error with the extras added.
func (s Defined) WithExtras(extras Extras) Error {
	return Defined{
		Code:      s.Code,
		Extras:    extras,
		Flags:     s.Flags,
		Message:   s.Message,
		Namespace: s.Namespace,
		Wrapped:   s.Wrapped,
	}
}

// WithFlags returns a new copy of the error with the given attributes applied.
func (s Defined) WithFlags(flags Flags) Error {
	return Defined{
		Code:      s.Code,
		Extras:    s.Extras,
		Flags:     s.Flags.Set(flags),
		Message:   s.Message,
		Namespace: s.Namespace,
		Wrapped:   s.Wrapped,
	}
}

// WithTags returns a new copy of the error with the additional tags added.
func (s Defined) WithTags(tags ...string) Error {
	return Defined{
		Code:      s.Code,
		Extras:    s.Extras.WithTags(tags...),
		Flags:     s.Flags,
		Message:   s.Message,
		Namespace: s.Namespace,
		Wrapped:   s.Wrapped,
	}
}

// String returns the Defined string representation.
//
// Interface: fmt.Stringer.
func (s Defined) String() string {
	return s.Error()
}

// Format returns a complex string representation of the Defined
// for the given verbs.
//
// Interface: fmt.Formatter.
func (s Defined) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		if st.Flag('+') {
			if _, err := io.WriteString(st, s.AsGroup().Error()); err != nil {
				panic(err)
			}
			return
		}
		fallthrough
	case 's':
		if _, err := io.WriteString(st, s.Error()); err != nil {
			panic(err)
		}
	case 'q':
		if _, err := io.WriteString(st, s.Error()); err != nil {
			panic(err)
		}
	}
}

// Defined returns the string representation of the Defined.
//
// Interface: error.
func (s Defined) Error() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("[%s:%s] %s", s.Namespace, s.Code, s.Message))
	if s.Wrapped != nil {
		sb.WriteString(fmt.Sprintf("\n-> %s", s.Wrapped.Error()))
	}
	return sb.String()
}

// Is implements error equality checking.
//
// Interface: HasIs.
func (s Defined) Is(target error) bool {
	var err Defined
	if !errors.As(target, &err) {
		return false
	}
	return s.Equal(err)
}

// Unwrap implements error unwrapping for nested errors.
//
// Interface: Unwrap.
func (s Defined) Unwrap() error {
	return s.Wrapped
}

// Wrap returns a new Error with the given err wrapped.
//
// If the given err is also an Defined and the current instance
// is a zero value, just return a copy of the given Defined. This
// allows us to avoid checking this case at every call-site; we
// can just Wrap the error and handle it.
func (s Defined) Wrap(err error) Error {
	if err == nil {
		return s
	}
	if s.IsZero() {
		var ce Defined
		if errors.As(err, &ce) {
			return ce.Copy()
		}
	}
	return Defined{
		Code:      s.Code,
		Extras:    s.Extras,
		Flags:     s.Flags,
		Message:   s.Message,
		Namespace: s.Namespace,
		Wrapped:   err,
	}
}

// Wrapf returns a new Defined with an error created by the given format + args.
func (s Defined) Wrapf(format string, a ...any) Error {
	return s.Wrap(fmt.Errorf(format, a...))
}
