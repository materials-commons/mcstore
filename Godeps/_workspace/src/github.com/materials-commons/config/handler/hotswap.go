package handler

import (
	"github.com/materials-commons/config/cfg"
)

// HotSwapHandler holds a swappable handler. It is safe to swap
// this handler at run even if used by multiple threads.
type HotSwapHandler struct {
	handler *syncHandler
}

// HotSwap creates a hot swappable handler.
func HotSwap(handler cfg.Handler) *HotSwapHandler {
	return &HotSwapHandler{handler: Sync(handler)}
}

// Init initializes the handler. It is thread safe.
func (h *HotSwapHandler) Init() error {
	return h.handler.Init()
}

// Get retrieves key values. It is thread safe.
func (h *HotSwapHandler) Get(key string, args ...interface{}) (interface{}, error) {
	return h.handler.Get(key, args...)
}

// Set sets a key to a value. It is thread safe.
func (h *HotSwapHandler) Set(key string, value interface{}, args ...interface{}) error {
	return h.handler.Set(key, value, args...)
}

// Args returns true if the underlying handler supports multiple args. It is thread safe.
func (h *HotSwapHandler) Args() bool {
	return h.handler.Args()
}

// SwapHandler allows a handler to be swapped in a thread safe fashion. You need
// to call init before using the handler.
func (h *HotSwapHandler) SwapHandler(newHandler cfg.Handler) {
	h.handler.swapHandler(newHandler)
}

// SwapHandlerInit allows a handler to be swapped in a thread safe fashion. It
// also calls Init and returns the result.
func (h *HotSwapHandler) SwapHandlerInit(newHandler cfg.Handler) error {
	return h.handler.swapHandlerInit(newHandler)
}
