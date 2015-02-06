package loader

import (
	"bytes"
	"testing"
)

var propTestData = []byte(`
stringkey = stringvalue
`)

func TestPropertiesLoader(t *testing.T) {
	var (
		vals map[string]string
	)

	l := Properties(bytes.NewReader(propTestData))
	err := l.Load(&vals)
	if err != nil {
		t.Fatalf("Failed loading properties: %s", err)
	}

	val, found := vals["stringkey"]
	if !found {
		t.Fatalf("Failed looking up stringkey")
	}

	if val != "stringvalue" {
		t.Fatalf("Unexpected value for val: %s, expected 'stringvalue'")
	}
}
