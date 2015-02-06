package handler

import (
	"github.com/materials-commons/config/cfg"
	"os"
)

// EnvUpper returns a Env handler that first upper cases all keys before
// getting or setting them.
func EnvUpper() cfg.Handler {
	return UppercaseKey(Env())
}

type envHandler struct{}

// Env returns a Handler that access keys that are environment variables.
func Env() cfg.Handler {
	return &envHandler{}
}

// Init initializes access to the environment.
func (h *envHandler) Init() error {
	return nil
}

// Get retrieves a environment variable.
func (h *envHandler) Get(key string, args ...interface{}) (interface{}, error) {
	if len(args) != 0 {
		return "", cfg.ErrArgsNotSupported
	}
	val := os.Getenv(key)
	if val == "" {
		return val, cfg.ErrKeyNotFound
	}
	return val, nil
}

// Set sets an environment variable. Values will be stored as strings. It will
// convert the value to a string before it attempts to store it. If the value
// cannot be converted to a string it returns ErrBadType.
func (h *envHandler) Set(key string, value interface{}, args ...interface{}) error {
	if len(args) != 0 {
		return cfg.ErrArgsNotSupported
	}
	sval, err := cfg.ToString(value)
	if err != nil {
		return cfg.ErrBadType
	}

	err = os.Setenv(key, sval)
	if err != nil {
		return cfg.ErrKeyNotSet
	}

	return nil
}

// Args returns false. This handler doesn't accept additional arguments.
func (h *envHandler) Args() bool {
	return false
}
