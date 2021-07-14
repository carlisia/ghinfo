package github

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	// baseURL is the default public GitHub API url.
	baseURL   = "api.github.com"
	userAgent = "https://github.com/carlisia/ghinfo"
)

type Client struct {
	client *http.Client

	BaseURL, UserAgent string
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	c := &Client{
		client:    httpClient,
		BaseURL:   baseURL,
		UserAgent: userAgent,
	}

	return c
}

type Repo struct {
	ID int `json:"id,omitempty"`
}

// PublicRepos returns all public repositories, in the order they were created, starting at
// the specified ID indicated by the `since` query value.
func (c Client) PublicRepos(method, path string, query map[string]string) ([]Repo, error) {
	rel := &url.URL{
		Scheme: "https",
		Host:   string(c.BaseURL),
		Path:   path,
	}
	url := rel.ResolveReference(rel)
	q := url.Query()
	q.Set("since", query["since"])
	url.RawQuery = q.Encode()

	var buf io.ReadWriter
	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	fmt.Println("Rate limiting requests remaining: ", resp.Header["X-Ratelimit-Remaining"][0])

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !success {
		return nil, fmt.Errorf("something went wrong with the request: %s", resp.Status)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var repos []Repo
	err = json.Unmarshal(b, &repos)
	if err != nil {
		return nil, err
	}

	return repos, nil
}
