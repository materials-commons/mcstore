package handler

import (
	"github.com/materials-commons/config/cfg"
)

// A HandlerName associates a name with a given handler.
type HandlerName struct {
	Name    string
	Handler cfg.Handler
}

// NameHandler is a convenience function for creating new instances
// of a HandlerName type.
func NameHandler(name string, handler cfg.Handler) *HandlerName {
	return &HandlerName{
		Name:    name,
		Handler: handler,
	}
}

// HandlerNames is list of HandlerName. The type allows us to add
// convenience functions to lists of HandlerName.
type HandlerNames []*HandlerName

// ToHandlers returns the handlers in an array of HandlerName.
func (hns HandlerNames) ToHandlers() []cfg.Handler {
	var handlers = make([]cfg.Handler, len(hns))
	for i, hn := range hns {
		handlers[i] = hn.Handler
	}
	return handlers
}

// ToNames returns the name in an arry of HandlerName.
func (hns HandlerNames) ToNames() []string {
	var names = make([]string, len(hns))
	for i, hn := range hns {
		names[i] = hn.Name
	}
	return names
}

// ToMap returns a map from an array of HandlerName.
func (hns HandlerNames) ToMap() map[string]*HandlerName {
	hnMap := make(map[string]*HandlerName)

	for _, nh := range hns {
		hnMap[nh.Name] = nh
	}

	return hnMap
}

type byNameHandler struct {
	handlers map[string]*HandlerName
}

// ByName creates a new handler that maps a list of handlers to their name. You
// can then do get and set operations against a named handler.
func ByName(handlers ...*HandlerName) cfg.Handler {
	nhandler := &byNameHandler{
		handlers: HandlerNames(handlers).ToMap(),
	}
	return nhandler
}

// Init initializes each of the handlers. If any of the handlers Init method returns
// an error then initialization stops and the error is returned.
func (h *byNameHandler) Init() error {
	for _, handler := range h.handlers {
		if err := handler.Handler.Init(); err != nil {
			return err
		}
	}
	return nil
}

// Get looks up a handler and returns the value of the key given. The name of
// the handler is the first argument in args. Additional args after the first
// are passed through to the given handler.
func (h *byNameHandler) Get(key string, args ...interface{}) (interface{}, error) {
	switch length := len(args); length {
	case 0:
		return nil, cfg.ErrBadArgs
	default:
		handler, err := h.getNamedHandler(args[0])
		switch {
		case err != nil:
			return nil, err
		case length == 1:
			return handler.Get(key)
		default:
			otherArgs := args[1:]
			return handler.Get(key, otherArgs...)
		}
	}
}

// Set lookup up a handler and sets the value for the given key. The name of
// the handler is the first argument in args. Additional args after the first
// are passed through to the given handler.
func (h *byNameHandler) Set(key string, value interface{}, args ...interface{}) error {
	switch length := len(args); length {
	case 0:
		return cfg.ErrBadArgs
	default:
		handler, err := h.getNamedHandler(args[0])
		switch {
		case err != nil:
			return err
		case length == 1:
			return handler.Set(key, value)
		default:
			otherArgs := args[1:]
			return handler.Set(key, value, otherArgs...)
		}
	}
}

// getNamedHandler takes the name attempts to cast it to a string and looks up
// the named handler.
func (h *byNameHandler) getNamedHandler(name interface{}) (cfg.Handler, error) {
	if handlerName, ok := name.(string); ok {
		handler, found := h.handlers[handlerName]
		if !found {
			return nil, cfg.ErrBadArgs
		}
		return handler.Handler, nil
	}
	return nil, cfg.ErrBadArgs
}

// Args always returns true. A ByName handler always takes at least one additional
// argument, which is the name of the handler.
func (h *byNameHandler) Args() bool {
	return true
}
