package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Request struct {
	// Method and URL are set from filename in main.go
	Method  string            // Not in JSON
	URL     string            // Not in JSON
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
}

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

var envVarRegex = regexp.MustCompile(`\{\{([^}]+)}}`)

func replaceEnvVars(s string) string {
	return envVarRegex.ReplaceAllStringFunc(s, func(match string) string {
		key := strings.Trim(match, "{}")
		value := os.Getenv(key)
		if value == "" {
			return match
		}
		return value
	})
}

func ExecuteRequest(req Request) (status string, body string, err error) {
	req.URL = replaceEnvVars(req.URL)
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		return "", "", fmt.Errorf("invalid URL after env substitution: %s (missing BASE_URL?)", req.URL)
	}
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
