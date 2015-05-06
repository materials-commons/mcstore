package app

import (
	"errors"
	"fmt"
)

var (
	// ErrNotFound Item not found
	ErrNotFound = errors.New("not found")

	// ErrInvalid Invalid request
	ErrInvalid = errors.New("invalid")

	// ErrExists Item already exists
	ErrExists = errors.New("exists")

	// ErrNoAccess Access to item not allowed
	ErrNoAccess = errors.New("no access")

	// ErrInternal Internal fatal error
	ErrInternal = errors.New("internal error")

	// ErrCreate Create of object failed
	ErrCreate = errors.New("unable to create")

	// ErrInUse object is locked and in use by someone else
	ErrInUse = errors.New("in use")
)

// Error holds the error code and additional messages.
type Error struct {
	Err     error  // Error code
	Message string // Message related to error
}

// Implement error interface.
func (e *Error) Error() string {
	if e.Message != "" {
		return e.Err.Error() + ":" + e.Message
	}

	return e.Err.Error()
}

// newError creates a new instance of an Error.
func newError(err error, msg string) *Error {
	return &Error{
		Message: msg,
		Err:     err,
	}
}

// Is returns true if the particular error code in an Error
// is equal to the expected error. This is useful comparing
// without having to cast and unpack.
func Is(err error, what error) bool {
	if e, ok := err.(*Error); ok {
		return e.Err == what
	}

	return err == what
}

// Errorf takes and error, a message string and a set of arguments and produces
// a new Error.
func Errorf(err error, message string, args ...interface{}) *Error {
	msg := fmt.Sprintf(message, args...)
	return newError(err, msg)
}

// Panicf will format the message and call panic.
func Panicf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	panic(msg)
}
