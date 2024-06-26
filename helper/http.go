package helper

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func HttpRequest(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	client := &http.Client{}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}

	return resp, nil
}

func IsURLAccessible(url string) bool {
	_, err := http.Head(url)
	if err != nil {
		return false
	}
	return true
}

func IsURLAccessibleRecursive(url string, retries int, interval time.Duration) bool {
	if retries == 0 {
		return false
	}

	_, err := http.Head(url)
	if err != nil {
		time.Sleep(interval)
		return IsURLAccessibleRecursive(url, retries-1, interval)
	}

	return true
}
