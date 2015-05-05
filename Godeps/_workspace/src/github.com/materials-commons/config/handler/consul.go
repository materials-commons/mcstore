package handler

import (
	"bytes"
	"encoding/gob"

	consul "github.com/hashicorp/consul/api"
	"github.com/materials-commons/config/cfg"
)

type consulHandler struct {
	client *consul.Client // consul client connection
}

// Consul returns a Handler that accesses keys stored in Consul.
func Consul(client *consul.Client) cfg.Handler {
	return &consulHandler{
		client: client,
	}
}

// Init doesn't do anything. The connection to consul is setup
// in the Consul call.
func (h *consulHandler) Init() error {
	return nil
}

// Get retrieves a key from Consul. Any error calling Consul is mapped to ErrKeyNotFound.
func (h *consulHandler) Get(key string, args ...interface{}) (interface{}, error) {
	if len(args) != 0 {
		return nil, cfg.ErrArgsNotSupported
	}

	kv, _, err := h.client.KV().Get(key, nil)
	if err != nil {
		return nil, cfg.ErrKeyNotFound
	}

	return kv.Value, nil
}

// Set will set a key value. Values will be converted and stored as bytes. It will
// attempt to convert the value to bytes before it stores it. If the value cannot
// be converted to bytes it will return ErrBadType. Any error calling etcd will
// be returned as ErrKeyNotSet.
func (h *consulHandler) Set(key string, value interface{}, args ...interface{}) error {
	if len(args) != 0 {
		return cfg.ErrArgsNotSupported
	}

	asBytes, err := toBytes(value)
	if err != nil {
		return cfg.ErrBadType
	}

	kv := &consul.KVPair{
		Key:   key,
		Value: asBytes,
	}

	if _, err := h.client.KV().Put(kv, nil); err != nil {
		return cfg.ErrKeyNotSet
	}
	return nil
}

// Args returns false. This handler doesn't accept addtional arguments.
func (h *consulHandler) Args() bool {
	return false
}

// toBytes will convert a value to bytes buffer.
func toBytes(value interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
