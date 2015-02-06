package handler

import (
	"github.com/materials-commons/config/cfg"
	"testing"
)

func TestMapHandler(t *testing.T) {
	var val interface{}
	h := Map()
	if err := h.Init(); err != nil {
		t.Fatalf("Init failed: %s", err)
	}

	// Get Non existent key
	_, err := h.Get("TEST")
	if err != cfg.ErrKeyNotFound {
		t.Fatalf("Looked up of nonexistent key should have returned ErrKeyNotFound, instead: %s", err)
	}

	// Set Key
	if err = h.Set("TEST", "TEST1"); err != nil {
		t.Fatalf("Setting key failed: %s", err)
	}

	if val, err = h.Get("TEST"); err != nil {
		t.Fatalf("Failed looking up key TEST: %s", err)
	}

	sval := val.(string)
	if sval != "TEST1" {
		t.Fatalf("TEST should be equal to 'TEST1', got %s", sval)
	}

	// Make sure we can reset key type
	if err = h.Set("TEST", 1); err != nil {
		t.Fatalf("Setting key failed: %s", err)
	}

	if val, err = h.Get("TEST"); err != nil {
		t.Fatalf("Failed looking up key TEST: %s", err)
	}

	ival := val.(int)
	if ival != 1 {
		t.Fatalf("Test should be be equal to 1, got %d", ival)
	}

	// Make sure that Args is false
	if h.Args() {
		t.Fatalf("EnvHandler Args should have returned false.")
	}

	// Make sure that calls with extra args fail.
	if _, err := h.Get("TEST", "EXTRA_ARG"); err != cfg.ErrArgsNotSupported {
		t.Fatalf("Get with extra args returned wrong err: %s", err)
	}

	if err := h.Set("TEST", "BLAH", "EXTRA_ARG"); err != cfg.ErrArgsNotSupported {
		t.Fatalf("Set with extra args returned wrong err: %s", err)
	}

	// Make sure that the value wasn't set
	if val, err = h.Get("TEST"); err != nil {
		t.Fatalf("Failed looking up key ENV_TEST2: %s", err)
	}

	if val == "BLAH" {
		t.Fatalf("Set with extra args set the value when it shouldn't have")
	}
}
