package loader

import (
	"bytes"
	"fmt"
	"testing"
)

var _ = fmt.Println

var tomlTestData = []byte(`
[test]
stringkey = "stringvalue"
boolkey = true
intkey = 123
`)

func TestTOMLLoader(t *testing.T) {
	var (
		vals  map[string]interface{}
		tvals map[string]interface{}
		val   interface{}
		found bool
	)

	l := TOML(bytes.NewReader(tomlTestData))
	err := l.Load(&vals)
	if err != nil {
		t.Fatalf("failed loading TOML: %s", err)
	}

	ttvals, found := vals["test"]
	if !found {
		t.Fatalf("Failed to lookup key 'test'")
	}

	tvals = ttvals.(map[string]interface{})
	if val, found = tvals["stringkey"]; !found {
		t.Fatalf("Failed to lookup key 'stringkey'")
	}

	sval := val.(string)
	if sval != "stringvalue" {
		t.Fatalf("Unexpected value for sval '%s', expected 'stringvalue'", sval)
	}

	if val, found = tvals["boolkey"]; !found {
		t.Fatalf("Failed to lookup key 'boolkey'")
	}

	bval := val.(bool)
	if !bval {
		t.Fatalf("Unexpected value for bval, expected true")
	}

	if val, found = tvals["intkey"]; !found {
		t.Fatalf("Failed to lookup key 'intkey'")
	}

	ival := val.(int64)
	if ival != 123 {
		t.Fatalf("Unexpected value for ival '%d', expected 123", ival)
	}
}
