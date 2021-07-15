package github

type Owner struct {
	Login string `json:"login,omitempty"`
}

type Repo struct {
	Owner Owner  `json:"owner,omitempty"`
	Name  string `json:"name,omitempty"`
	ID    int    `json:"id,omitempty"`
}

type Star struct {
	Repo  Repo
	Count int
}

// Paging is used to handle parameters for querying and filtering the public repo API.
//
// Note: fo the public repositories endpoint, the `page` and `per_page` parameters are not
// being respected, therefore not used here. It is a known issue:
// https://docs.github.com/en/rest/overview/resources-in-the-rest-api#pagination.
// To test:
//
// curl \
//   -H "Accept: application/vnd.github.v3+json" \
//   https://api.github.com/repositories?since=65624569&page=1&per_page=1
type Paging struct {
	// Since is a repository ID. The API will only return repositories
	// with an ID greater than this ID. If not specified, the API will
	// return its first existing record.
	Since int `json:"since,omitempty"`

	// MaxID is the cut off repo ID we want to to return.
	MaxID int `json:"max_id,omitempty"`
}
