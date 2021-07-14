package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

const (
	// baseURL is the default public GitHub API url.
	baseURL   = "https://api.github.com"
	userAgent = "https://github.com/carlisia/ghinfo"
)

type Repo struct {
	ID int `json:"id,omitempty"`
}

func main() {
	baseURL, err := url.Parse(baseURL)
	if err != nil {
		fmt.Println("error parsing the base URL", err)
		return
	}

	path := "/repositories"
	rel := &url.URL{Path: path}
	url := baseURL.ResolveReference(rel)

	method := "GET"
	var buf io.ReadWriter
	req, err := http.NewRequest(method, url.String(), buf)
	if err != nil {
		fmt.Println("error creating the request", err)
		return
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", userAgent)

	// Instantiate the client.
	token := oauth2.StaticTokenSource(
		// Temporary token with very limited read permissions.
		&oauth2.Token{AccessToken: "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"},
	)

	tc := oauth2.NewClient(oauth2.NoContext, token)
	resp, err := tc.Do(req)
	if err != nil {
		fmt.Println("error making the request", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Rate limiting requests remaining: ", resp.Header["X-Ratelimit-Remaining"][0])

	if resp.StatusCode != http.StatusOK {
		fmt.Println("error making the request", resp.Status)
	}

	defer resp.Body.Close()

	var repos []Repo
	err = json.NewDecoder(resp.Body).Decode(&repos)
	if err != nil {
		fmt.Println("error decoding the request", err)
		return
	}

	if len(repos) < 1 {
		log.Println("No public repos found")
	}

	fmt.Printf("%d public repos found\n", len(repos))
}
