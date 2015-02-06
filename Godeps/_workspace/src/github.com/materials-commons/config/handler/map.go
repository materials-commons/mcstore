package handler

import (
	"github.com/materials-commons/config/cfg"
)

type mapHandler struct {
	values map[string]interface{}
}

// MapUse creates a Map handler that is preloaded with the hashmap passed in.
func MapUse(values map[string]interface{}) cfg.Handler {
	return &mapHandler{values: values}
}

// Map creates a handler that stores all values in a hashmap. It is commonly used
// as a component to build more complex handlers.
func Map() cfg.Handler {
	return &mapHandler{values: make(map[string]interface{})}
}

// Init initializes the handler.
func (h *mapHandler) Init() error {
	return nil
}

// Get retrieves a keys value.
func (h *mapHandler) Get(key string, args ...interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, cfg.ErrArgsNotSupported
	}
	val, found := h.values[key]
	if !found {
		return val, cfg.ErrKeyNotFound
	}
	return val, nil
}

// Set sets the value of keys. You can create new keys, or modify existing ones.
// Values are not persisted across runs.
func (h *mapHandler) Set(key string, value interface{}, args ...interface{}) error {
	if len(args) != 0 {
		return cfg.ErrArgsNotSupported
	}
	h.values[key] = value
	return nil
}

// Args returns false. This handler doesn't accept additional arguments.
func (h *mapHandler) Args() bool {
	return false
}
