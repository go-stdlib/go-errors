package internal

import "strconv"

const (
	// FlagUnknown is set to represent unknown/unregistered errors.
	FlagUnknown Flags = 1 << iota
	// FlagRetryable is set to represent errors that can be retried.
	FlagRetryable
	// FlagTimeout is set to represent errors indicating a timeout occurred.
	FlagTimeout
)

// Flags is a `uint8` with helper methods for bitwise operations
// to store additional properties about errors.
type Flags uint8

// MarshalText implements the text marshaller method.
func (f Flags) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

// String returns the Flags in binary string (001101010) form.
func (f Flags) String() string {
	return strconv.FormatUint(uint64(f), 2)
}

// Clear given bits from the current mask and return a new copy.
func (f Flags) Clear(bits Flags) Flags {
	return f &^ bits
}

// Has checks if bits are set.
func (f Flags) Has(bits Flags) bool {
	return f&bits != 0
}

// Set bits in the current mask and return a new copy.
func (f Flags) Set(bits Flags) Flags {
	return f | bits
}

// Toggle bits on/off and return a new copy.
func (f Flags) Toggle(bits Flags) Flags {
	return f ^ bits
}
