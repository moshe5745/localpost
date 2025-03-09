package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// Request represents an HTTP request stored in JSON
type Request struct {
	Name    string            `json:"name"`
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

// ParseRequest reads and parses a JSON request file
func ParseRequest(filePath string) (Request, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return Request{}, fmt.Errorf("error reading request file: %v", err)
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return Request{}, fmt.Errorf("error parsing request file: %v", err)
	}

	return req, nil
}

// replaceEnvVars substitutes environment variables in a string
func replaceEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		return os.Getenv(key)
	})
}

// ExecuteRequest sends the HTTP request and returns the response
func ExecuteRequest(req Request) (status string, body string, err error) {
	// Replace env vars
	req.URL = replaceEnvVars(req.URL)
	for key, value := range req.Headers {
		req.Headers[key] = replaceEnvVars(value)
	}
	if req.Body != "" {
		req.Body = replaceEnvVars(req.Body)
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest(req.Method, req.URL, strings.NewReader(req.Body))
	if err != nil {
		return "", "", fmt.Errorf("error creating request: %v", err)
	}

	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("error executing request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("error reading response: %v", err)
	}

	return resp.Status, string(respBody), nil
}
