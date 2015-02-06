package handler

import (
	"bytes"
	"github.com/materials-commons/config/cfg"
	"github.com/materials-commons/config/loader"
	"testing"
)

func TestByNameHandler(t *testing.T) {
	var val interface{}
	hmap := Map()
	l := loader.JSON(bytes.NewReader(jsonTestData))
	hl := Loader(l)
	h := ByName(NameHandler("map", hmap), NameHandler("loader", hl))
	if err := h.Init(); err != nil {
		t.Fatalf("Init failed: %s", err)
	}

	// Get Non existent key
	_, err := h.Get("TEST", "map")
	if err != cfg.ErrKeyNotFound {
		t.Fatalf("Looked up of nonexistent key should have returned ErrKeyNotFound, instead: %s", err)
	}

	// Make sure testkey exists
	if val, err = h.Get("testkey", "loader"); err != nil {
		t.Fatalf("Failed looking up key testkey: %s", err)
	}

	sval := val.(string)
	if sval != "testval" {
		t.Fatalf("testkey should be equal to 'testval', got %s", sval)
	}

	// Set Key
	if err = h.Set("TEST", "TEST1", "map"); err != nil {
		t.Fatalf("Setting key failed: %s", err)
	}

	if val, err = h.Get("TEST", "map"); err != nil {
		t.Fatalf("Failed looking up key TEST: %s", err)
	}

	sval = val.(string)
	if sval != "TEST1" {
		t.Fatalf("TEST should be equal to 'TEST1', got %s", sval)
	}

	// Make sure we can reset key type
	if err = h.Set("TEST", 1, "map"); err != nil {
		t.Fatalf("Setting key failed: %s", err)
	}

	if val, err = h.Get("TEST", "map"); err != nil {
		t.Fatalf("Failed looking up key TEST: %s", err)
	}

	ival := val.(int)
	if ival != 1 {
		t.Fatalf("Test should be be equal to 1, got %d", ival)
	}

	// Make sure that Args is false since none of our handlers takes extra args.
	if !h.Args() {
		t.Fatalf("EnvHandler Args should have returned false.")
	}
}
