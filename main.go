package main

import (
	"fmt"
	"log"

	"github.com/carlisia/ghinfo/config"
	"github.com/carlisia/ghinfo/github"
	"golang.org/x/oauth2"
)

const (
	// baseURL is the default public GitHub API url.
	baseURL   = "https://api.github.com"
	userAgent = "https://github.com/carlisia/ghinfo"
)

func main() {
	method := "GET"
	path := fmt.Sprintf("/repositories")

	// Retrieve the user access token
	userToken := config.EnvironmentAuthentication()
	if userToken == "" {
		fmt.Println("error: it is likely the personal access token was not set")
		return
	}
	token := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: userToken},
	)

	// Instantiate the client and send the request.
	tc := oauth2.NewClient(oauth2.NoContext, token)
	client := github.NewClient(tc)

	query := map[string]string{
		"since": "1000",
	}
	repos, err := client.PublicRepos(method, path, query)
	if err != nil {
		log.Fatal(err)
	}

	if len(repos) < 1 {
		log.Println("No public repos found")
	}

	fmt.Println("Total repos:", len(repos))
}
