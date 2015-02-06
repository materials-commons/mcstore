package handler

import (
	"github.com/materials-commons/config/cfg"
	"sync"
)

// syncHandler holds all the attributes needed to provide
// safe, synchronized access to a handler.
type syncHandler struct {
	handler cfg.Handler
	loaded  bool
	mutex   sync.RWMutex
}

// Sync creates a Handler that can be safely accessed by multiple threads. It
// ensures that the Init method only initializes a handler one time, regardless
// of the number of threads that call it.
func Sync(handler cfg.Handler) *syncHandler {
	return &syncHandler{handler: handler}
}

// Init safely initializes the handler once. If Init has already been successfully called
// additional calls to it don't do anything.
func (h *syncHandler) Init() error {
	defer h.mutex.Unlock()
	h.mutex.Lock()

	switch {
	case h.loaded:
		return nil
	default:
		if err := h.handler.Init(); err != nil {
			return err
		}
	}

	h.loaded = true
	return nil
}

// Get provides synchronized access to key retrieval.
func (h *syncHandler) Get(key string, args ...interface{}) (interface{}, error) {
	defer h.mutex.RUnlock()
	h.mutex.RLock()
	return h.handler.Get(key, args...)
}

// Set provides synchronized access to setting a key.
func (h *syncHandler) Set(key string, value interface{}, args ...interface{}) error {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	return h.handler.Set(key, value, args...)
}

// Args returns true if the handler takes additional arguments.
func (h *syncHandler) Args() bool {
	defer h.mutex.RUnlock()
	h.mutex.RLock()
	return h.handler.Args()
}

// swapHandler allows a handler to be swapped. You need to call Init after
// the handler has been swapped.It is an internal function that is used to
// implement other handlers.
func (h *syncHandler) swapHandler(newHandler cfg.Handler) {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	h.handler = newHandler
	h.loaded = false
}

// swapHandlerInit allows a handler to be swapped. It calls Init and returns the
// results. It is an internal function that is used to implement other handlers.
func (h *syncHandler) swapHandlerInit(newHandler cfg.Handler) error {
	defer h.mutex.Unlock()
	h.mutex.Lock()
	h.handler = newHandler
	h.loaded = false
	if err := h.handler.Init(); err != nil {
		return err
	}
	h.loaded = true
	return nil
}
