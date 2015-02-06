package handler

import (
	"github.com/materials-commons/config/cfg"
)

type prioritizedHandler struct {
	byName     cfg.Handler
	byPosition cfg.Handler
}

// Prioritized creates a new Prioritized Handler. A Prioritized Handler performs
// look ups in the order they were given. In addition the name of a handler can
// be passed as the last argument to a set or get. If this is done, then the
// named handler is used.
func Prioritized(handlers ...*HandlerName) cfg.Handler {
	var hn = HandlerNames(handlers)
	phandler := &prioritizedHandler{
		byName:     ByName(handlers...),
		byPosition: Multi(hn.ToHandlers()...),
	}
	return phandler
}

// Init initializes each of the handlers. If any handlers Init method returns
// an error then initialization stops and the error is returned.
func (h *prioritizedHandler) Init() error {
	// Only need to initialize the handlers once. We can choose to do
	// this with either set of handlers. We'll do the Multi handler.
	return h.byName.Init()
}

// Get looks up the given key. The first optional arg is the name of the handler to
// use. If name is given, then use the ByName handler, otherwise use the Multi handler.
func (h *prioritizedHandler) Get(key string, args ...interface{}) (interface{}, error) {
	switch length := len(args); length {
	case 0:
		return h.byPosition.Get(key)
	default:
		return h.byName.Get(key, args...)
	}
}

// Set sets the value of the given key. The first optional arg is the name of the handler
// to use. If name is given then use the ByName handler, otherwise use the Multi handler.
func (h *prioritizedHandler) Set(key string, value interface{}, args ...interface{}) error {
	switch length := len(args); length {
	case 0:
		return h.byPosition.Set(key, value)
	default:
		return h.byName.Set(key, value, args...)
	}
}

// Args returns true. Prioritized handlers always accept an optional argument that
// is the name of the handler to use.
func (h *prioritizedHandler) Args() bool {
	return true
}
