package handler

import (
	"testing"
)

func TestLowercaseKey(t *testing.T) {
	var val interface{}
	underlyingHandler := Map()
	h := LowercaseKey(underlyingHandler)

	err := h.Init()
	if err != nil {
		t.Fatalf("Init failed: %s", err)
	}

	if err = h.Set("tEsTKey", "value1"); err != nil {
		t.Fatalf("Setting key failed: %s", err)
	}

	if val, err = h.Get("testkey"); err != nil {
		t.Fatalf("Failed looing up testkey: %s", err)
	}

	sval := val.(string)
	if sval != "value1" {
		t.Fatalf("expected testkey value 'value1', got %s", sval)
	}
}
