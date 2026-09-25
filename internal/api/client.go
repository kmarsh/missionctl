// Package api talks to the Mission Control JSON API.
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const BaseURL = "https://missioncontrol.dev/api/v1/"

// Error is a non-2xx response. Body is the API's {"error": "..."} or
// {"errors": {"field": ["message"]}}, or whatever else came back.
type Error struct {
	Method string
	URL    string
	Status int
	Body   []byte
}

func (e *Error) Error() string {
	return e.Message()
}

// Message is the API's error message, with validation errors as "field message; ...".
func (e *Error) Message() string {
	var body struct {
		Error  string              `json:"error"`
		Errors map[string][]string `json:"errors"`
	}
	if json.Unmarshal(e.Body, &body) == nil {
		if body.Error != "" {
			return body.Error
		}
		if len(body.Errors) > 0 {
			fields := make([]string, 0, len(body.Errors))
			for field := range body.Errors {
				fields = append(fields, field)
			}
			sort.Strings(fields)
			var messages []string
			for _, field := range fields {
				for _, message := range body.Errors[field] {
					messages = append(messages, field+" "+message)
				}
			}
			return strings.Join(messages, "; ")
		}
	}
	if text := strings.TrimSpace(string(e.Body)); text != "" {
		return text
	}
	return http.StatusText(e.Status)
}

type Client struct {
	Key       string
	UserAgent string
	HTTP      *http.Client
}

func New(key, version string) *Client {
	return &Client{Key: key, UserAgent: "missionctl/" + version, HTTP: &http.Client{Timeout: 30 * time.Second}}
}

// Do sends method to path under BaseURL and returns the parsed JSON response,
// or nil for an empty one. Numbers come back as json.Number.
func (c *Client) Do(method, path string, query url.Values, body any) (any, error) {
	target := BaseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(encoded)
	}
	request, err := http.NewRequest(method, target, reader)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+c.Key)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", c.UserAgent)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.HTTP.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, &Error{Method: method, URL: target, Status: response.StatusCode, Body: data}
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var parsed any
	if err := decoder.Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%s %s returned unreadable JSON: %w", method, target, err)
	}
	return parsed, nil
}
