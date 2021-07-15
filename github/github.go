// Package github provides api access to getting repos and stars.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/pkg/errors"
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

func (gh *Github) QuerySearch(ctx context.Context, paging Paging) ([]Repo, error) {
	// https://github.com/search?utf8=%E2%9C%93&q=cats+stars%3A10..50&type=Repositories

	endPoint := url.URL{Path: "/search/repositories"}
	githubURL := gh.baseURL.ResolveReference(&endPoint)

	q := githubURL.Query()
	q.Set("since", fmt.Sprint(paging.Since))
	githubURL.RawQuery = q.Encode()

	requestPath := githubURL.String()
	var allRepos []Repo

	hasNextPage := true
	for hasNextPage {
		fmt.Fprintln(os.Stderr, "\t"+requestPath)

		var repos []Repo
		resp, err := gh.do(requestPath, &repos)
		if err != nil {
			return nil, err
		}

		// var curatedRepos []Repo
		// if curatedRepos = curateRepos(repos); curatedRepos == nil {
		// 	break
		// }

		// allRepos = append(allRepos, curatedRepos...)
		requestPath, hasNextPage = findNextPage(resp)
	}
	fmt.Println("Total repos on server:", len(allRepos))
	return allRepos, nil
}

// QueryRepos returns a list of public GH repositories. It starts the query based on the
// value of the `since` parameter, and it stops and trims the returned results based
// on the specified `maxID`.
//
// Note that this endpoint does not respect the (asc/desc) direction parameter, but does
// return the elements in ascending order.
func (gh *Github) QueryRepos(ctx context.Context, paging Paging) ([]Repo, error) {
	endPoint := url.URL{Path: "/repositories"}
	githubURL := gh.baseURL.ResolveReference(&endPoint)

	q := githubURL.Query()
	q.Set("since", fmt.Sprint(paging.Since))
	githubURL.RawQuery = q.Encode()

	requestPath := githubURL.String()
	var allRepos []Repo

	curateRepos := func(repos []Repo) []Repo {
		if rangeTooHigh(repos[0].ID, paging.MaxID) {
			return nil
		}
		lastID := repos[len(repos)-1].ID
		if lastID == paging.MaxID {
			return repos
		}
		if lastID >= paging.MaxID {
			// Exclude additional repos that don't meet the cutoff max id.
			return trim(paging.MaxID, repos)
		}
		return repos
	}

	hasNextPage := true
	for hasNextPage {
		fmt.Fprintln(os.Stderr, "\t"+requestPath)

		var repos []Repo
		resp, err := gh.do(requestPath, &repos)
		if err != nil {
			return nil, err
		}

		var curatedRepos []Repo
		if curatedRepos = curateRepos(repos); curatedRepos == nil {
			break
		}

		allRepos = append(allRepos, curatedRepos...)
		requestPath, hasNextPage = findNextPage(resp)
	}
	fmt.Println("Total repos on server:", len(allRepos))
	return allRepos, nil
}

func (gh *Github) QueryStars(ctx context.Context, repos []Repo) ([]Star, []error) {
	var stars []Star

	var errs []error
	for _, repo := range repos {
		path := "/repos" + "/" + repo.Owner.Login + "/" + repo.Name
		endPoint := url.URL{Path: path}
		githubURL := gh.baseURL.ResolveReference(&endPoint)

		var star struct {
			Count int `json:"stargazers_count"`
		}
		_, err := gh.do(githubURL.String(), &star)
		if err != nil {
			err = errors.Wrapf(err, fmt.Sprintf("-- not possible to retrieve startgazers for login: %s/ name: %s", repo.Owner.Login, repo.Name))
			errs = append(errs, err)
			continue
		}

		stars = append(stars, Star{
			Repo:  Repo{repo.Owner, repo.Name, repo.ID},
			Count: star.Count},
		)
	}

	return stars, errs
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something went wrong with the request: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(data); err != nil {
		return nil, err
	}

	return resp, nil
}

// trim excludes repos that have IDs above the cutoff max ID.
// Because the retrived repos have IDs are not consecutive, it might
// be the case that the trimmed last ID is maxID but also maxID - n.
func trim(maxID int, repos []Repo) []Repo {
	fmt.Println("inside trim func")
	fmt.Println("TRIMMING incoming lenght...", len(repos))

	// Returns the first number it finds that is greater than or equal to maxID.
	i := sort.Search(len(repos), func(i int) bool {
		return repos[i].ID >= maxID
	})

	switch {
	case i < len(repos) && repos[i].ID == maxID:
		// max id was found in the set
		// this means search stopped at maxid
		// cut starting at i+1
		//
		// ex: 3,10,50
		// max=10
		// ex: 3,10,30
		// max=30
		return repos[:i+1]
	default:
		// max id was not found in the set
		// this means search stopped at maxid+1
		// cut starting at i
		//
		// ex: 3,5,50
		// max=10
		return repos[:i]
	}
}

// rangeTooHigh checks if the first int is higher
// than the second. It is useful to check for cases such as:
// a) a request was received that had the `since` value higher than
// the given max ID value;
// b) a request was received with a given `since` value, but the first
// existing record returned by the GH API has a value that crosses past
// the given max ID value.
func rangeTooHigh(firstID, maxID int) bool {
	return firstID > maxID
}
