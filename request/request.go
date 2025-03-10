package request

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type Request struct {
	Method  string            // Not in JSON
	URL     string            // Not in JSON
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
	EnvName string            // Not in JSON, set by main.go
	Env     map[string]struct {
		Header string `json:"header,omitempty"`
		Body   string `json:"body,omitempty"`
	} `json:"env,omitempty"`
}

type Config struct {
	DefaultEnv string                       `yaml:"default_env"`
	Envs       map[string]map[string]string `yaml:"envs"`
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

var envVarRegex = regexp.MustCompile(`\{\{([^}]+)\}\}`)

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

	var bodyStr string
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			return "", "", fmt.Errorf("error marshaling body: %v", err)
		}
		bodyStr = replaceEnvVars(string(bodyBytes))
	}

	client := &http.Client{}
	httpReq, err := http.NewRequest(req.Method, req.URL, strings.NewReader(bodyStr))
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

	// Save env vars to .localpost under envs section
	configFilePath := filepath.Join(os.Getenv("HOME"), ".localpost")
	config := Config{
		DefaultEnv: "dev",
		Envs:       make(map[string]map[string]string),
	}
	if data, err := os.ReadFile(configFilePath); err == nil {
		if err := yaml.Unmarshal(data, &config); err != nil {
			return "", "", fmt.Errorf("error parsing .localpost: %v", err)
		}
	}

	// Safeguard against empty EnvName
	if req.EnvName == "" {
		return "", "", fmt.Errorf("environment name is empty, cannot save env vars")
	}

	// Ensure the env section exists
	if config.Envs[req.EnvName] == nil {
		config.Envs[req.EnvName] = make(map[string]string)
	}

	var respData map[string]interface{}
	if len(respBody) > 0 {
		json.Unmarshal(respBody, &respData)
	}

	for envKey, rule := range req.Env {
		var value string
		if rule.Header != "" {
			value = resp.Header.Get(rule.Header)
		} else if rule.Body != "" && respData != nil {
			if v, ok := respData[rule.Body].(string); ok {
				value = v
			}
		}
		if value != "" {
			config.Envs[req.EnvName][envKey] = value
		}
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return "", "", fmt.Errorf("error marshaling config: %v", err)
	}
	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		return "", "", fmt.Errorf("error writing .localpost: %v", err)
	}

	return resp.Status, string(respBody), nil
}
