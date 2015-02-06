package handler

import (
	"github.com/materials-commons/config/cfg"
)

// Keeps a list of all the handlers to use.
type multiHandler struct {
	handlers []cfg.Handler
}

// Multi takes a list of Handlers and returns a single Handler that calls them.
// The handlers are called in the order they are specified.
func Multi(handlers ...cfg.Handler) cfg.Handler {
	return &multiHandler{handlers: handlers}
}

// Init initializes each of the handlers. If any of the Handlers returns an error
// then Init returns an error. The results of calling Set or Get if Init returns
// an error are not specified.
func (h *multiHandler) Init() error {
	for _, handler := range h.handlers {
		if err := handler.Init(); err != nil {
			return err
		}
	}
	return nil
}

// Get iterates through each of the handlers in the order given in Multi. It stops
// when one of the handlers returns a value. Get checks each handler to see if it
// should call it with the additional arguments.
func (h *multiHandler) Get(key string, args ...interface{}) (interface{}, error) {
	lengthArgs := len(args)
	if lengthArgs > 0 && !h.Args() {
		return nil, cfg.ErrArgsNotSupported
	}

	for _, handler := range h.handlers {
		switch {
		case lengthArgs != 0 && handler.Args():
			if val, err := handler.Get(key, args...); err == nil {
				return val, nil
			}
		default:
			if val, err := handler.Get(key); err == nil {
				return val, nil
			}
		}
	}
	return nil, cfg.ErrKeyNotFound
}

// Set iterates through each of the handlers in the order given in Multi. It stops
// when one of the handlers successfully sets the key value. Set checks each handler
// to see if it should call it with the additional arguments.
func (h *multiHandler) Set(key string, value interface{}, args ...interface{}) error {
	lengthArgs := len(args)
	if lengthArgs > 0 && !h.Args() {
		return cfg.ErrArgsNotSupported
	}

	for _, handler := range h.handlers {
		switch {
		case lengthArgs != 0 && handler.Args():
			if err := handler.Set(key, value, args...); err == nil {
				return nil
			}
		default:
			if err := handler.Set(key, value); err == nil {
				return nil
			}
		}

	}
	return cfg.ErrKeyNotSet
}

// Args returns true if any of the handlers takes additional arguments.
// Otherwise it returns false.
func (h *multiHandler) Args() bool {
	for _, handler := range h.handlers {
		if handler.Args() {
			return true
		}
	}
	return false
}
