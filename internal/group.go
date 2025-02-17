package internal

import (
	"errors"
	"fmt"
	"strings"
)

// NewGroup creates a new *Group with sane defaults.
func NewGroup(errs ...error) *Group {
	eg := &Group{
		Errors:    make([]Error, 0, len(errs)),
		Formatter: GroupFormatterDefault,
	}
	eg.Append(errs...)
	return eg
}

// Group stores multiple Canonical instances.
//
// TODO(ahawker) Flatten JSON output to a single error when group only has one.
type Group struct {
	// Errors in the group.
	Errors []Error `json:"errors"`
	// Formatter to convert error group to string representation.
	Formatter GroupFormatter `json:"-"`
}

// Append adds a new error to the group.
//
// If one of the errors is a Group, it will be flatten into this group.
// If the given error is not an Error, it will be wrapped
// with 'ErrUndefined'.
func (g *Group) Append(errs ...error) {
	for _, err := range errs {
		if err == nil {
			continue
		}

		// When given an error that's a group, we want to flatten & merge
		// the items.
		var eg *Group
		if errors.As(err, &eg) {
			g.Append(eg.Slice()...)
			continue
		}

		// When given an error that isn't Canonical, wrap it.
		// TODO (ahawker) Might have broke this.
		var e Error
		if !errors.As(err, &e) {
			e = ErrUnknown.Wrap(err)
		}

		if e == nil {
			continue
		}

		g.Errors = append(g.Errors, e)
	}
}

// Empty will return true if the group is empty.
func (g *Group) Empty() bool {
	if g == nil {
		return true
	}
	return len(g.Errors) == 0
}

// ErrorOrNil returns an error interface if this Canonical represents
// a list of errors, or returns nil if the list of errors is empty. This
// function is useful at the end of accumulation to make sure that the value
// returned represents the existence of errors.
func (g *Group) ErrorOrNil() error {
	if g == nil {
		return nil
	}
	if g.Errors == nil || len(g.Errors) == 0 {
		return nil
	}
	return g
}

// Slice returns the group as a slice of error.
func (g *Group) Slice() []error {
	errs := make([]error, 0, len(g.Errors))
	for _, e := range g.Errors {
		errs = append(errs, e)
	}
	return errs
}

// Unwrap returns the next error in the group or nil if there are no more errors.
//
// Interface: errors.Unwrap, HasUnwrap.
func (g *Group) Unwrap() error {
	// If we have no errors then we do nothing
	if g == nil || len(g.Errors) == 0 {
		return nil
	}

	// If we have exactly one error, we can just return that directly.
	if len(g.Errors) == 1 {
		return g.Errors[0]
	}

	// Shallow copy the errors slice.
	errs := make([]Error, len(g.Errors))
	copy(errs, g.Errors)
	return chain(errs)
}

// String returns the string value of the Group.
//
// Interface: fmt.Stringer.
func (g *Group) String() string {
	return g.Error()
}

// Canonical string value of the Group struct.
//
// Interface: error.
func (g *Group) Error() string {
	return g.Formatter(g.Errors)
}

// Len returns the number of errors in the group.
//
// Interface: sort.Interface.
func (g *Group) Len() int {
	return len(g.Errors)
}

// Less determines order for sorting a group.
//
// Interface: sort.Interface.
func (g *Group) Less(i, j int) bool {
	return g.Errors[i].Error() < g.Errors[j].Error()
}

// Swap moves errors in the group during sorting.
//
// Interface: sort.Interface.
func (g *Group) Swap(i, j int) {
	g.Errors[i], g.Errors[j] = g.Errors[j], g.Errors[i]
}

// chain implements the interfaces necessary for errors.Is/As/Unwrap to
// work in a deterministic way. Is/As/Canonical will work on the error stored
// in the slice at index zero. Upon an Unwrap call, we will return a chain
// with a new slice with an index shifted by one.
//
// Based on ideas from https://github.com/hashicorp/go-multierror.
type chain []Error

// Canonical implements the error interface.
func (e chain) Error() string {
	if len(e) == 0 {
		return ""
	}
	return e[0].Error()
}

// Unwrap implements errors.Unwrap by returning the next error in the
// chain or nil if there are no more errors.
func (e chain) Unwrap() error {
	if len(e) <= 1 {
		return nil
	}
	return e[1:]
}

// As implements errors.As by attempting to map to the current value.
func (e chain) As(target any) bool {
	if len(e) == 0 {
		return false
	}
	return errors.As(e[0], target)
}

// Is implements errors.Is by comparing the current value directly.
func (e chain) Is(target error) bool {
	if len(e) == 0 {
		return false
	}
	return errors.Is(e[0], target)
}

// GroupFormatter is a function callback that is called by Group to
// turn the list of errors into a string.
type GroupFormatter func([]Error) string

// GroupFormatterDefault is a basic Formatter that outputs the number of errors
// that occurred along with a bullet point list of the errors.
func GroupFormatterDefault(errors []Error) string {
	switch len(errors) {
	case 0:
		return ""
	case 1:
		return errors[0].Error()
	default:
		points := make([]string, len(errors))
		for i, err := range errors {
			points[i] = fmt.Sprintf("* %s", err)
		}
		return fmt.Sprintf("\n%s\n\n", strings.Join(points, "\n"))
	}
}
