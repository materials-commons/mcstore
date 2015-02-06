package loader

import (
	"bytes"
	"testing"
)

var jsonTestData = []byte(`
{
  "stringkey": "stringvalue",
  "boolkey": true,
  "intkey": 123
}`)

func TestJSONLoader(t *testing.T) {
	var (
		vals  map[string]interface{}
		val   interface{}
		found bool
	)

	l := JSON(bytes.NewReader(jsonTestData))
	err := l.Load(&vals)
	if err != nil {
		t.Fatalf("Failed loading JSON: %s", err)
	}

	if val, found = vals["stringkey"]; !found {
		t.Fatalf("Failed to lookup key 'stringkey'")
	}

	sval := val.(string)
	if sval != "stringvalue" {
		t.Fatalf("Unexpected value for sval '%s', expected 'stringvalue'", sval)
	}

	if val, found = vals["boolkey"]; !found {
		t.Fatalf("Failed to lookup key 'boolkey'")
	}

	bval := val.(bool)
	if !bval {
		t.Fatalf("bval unexpectedly false")
	}

	if val, found = vals["intkey"]; !found {
		t.Fatalf("Failed to lookup key 'intkey'")
	}

	ival := val.(float64)
	if ival != 123 {
		t.Fatalf("Unexpected value for ival %f, expected 123", ival)
	}
}
