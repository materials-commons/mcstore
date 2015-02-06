package cfg

import (
	"time"
)

// A Getter retrieves values for keys. It returns true if
// a value was found, and false otherwise.
type Getter interface {
	Get(key string, args ...interface{}) (interface{}, error)
}

// A TypeGetterErr performs type safe conversion and retrieval of key values. The routines
// have return ErrBadType if the value doesn't match or can't be converted to the
// getter method called. ErrKeyNotFound is returned if the specified key is not found.
type TypeGetterErr interface {
	GetIntErr(key string, args ...interface{}) (int, error)
	GetStringErr(key string, args ...interface{}) (string, error)
	GetTimeErr(key string, args ...interface{}) (time.Time, error)
	GetBoolErr(key string, args ...interface{}) (bool, error)
}

// ErrorFunc is the signature of the function to call when using
// one of the TypeGetterDefault calls and the call detects an error.
type ErrorFunc func(key string, err error, args ...interface{})

// TypeGetterDefault performs type safe conversion and retrieval of key values.
// The routines mask the error and instead return a default value if an error
// did occur. You can check for the error by calling GetLastError. Alternatively
// you can call SetErrorHandler to a function to call when one of the getters
// sees an error.
type TypeGetterDefault interface {
	GetInt(key string, args ...interface{}) int        // return 0 when error occurs
	GetString(key string, args ...interface{}) string  // return "" when error occurs
	GetTime(key string, args ...interface{}) time.Time // return empty time.Time when error occurs
	GetBool(key string, args ...interface{}) bool      // return false when error occurs
	GetLastError() error                               // Return last error from above Gets
	SetErrorHandler(f ErrorFunc)                       // Called when one of the above Gets sees an error
}

// A Setter stores a value for a key. It returns nil on success and ErrKeyNotSet on failure.
type Setter interface {
	Set(key string, value interface{}, args ...interface{}) error
}

// A Loader loads values into out.
type Loader interface {
	Load(out interface{}) error
}

// A Initer initializes a data structure.
type Initer interface {
	Init() error
}

// A Handler defines the interface to different types of stores. The Init method should
// always be called before attempting to get or set values for a handler.
type Handler interface {
	Initer
	Getter
	Setter
	Args() bool
}
