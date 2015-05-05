package handler

import (
	"fmt"
	"strings"

	"github.com/materials-commons/config/cfg"
)

// KeyFunc is the type of a function to apply to a key to produce a transformed one.
type KeyFunc func(key string) (string, error)

// ValueFunc is the type of a function to apply to a value to produce a transformed one.
type ValueFunc func(value interface{}) (interface{}, error)

type applyHandler struct {
	keyFunc   KeyFunc
	valueFunc ValueFunc
	handler   cfg.Handler
}

// Apply creates a new Apply Handler. An Apply handler applies the given key and
// value funcs to the key/value before calling the underlying handler. The funcs
// are only called if they aren't nil.
//
// An Apply handler can be used to transform values. For example, if keys should
// always be in lower case you can apply a key func to lower case the keys.
func Apply(keyFunc KeyFunc, valueFunc ValueFunc, handler cfg.Handler) cfg.Handler {
	return &applyHandler{
		keyFunc:   keyFunc,
		valueFunc: valueFunc,
		handler:   handler,
	}
}

// Init initializes the underlying handler.
func (h *applyHandler) Init() error {
	return h.handler.Init()
}

// Get first transforms the key by calling the KeyFunc if it is not nil. It then
// calls the underlying handler with the (optionally) transformed key.
func (h *applyHandler) Get(key string, args ...interface{}) (interface{}, error) {
	var err error
	if h.keyFunc != nil {
		key, err = h.keyFunc(key)
		if err != nil {
			return nil, err
		}
	}
	return h.handler.Get(key, args...)
}

// Set first tranforms the key, and the value if one or both of the KeyFunc and
// ValueFunc funcs aren't nil. It then calls the underlying handler with the
// (optionally) transformed key and value.
func (h *applyHandler) Set(key string, value interface{}, args ...interface{}) error {
	var err error
	if h.keyFunc != nil {
		key, err = h.keyFunc(key)
		if err != nil {
			return err
		}
	}

	if h.valueFunc != nil {
		value, err = h.valueFunc(value)
		if err != nil {
			return err
		}
	}
	return h.handler.Set(key, value, args...)
}

// Args returns the results of calling the underlying handler's Args method.
func (h *applyHandler) Args() bool {
	return h.handler.Args()
}

// ApplyKey creates a new Apply handler with the specified KeyFunc.
func ApplyKey(keyFunc KeyFunc, handler cfg.Handler) cfg.Handler {
	return Apply(keyFunc, nil, handler)
}

// ApplyValue creates a new Apply handler with the specified ValueFunc.
func ApplyValue(valueFunc ValueFunc, handler cfg.Handler) cfg.Handler {
	return Apply(nil, valueFunc, handler)
}

// LowercaseKey creates a new Apply handler that will lowercase keys.
func LowercaseKey(handler cfg.Handler) cfg.Handler {
	return Apply(func(key string) (string, error) { return strings.ToLower(key), nil }, nil, handler)
}

// UppercaseKey creates a new Apply handler that will uppercase keys.
func UppercaseKey(handler cfg.Handler) cfg.Handler {
	return Apply(func(key string) (string, error) { return strings.ToUpper(key), nil }, nil, handler)
}

// PrefixKey will prefix all keys with the given prefix string.
func PrefixKey(handler cfg.Handler, prefix string) cfg.Handler {
	return Apply(func(key string) (string, error) { return fmt.Sprintf("%s%s", prefix, key), nil }, nil, handler)
}
