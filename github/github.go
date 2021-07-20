// Package github provides api access to getting repos and stars.
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Github struct {
	client    HTTPClient
	baseURL   *url.URL
	userAgent string
}

func New(httpClient HTTPClient, baseURL string, userAgent string) (*Github, error) {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	githubURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	gh := Github{
		client:    httpClient,
		baseURL:   githubURL,
		userAgent: userAgent,
	}
	return &gh, nil
}

// do only processes `GET` requests.
func (gh *Github) do(url string, data interface{}) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", gh.userAgent)

	resp, err := gh.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	success := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
	if !success {
		return nil, fmt.Errorf("something went wrong with the request: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}

	return resp, nil
}
