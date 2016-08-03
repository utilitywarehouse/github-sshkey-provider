package simplecache

import "errors"

var (
	// ErrValueHasNotChanged is returned when trying to set a value in the
	// cache and the value is already there.
	ErrValueHasNotChanged = errors.New("Value has not changed")
)

// Interface to be implemented by simplecache variations.
type Interface interface {
	Set(string, string) error
	Get(string) (string, error)
	Flush() error
}
