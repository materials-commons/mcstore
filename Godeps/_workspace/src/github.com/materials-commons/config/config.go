package config

import (
	"time"

	"github.com/materials-commons/config/cfg"
)

// LoggerFunc is the function to call to log config events.
type LoggerFunc func(event Event, err error, args ...interface{})

// A Configer is a configuration object that can store and retrieve key/value pairs.
type Configer interface {
	cfg.Initer
	cfg.Getter
	cfg.TypeGetterErr
	cfg.TypeGetterDefault
	cfg.Setter
	SetHandler(handler cfg.Handler)
	SetHandlerInit(handler cfg.Handler) error
	SetLogger(l LoggerFunc)
}

// config is a private type for storing configuration information.
type config struct {
	handler   cfg.Handler   // Handler being used
	lastError error         // Last error see on get
	efunc     cfg.ErrorFunc // Error function to call see SetErrorHandler
	lfunc     LoggerFunc    // Logger function to call
}

// New creates a new Configer instance that uses the specified Handler for
// key/value retrieval and storage.
func New(handler cfg.Handler) Configer {
	return &config{handler: handler}
}

// Init initializes the Configer. It should be called before retrieving
// or setting keys.
func (c *config) Init() error {
	err := c.handler.Init()
	c.log(INIT, err)
	return err
}

// Get returns the value for a key. It can return any value type.
func (c *config) Get(key string, args ...interface{}) (interface{}, error) {
	value, err := c.handler.Get(key, args...)
	c.lastError = err
	c.log(GET, err, toVarArgs(args, key)...)
	return value, err
}

// GetIntErr returns an integer value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetIntErr(key string, args ...interface{}) (int, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return 0, err
	}
	ival, err := cfg.ToInt(val)
	c.log(TOINT, err, toVarArgs(args, val, ival)...)
	return ival, err
}

// GetStringErr returns an string value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetStringErr(key string, args ...interface{}) (string, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return "", err
	}
	sval, err := cfg.ToString(val)
	c.log(TOSTRING, err, toVarArgs(args, val, sval)...)
	return sval, err
}

// GetTimeErr returns an time.Time value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetTimeErr(key string, args ...interface{}) (time.Time, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return time.Time{}, err
	}
	tval, err := cfg.ToTime(val)
	c.log(TOTIME, err, toVarArgs(args, val, tval)...)
	return tval, err
}

// GetBoolErr returns an bool value for a key. See TypeGetter interface for
// error codes.
func (c *config) GetBoolErr(key string, args ...interface{}) (bool, error) {
	val, err := c.Get(key, args...)
	if err != nil {
		return false, err
	}
	bval, err := cfg.ToBool(val)
	c.log(TOBOOL, err, toVarArgs(args, val, bval)...)
	return bval, err
}

// GetInt gets an integer key. It returns the default value of 0 if
// there is an error. GetLastError can be called to see the error.
// If a function is set with SetErrorHandler then the function will
// be called when an error occurs.
func (c *config) GetInt(key string, args ...interface{}) int {
	val, err := c.GetIntErr(key, args...)
	if err != nil {
		c.errorHandler(key, err, args...)
	}
	return val
}

// GetString gets an integer key. It returns the default value of "" if
// there is an error. GetLastError can be called to see the error.
// If a function is set with SetErrorHandler then the function will
// be called when an error occurs.
func (c *config) GetString(key string, args ...interface{}) string {
	val, err := c.GetStringErr(key, args...)
	if err != nil {
		c.errorHandler(key, err, args...)
	}
	return val
}

// GetTime gets an integer key. It returns the default value of an
// empty time.Time if there is an error. GetLastError can be
// called to see the error. If a function is set with
// SetErrorHandler then the function will be called when an error occurs.
func (c *config) GetTime(key string, args ...interface{}) time.Time {
	val, err := c.GetTimeErr(key, args...)
	if err != nil {
		c.errorHandler(key, err, args...)
	}
	return val
}

// GetBool gets an integer key. It returns the default value of false if
// there is an error. GetLastError can be called to see the error. if a
// function is set with SetErrorHandler then the function will be called
// when an error occurs.
func (c *config) GetBool(key string, args ...interface{}) bool {
	val, err := c.GetBoolErr(key, args...)
	if err != nil {
		c.errorHandler(key, err, args...)
	}
	return val
}

// GetLastError returns any error that occured when GetInt, GetString,
// GetBool, or GetTime are called. It will return nil if there was
// no error.
func (c *config) GetLastError() error {
	return c.lastError
}

// SetErrorHandler sets a function to call when GetInt, GetString,
// GetBool, or GetTime return an error. You can use this function
// to handle error in an application specific way. For example if
// an error is fatal you can have this function call os.Exit() or
// panic. Alternatively you can easily log errors with this.
func (c *config) SetErrorHandler(f cfg.ErrorFunc) {
	c.efunc = f
}

// errorHandler calls the error function set with SetErrorHandler.
func (c *config) errorHandler(key string, err error, args ...interface{}) {
	if c.efunc != nil {
		c.efunc(key, err, args...)
	}
}

// SetHandler changes the handler for a Configer. If this method is called
// then you must call Init before accessing any of the keys.
func (c *config) SetHandler(handler cfg.Handler) {
	c.handler = handler
}

// SetHandlerInit changes the handler for a Configer. It also immediately calls
// Init and returns the error from this call.
func (c *config) SetHandlerInit(handler cfg.Handler) error {
	c.handler = handler
	return c.Init()
}

// Set sets key to value. See Setter interface for error codes.
func (c *config) Set(key string, value interface{}, args ...interface{}) error {
	err := c.handler.Set(key, value, args...)
	c.log(SET, err, toVarArgs(args, key, value)...)
	return err
}

// SetLogger sets the logging function to call for configuration events.
func (c *config) SetLogger(l LoggerFunc) {
	c.lfunc = l
}

// log calls the logging function set with SetLogger.
func (c *config) log(event Event, err error, args ...interface{}) {
	if c.lfunc != nil {
		c.lfunc(event, err, args...)
	}
}

// toVarArgs prepends additionalArgs to args. Used to create a
// new set of variable arguments.
func toVarArgs(args []interface{}, additionalArgs ...interface{}) []interface{} {
	return append(additionalArgs, args...)
}
