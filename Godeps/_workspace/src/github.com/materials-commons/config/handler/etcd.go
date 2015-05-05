package handler

import (
	"github.com/coreos/go-etcd/etcd"
	"github.com/materials-commons/config/cfg"
)

type etcdHandler struct {
	client *etcd.Client // Etcd client connection
}

// Etcd returns a Handler that accesses keys that are stored in Etcd.
func Etcd(client *etcd.Client) cfg.Handler {
	return &etcdHandler{
		client: client,
	}
}

// Init doesn't do anything
func (h *etcdHandler) Init() error {
	return nil
}

// Get retrieves a key from etcd. Any error calling etcd is mapped to ErrKeyNotFound.
func (h *etcdHandler) Get(key string, args ...interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, cfg.ErrArgsNotSupported
	}

	resp, err := h.client.Get(key, false, false)
	if err != nil {
		return nil, cfg.ErrKeyNotFound
	}

	return resp.Node.Value, nil
}

// Set will set a key value. Values will be converted and stored as strings. It will
// attempt to convert the value to a string before it stores it. If the value cannot
// be converted to a string it will return ErrBadType. Any error calling etcd will
// be returned as ErrKeyNotSet.
func (h *etcdHandler) Set(key string, value interface{}, args ...interface{}) error {
	if len(args) != 0 {
		return cfg.ErrArgsNotSupported
	}

	sval, err := cfg.ToString(value)
	if err != nil {
		return cfg.ErrBadType
	}

	if _, err := h.client.Set(key, sval, 0); err != nil {
		return cfg.ErrKeyNotSet
	}

	return nil
}

// Args returns false. This handler doesn't accept addtional arguments.
func (h *etcdHandler) Args() bool {
	return false
}
