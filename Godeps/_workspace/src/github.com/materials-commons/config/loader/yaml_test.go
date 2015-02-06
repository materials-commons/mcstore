package loader

import (
	"bytes"
	"fmt"
	"testing"
)

var _ = fmt.Println

var yamlTestData = []byte(`
stringkey: stringvalue
boolkey: true
intkey: 123
`)

func TestYAMLLoader(t *testing.T) {
	var (
		vals  map[string]interface{}
		val   interface{}
		found bool
	)

	l := YAML(bytes.NewReader(yamlTestData))
	if err := l.Load(&vals); err != nil {
		t.Fatalf("Failed loading YAML: %s", err)
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

	ival := val.(int)
	if ival != 123 {
		t.Fatalf("Unexpected value for ival %d, expected 123", ival)
	}
}
