package config

import (
	"time"

	"github.com/materials-commons/config/cfg"
	"github.com/materials-commons/config/handler"
)

// Store configuration in environment as specified for 12 Factor Applications:
// http://12factor.net/config. The handler is thread safe and can safely be
// used across multiple go routines.
var TwelveFactor = handler.Env()

// Store configuration in environment, but allow overrides, either by the
// application setting them internally as defaults, or setting them from
// the command line. See http://12factor.net/config. Overrides are an
// extension to the specification. This handler is thread safe and can
// safely be used across multiple go routines.
var TwelveFactorWithOverride = handler.Multi(handler.Sync(handler.Map()), handler.Env())

// std is the package global that all the methods in this file refer to. It provides
// a convenient interface to a package global config.
var std Configer

func init() {
	std = New(TwelveFactor)
	std.Init()
}

// Init initializes the standard Configer using the specified handler. The
// standard configer is a global config that can be conveniently accessed
// from the config package.
func Init(handler cfg.Handler) error {
	std = New(handler)
	return std.Init()
}

// Get gets a key from the standard Configer.
func Get(key string, args ...interface{}) (interface{}, error) {
	return std.Get(key, args...)
}

// GetIntErr gets an integer key from the standard Configer.
func GetIntErr(key string, args ...interface{}) (int, error) {
	return std.GetIntErr(key, args...)
}

// GetInt gets an integer key. It returns the default value of 0 if
// there is an error. GetLastError can be called to see the error.
// If a function is set with SetErrorHandler then the function will
// be called when an error occurs.
func GetInt(key string, args ...interface{}) int {
	return std.GetInt(key, args...)
}

// GetStringErr gets an string key from the standard Configer.
func GetStringErr(key string, args ...interface{}) (string, error) {
	return std.GetStringErr(key, args...)
}

// GetString gets an integer key. It returns the default value of "" if
// there is an error. GetLastError can be called to see the error.
// If a function is set with SetErrorHandler then the function will
// be called when an error occurs.
func GetString(key string, args ...interface{}) string {
	return std.GetString(key, args...)
}

// GetTimeErr gets an time key from the standard Configer.
func GetTimeErr(key string, args ...interface{}) (time.Time, error) {
	return std.GetTimeErr(key, args...)
}

// GetTime gets an integer key. It returns the default value of an
// empty time.Time if there is an error. GetLastError can be
// called to see the error. If a function is set with
// SetErrorHandler then the function will be called when an error occurs.
func GetTime(key string, args ...interface{}) time.Time {
	return std.GetTime(key, args...)
}

// GetBoolErr gets an bool key from the standard Configer.
func GetBoolErr(key string, args ...interface{}) (bool, error) {
	return std.GetBoolErr(key, args...)
}

// GetBool gets an integer key. It returns the default value of false if
// there is an error. GetLastError can be called to see the error. if a
// function is set with SetErrorHandler then the function will be called
// when an error occurs.
func GetBool(key string, args ...interface{}) bool {
	return std.GetBool(key, args...)
}

// GetLastError returns any error that occured when GetInt, GetString,
// GetBool, or GetTime are called. It will return nil if there was
// no error.
func GetLastError() error {
	return std.GetLastError()
}

// SetErrorHandler sets a function to call when GetInt, GetString,
// GetBool, or GetTime return an error. You can use this function
// to handle error in an application specific way. For example if
// an error is fatal you can have this function call os.Exit() or
// panic. Alternatively you can easily log errors with this.
func SetErrorHandler(f cfg.ErrorFunc) {
	std.SetErrorHandler(f)
}

// Set sets key to value in the standard Configer.
func Set(key string, value interface{}, args ...interface{}) error {
	return std.Set(key, value, args...)
}
