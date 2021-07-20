package github

// Query is used to handle parameters for querying and filtering the GH API.
//
// Note: fo the public repositories endpoint, the `page` and `per_page` parameters are not
// being respected, therefore not used here. It is a known issue:
// https://docs.github.com/en/rest/overview/resources-in-the-rest-api#pagination.
// To test:
//
// curl \
//   -H "Accept: application/vnd.github.v3+json" \
//   https://api.github.com/repositories?since=65624569&page=1&per_page=1
type Query struct {
	// Since is a repository ID. The API will only return repositories
	// with an ID greater than this ID. If not specified, the API will
	// return its first existing record.
	Since int `url:"since,omitempty"`

	// MaxID is the cut off repo ID we want to to return.
	MaxID int `url:"max_id,omitempty"`

	Sort string `url:"sort,omitempty"`

	Q string `url:"q,omitempty"`
}
type Data struct {
	TotalCount        int     `json:"total_count"`
	IncompleteResults bool    `json:"incomplete_results"`
	Repos             []Repos `json:"items"`
}
type Owner struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
}

type Repos struct {
	ID              int     `json:"id"`
	NodeID          string  `json:"node_id"`
	Name            string  `json:"name"`
	FullName        string  `json:"full_name"`
	Private         bool    `json:"private"`
	Owner           Owner   `json:"owner"`
	StargazersCount int     `json:"stargazers_count"`
	License         License `json:"license"`
}

// https://docs.github.com/en/rest/reference/licenses#get-the-license-for-a-repository
type License struct {
	Key    string `json:"key"`
	Name   string `json:"name"`
	SpdxID string `json:"spdx_id"`
	URL    string `json:"url"`
	NodeID string `json:"node_id"`
}

type RepoInfo struct {
	ID              int `json:"id"`
	StarGazersCount int `json:"stargazers_count"`
}

var bucketTiers = []string{"0..10", "10..100", "100..1000", "1000..5000", "5000..10000", ">=10000"}

func bucketTier(count int) string {
	switch {
	case count <= 10:
		return bucketTiers[0]
	case count <= 100:
		return bucketTiers[1]
	case count <= 1000:
		return bucketTiers[2]
	case count <= 5000:
		return bucketTiers[3]
	case count <= 10000:
		return bucketTiers[4]
	default:
		return bucketTiers[5]
	}
}
