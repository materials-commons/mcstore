package cfg

import (
	"errors"
)

var (
	// ErrBadType conversion of value to specified type failed.
	ErrBadType = errors.New("bad type conversion")

	// ErrKeyNotFound specified key not found.
	ErrKeyNotFound = errors.New("key not found")

	// ErrKeyNotSet key could not be set to value
	ErrKeyNotSet = errors.New("key not set")

	// ErrArgsNotSupported set and/or get routine don't support extra arguments
	ErrArgsNotSupported = errors.New("args not supported")

	// ErrBadArgs args passed in are bad
	ErrBadArgs = errors.New("bad args")
)
