package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"

	"github.com/pkg/errors"
)

// QuerySearchRepos (not used) returns search results for the repository endpoint, based on queries.
// The results are returned on a best effort basis, and unfortunately is
// not as reliable as going against the endpoint for the repository resource.
func (gh *Github) QuerySearchRepos(ctx context.Context, query Query) ([]Repos, error) {
	endPoint := url.URL{Path: "/search/repositories"}
	githubURL := gh.baseURL.ResolveReference(&endPoint)

	q := githubURL.Query()
	q.Set("q", "is:public stars:500..1000")
	q.Set("sort", "stars")
	q.Set("per_page", "100")
	q.Set("since", fmt.Sprint(query.Since))
	q.Set("order", "asc")
	githubURL.RawQuery = q.Encode()
	requestPath := githubURL.String()

	curateRepos := func(repos []Repos) []Repos {
		if len(repos) == 0 {
			return nil
		}

		if rangeTooHigh(repos[0].ID, query.MaxID) {
			return nil
		}
		lastID := repos[len(repos)-1].ID
		if lastID == query.MaxID {
			return repos
		}
		if lastID >= query.MaxID {
			// Exclude additional repos that don't meet the cutoff max id.
			return trim(query.MaxID, repos)
		}
		return repos
	}

	i := 1
	var allRepos []Repos
	hasNextPage := true
	for hasNextPage {
		fmt.Fprintln(os.Stderr, "\t"+requestPath)
		fmt.Println("Page number:", i)
		i++

		var data Data
		resp, err := gh.do(requestPath, &data)
		if err != nil {
			return nil, err
		}

		fmt.Println("data repo count:", len(data.Repos))
		var curatedRepos []Repos
		if curatedRepos = curateRepos(data.Repos); curatedRepos == nil {
			break
		}
		allRepos = append(allRepos, curatedRepos...)
		requestPath, hasNextPage = findNextPage(resp)
	}
	return allRepos, nil
}

// QueryRepos returns a list of public GH repositories. It starts the query based on the
// value of the `since` parameter, and it stops and trims the returned results based
// on the specified `maxID`.
//
// Note that this endpoint does not respect the (asc/desc) direction parameter, but does
// return the elements in ascending order.
func (gh *Github) QueryRepos(ctx context.Context, query Query) ([]Repos, error) {
	endPoint := url.URL{Path: "/repositories"}
	githubURL := gh.baseURL.ResolveReference(&endPoint)

	q := githubURL.Query()
	q.Set("since", fmt.Sprint(query.Since))
	githubURL.RawQuery = q.Encode()

	requestPath := githubURL.String()
	var allRepos []Repos

	curateRepos := func(repos []Repos) []Repos {
		if len(repos) == 0 {
			return nil
		}

		if rangeTooHigh(repos[0].ID, query.MaxID) {
			return nil
		}
		lastID := repos[len(repos)-1].ID
		fmt.Println("Repo ID for the last record in this page:", lastID)
		if lastID == query.MaxID {
			return repos
		}
		if lastID >= query.MaxID {
			// Exclude additional repos that don't meet the cutoff max id.
			return trim(query.MaxID, repos)
		}
		return repos
	}

	hasNextPage := true
	for hasNextPage {
		fmt.Fprintln(os.Stderr, "\t"+requestPath)

		var repos []Repos
		resp, err := gh.do(requestPath, &repos)
		if err != nil {
			return nil, err
		}
		fmt.Printf("\nReturned records for this page: %d\n", len(repos))

		var curatedRepos []Repos
		if curatedRepos = curateRepos(repos); curatedRepos != nil {
			allRepos = append(allRepos, curatedRepos...)
			lastID := repos[len(allRepos)-1].ID
			if lastID == query.MaxID {
				break
			}
		}

		requestPath, hasNextPage = findNextPage(resp)
	}

	return allRepos, nil
}

func (gh *Github) QueryStars(ctx context.Context, repos []Repos) (map[string]map[int]int, []error) {
	var errs []error
	bucketTierStarCount := make(map[string]int)
	bucketTierRepoCount := make(map[string]int)
	buckets := make(map[string]map[int]int)

	for i := range repos {
		path := "/repos" + "/" + repos[i].Owner.Login + "/" + repos[i].Name
		endPoint := url.URL{Path: path}
		githubURL := gh.baseURL.ResolveReference(&endPoint)

		var star struct {
			Count int `json:"stargazers_count"`
		}

		res, err := gh.do(githubURL.String(), &star)
		if res != nil {
			success := res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices
			if !success {
				continue
			}
		}

		if err != nil {
			err = errors.Wrapf(err, fmt.Sprintf("-- not possible to retrieve startgazers for login: %s/ name: %s", repos[i].Owner.Login, repos[i].Name))
			errs = append(errs, err)
		}
		bucketTierRepoCount[bucketTier(star.Count)]++
		bucketTierStarCount[bucketTier(star.Count)] += star.Count
		buckets[bucketTier(star.Count)] = map[int]int{bucketTierRepoCount[bucketTier(star.Count)]: bucketTierStarCount[bucketTier(star.Count)]}
	}

	return buckets, errs
}

func (gh *Github) QueryLicenses(ctx context.Context, repos []Repos) map[string]int {
	licenses := make(map[string]int)

	for i := range repos {
		path := "/repos" + "/" + repos[i].Owner.Login + "/" + repos[i].Name + "/license"
		endPoint := url.URL{Path: path}
		githubURL := gh.baseURL.ResolveReference(&endPoint)

		var data struct {
			Name    string  `json:"name"`
			License License `json:"license"`
		}

		var res *http.Response
		var err error
		res, err = gh.do(githubURL.String(), &data)
		if res != nil {
			success := res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices
			if !success {

				continue
			}
		}
		if err != nil {
			data.License.Name = "Unknown Error for License Record"
		}

		licenses[data.License.Name]++
	}

	return licenses
}

// trim excludes repos that have IDs above the cutoff max ID.
// Because the retrived repos have IDs are not consecutive, it might
// be the case that the trimmed last ID is maxID but also maxID - n.
//
// TODO: Add a test for this function and remove the printouts
// for visually validating the results.
func trim(maxID int, repos []Repos) []Repos {
	// Returns the first number it finds that is greater than or equal to maxID.
	i := sort.Search(len(repos), func(i int) bool {
		return repos[i].ID >= maxID
	})

	switch {
	case i < len(repos) && repos[i].ID == maxID:
		fmt.Println("Trimmming at case 1: max id was found in the set")
		// max id was found in the set
		// this means search stopped at maxid
		// cut starting at i+1
		//
		// ex: 3,10,50
		// max=10
		// ex: 3,10,30
		// max=30
		repos = repos[:i+1]
		fmt.Printf("Last ID of the set being returned: %d\n\n", repos[len(repos)-1].ID)
		return repos
	default:
		fmt.Println("Trimming at max id was found in the set")
		// max id was not found in the set
		// this means search stopped at maxid+1
		// cut starting at i
		//
		// ex: 3,5,50
		// max=10
		repos = repos[:i]
		fmt.Printf("Last ID of the set being returned: %d\n\n", repos[len(repos)-1].ID)
		return repos
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
