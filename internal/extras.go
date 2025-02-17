package internal

import "time"

// Extras contains common additional info attached to errors.
type Extras struct {
	// Delay is the duration to wait before retrying the failed operation.
	Delay time.Duration `json:"delay,omitempty"`
	// Links to helpful documentation regarding the error.
	Links []string `json:"links,omitempty"`
	// StackTrace of the error.
	StackTrace string `json:"stack_trace,omitempty"`
	// Tags are additional labels that can be used to categorize errors.
	Tags []string `json:"tags,omitempty"`
}

// WithDelay returns a new copy of the Extras with the retry delay.
func (e Extras) WithDelay(delay time.Duration) Extras {
	return Extras{
		Delay:      delay,
		Links:      e.Links,
		StackTrace: e.StackTrace,
		Tags:       e.Tags,
	}
}

// WithStackTrace returns a new copy of the Extras with the stace trace.
func (e Extras) WithStackTrace(trace string) Extras {
	return Extras{
		Delay:      e.Delay,
		Links:      e.Links,
		StackTrace: trace,
		Tags:       e.Tags,
	}
}

// WithLinks returns a new copy of the Extras with the additional links added.
func (e Extras) WithLinks(links ...string) Extras {
	return Extras{
		Delay:      e.Delay,
		Links:      append(e.Links, links...),
		StackTrace: e.StackTrace,
		Tags:       e.Tags,
	}
}

// WithTags returns a new copy of the Canonical with the additional tags added.
func (e Extras) WithTags(tags ...string) Extras {
	return Extras{
		Delay:      e.Delay,
		Links:      e.Links,
		StackTrace: e.StackTrace,
		Tags:       append(e.Tags, tags...),
	}
}
