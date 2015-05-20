package ezhttp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

// EzClient is a container for the http requests.
type EzClient struct {
	*http.Client
	body io.Reader
}

// NewSSLClient creates a new SSL client but skips verification.
func NewSSLClient() *EzClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &EzClient{Client: &http.Client{Transport: tr}}
}

// NewClient creates a http connection.
func NewClient() *EzClient {
	return &EzClient{Client: &http.Client{}}
}

// JSONGet performs a get request and decodes the resulting JSON.
func (c *EzClient) JSONGet(url string, out interface{}) (int, error) {
	resp, err := c.Client.Get(url)
	if err != nil {
		return 0, err
	}
	return decodeJSONResponse(resp, out)
}

// FileGet downloads a file and writes it to path.
func (c *EzClient) FileGet(url, path string) (int, error) {
	resp, err := c.Client.Get(url)
	if err != nil {
		return 0, err
	}

	return writeToPath(resp, path)
}

// writeToPath writes the file in the Response to path.
func writeToPath(resp *http.Response, path string) (int, error) {
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		msg, _ := ioutil.ReadAll(resp.Body)
		return resp.StatusCode, errors.New(string(msg))
	}

	out, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	io.Copy(out, resp.Body)
	return resp.StatusCode, nil
}

// JSONStr takes a string formatted as JSON and makes it available for POST and PUT.
func (c *EzClient) JSONStr(j string) *EzClient {
	c.body = strings.NewReader(j)
	return c
}

// JSON takes an object, marshals it and make it avaiable for POST and PUT.
func (c *EzClient) JSON(j interface{}) *EzClient {
	b, err := json.Marshal(j)
	if err == nil {
		c.body = bytes.NewReader(b)
	}

	return c
}

// JSONPost performs a POST with JSON.
func (c *EzClient) JSONPost(url string, out interface{}) (int, error) {
	resp, err := c.Client.Post(url, "application/json", c.body)
	if err != nil {
		return 0, err
	}
	return decodeJSONResponse(resp, out)
}

// decodeJSONResponse decodes a response containing JSON.
func decodeJSONResponse(resp *http.Response, out interface{}) (int, error) {
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode > 299 {
		return resp.StatusCode, errors.New(string(body))
	}

	err := json.Unmarshal(body, out)
	if err != nil {
		return resp.StatusCode, err
	}
	return resp.StatusCode, nil
}

// PostFile posts a whole file
func (c *EzClient) PostFile(url, filepath, formName string, params map[string]string) (int, error) {
	// Read file into form
	file, err := os.Open(filepath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	fileContents, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, err
	}
	return c.PostFileBytes(url, filepath, formName, fileContents, params)
}

// PostFileBytes posts a range of bytes
func (c *EzClient) PostFileBytes(url, filepath, formName string, contents []byte, params map[string]string) (int, error) {
	// Setup body
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Create file form
	part, err := writer.CreateFormFile(formName, filepath)
	if err != nil {
		return 0, err
	}

	part.Write(contents)

	// Add additional params
	if params != nil {
		for key, value := range params {
			fw, _ := writer.CreateFormField(key)
			fw.Write([]byte(value))
		}
	}
	writer.Close()

	req, err := http.NewRequest("POST", url, &body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	req.Close = true

	// Post the file
	resp, err := c.Client.Do(req)
	defer resp.Body.Close()
	b, _ := ioutil.ReadAll(resp.Body)

	// Return status and any error code
	if resp.StatusCode > 299 {
		return resp.StatusCode, errors.New(string(b))
	}

	return resp.StatusCode, nil
}
