package ezhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

type TestData struct {
	Field1 string `json:"field1"`
	Field2 string `json:"field2"`
}

func handlerFunc(w http.ResponseWriter, r *http.Request) {
	t := TestData{
		Field1: "hello1",
		Field2: "hello2",
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(t)
}

func TestJSONGet(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handlerFunc))
	defer ts.Close()

	c := NewClient()
	var data TestData
	status, err := c.JSONGet(ts.URL, &data)
	if err != nil {
		t.Fatalf("err should be nil: %s", err.Error())
	}

	if status != 200 {
		t.Fatalf("JSONGet status should be 200, got %d\n", status)
	}

	if data.Field1 != "hello1" || data.Field2 != "hello2" {
		t.Fatalf("Incorrect decode, expected fields to be 'hello1' and 'hello2', got %s/%s",
			data.Field1, data.Field2)
	}
}

func TestJSONPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			decoder := json.NewDecoder(r.Body)
			var data TestData
			err := decoder.Decode(&data)
			if err != nil {
				t.Fatalf("Unable to decode post data\n")
			}

			if data.Field1 != "hello1" || data.Field2 != "hello2" {
				t.Fatalf("Incorrect decode, expected fields to be 'hello1' and 'hello2', got %s/%s\n",
					data.Field1, data.Field2)
			}

			encoder := json.NewEncoder(w)
			encoder.Encode(data)
		}))
	defer ts.Close()
	td := TestData{
		Field1: "hello1",
		Field2: "hello2",
	}

	var td2 TestData

	c := NewClient()
	status, err := c.JSON(&td).JSONPost(ts.URL, &td2)
	if err != nil {
		t.Fatalf("JSONPost returned unexpected error %s\n", err.Error())
	}

	if status != 200 {
		t.Fatalf("Expected status == 200, got %d\n", status)
	}

	if td2.Field1 != "hello1" || td2.Field2 != "hello2" {
		t.Fatalf("Incorrect decode, expected fields to be 'hello1' and 'hello2', got %s/%s\n",
			td2.Field1, td2.Field2)
	}

	status, err = c.JSONStr(`{"field1": "hello1", "field2": "hello2"}`).JSONPost(ts.URL, &td2)
	if err != nil {
		t.Fatalf("JSONPost returned unexpected error %s\n", err.Error())
	}

	if status != 200 {
		t.Fatalf("Expected status == 200, got %d\n", status)
	}

	if td2.Field1 != "hello1" || td2.Field2 != "hello2" {
		t.Fatalf("Incorrect decode, expected fields to be 'hello1' and 'hello2', got %s/%s\n",
			td2.Field1, td2.Field2)
	}
}

func TestFileGet(t *testing.T) {
	ts := httptest.NewServer(http.FileServer(http.Dir(".")))
	defer ts.Close()

	c := NewClient()
	path := filepath.Join(os.TempDir(), "http.go")
	c.FileGet(ts.URL+"/http.go", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Unable to get file %s\n", path)
	}
	os.Remove(path)
}

func TestPostFileBytes(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(uploadFileHandler))
	defer ts.Close()
	c := NewClient()
	s := "hello world"
	status, err := c.PostFileBytes(ts.URL, "/tmp/file.txt", "chunkData", []byte(s), nil)
	if err != nil {
		t.Fatalf("PostFileBytes errored with %s", err)
	}

	if status != 200 {
		t.Fatalf("PostFileBytes failed %d", status)
	}
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(1024 * 1024)
	file, _, err := r.FormFile("chunkData")
	if err != nil {
		fmt.Println("Err FormFile =", err)
		w.WriteHeader(500)
		return
	}
	defer file.Close()
	var b bytes.Buffer
	io.Copy(&b, file)
	if b.String() != "hello world" {
		w.WriteHeader(500)
	} else {
		w.WriteHeader(200)
	}
}
