package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/carlisia/ghinfo/github"
	"golang.org/x/oauth2"
)

const baseURL = "https://api.github.com"
const userAgent = "https://github.com/carlisia/ghinfo"

func main() {
	const gitHubToken = "GH_TOKEN"
	userToken := os.Getenv(gitHubToken)
	if userToken == "" {
		log.Fatal("It is likely that the personal access token was not set")
	}

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: userToken,
		},
	)

	ctx := context.Background()
	client := oauth2.NewClient(ctx, tokenSource)
	gh, err := github.New(client, baseURL, userAgent)
	if err != nil {
		log.Fatalln("Error trying to initalize github:", err)
	}

	paging := github.Paging{
		Since: 9950000,
		MaxID: 9950020,
	}

	if err = validate(paging); err != nil {
		log.Fatalln("Invalid parameters", err)
	}

	repos, err := gh.QueryRepos(ctx, paging)
	if err != nil {
		log.Fatalln("Error trying to retrieve the repository list:", err)
	}

	log.Println("Total repos:", len(repos))

	stars, errs := gh.QueryStars(ctx, repos)
	if len(errs) > 0 {
		log.Printf("%d errors trying to retrieve the startgazers count, skipping:\n", len(errs))
		for _, err := range errs {
			log.Println(err.Error())
		}
	}

	if len(errs) > 0 && len(stars) == 0 {
		log.Fatalln("No start stats was retrived, possibly due to errors")
	}

	log.Println("Total repos:", len(stars))

	buckets := github.AggregateStarStats(stars)
	log.Printf("%+v\n", buckets)
	// buckets.Bucket1
}

func validate(paging github.Paging) error {
	const maxNumIDs = 500

	if paging.MaxID < paging.Since {
		msg := fmt.Sprintf("the `maxID` value (%d) cannot be smaller than the `since` value (%d)", paging.MaxID, paging.Since)
		return errors.New(msg)
	}

	numIDs := paging.MaxID - paging.Since
	if numIDs > maxNumIDs {
		msg := fmt.Sprintf("the number of IDs (%d)  has exceeded the limit (%d)", numIDs, maxNumIDs)
		return errors.New(msg)
	}

	return nil
}
