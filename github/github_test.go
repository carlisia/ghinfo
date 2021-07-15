package github_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/carlisia/ghinfo/github"
	"github.com/stretchr/testify/require"
)

func newUnitTest(t *testing.T) (*github.Github, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	teardown := func() {
		server.Close()
	}

	gh, err := github.New(&mockClient{}, server.URL, "test-user-agent")
	if err != nil {
		t.Fatal(err)
	}

	return gh, teardown
}

type payload struct {
	body   []byte
	status int
}

// TestQueryRepos asserts that the list of GH repositories being
// returned by this call is consistent with the query criteria.
func TestQueryRepos(t *testing.T) {
	gh, teardown := newUnitTest(t)
	t.Cleanup(teardown)

	// A note about this test data set:
	// Because the GH endpoint for this call already only returns elements whose IDs are higher
	// (not equal) than the `since` value, there really is nothing to be done with it.
	// Nevertheless, the `paging.since` value is included for documentation purposes in a way that is consistent
	// with how it is used by the API. Specifically, the payload should only contain ID values that are higher then
	// `since` value.
	testCases := []struct {
		name          string
		payload       payload
		paging        github.Paging
		reposExpected []github.Repo
		expectedError string
	}{
		{
			name:          "1 repo, id is in betweek since and max id",
			payload:       payload{body: []byte(`[{"id": 5}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 10},
			reposExpected: []github.Repo{{ID: 5}},
		},
		{
			name:          "1 repo, id is higher than since and max id",
			payload:       payload{body: []byte(`[{"id": 10}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 5},
			reposExpected: []github.Repo{},
		},
		{
			name:          "2 repos, both with ids after since and below max id",
			payload:       payload{body: []byte(`[{"id": 4}, {"id": 7}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 10},
			reposExpected: []github.Repo{{ID: 4}, {ID: 7}},
		},

		{
			name:          "3 repos, last one has max id",
			payload:       payload{body: []byte(`[{"id": 2}, {"id": 3}, {"id": 5}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 5},
			reposExpected: []github.Repo{{ID: 2}, {ID: 3}, {ID: 5}},
		},
		{
			name:          "1 repos, id equal to max id",
			payload:       payload{body: []byte(`[{"id": 6}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 6},
			reposExpected: []github.Repo{{ID: 6}},
		},
		{
			name:          "1 repos, id one less than max id",
			payload:       payload{body: []byte(`[{"id": 5}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 6},
			reposExpected: []github.Repo{{ID: 5}},
		},
		{
			name:          "1 repos, id one higher than max id",
			payload:       payload{body: []byte(`[{"id": 7}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 6},
			reposExpected: []github.Repo{},
		},
		{
			name:          "1 repos, id one higher than max id",
			payload:       payload{body: []byte(`[{"id": 7}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 6},
			reposExpected: []github.Repo{},
		},
		{
			name:          "4 repos, last one with id higher than max id and none matches the max id",
			payload:       payload{body: []byte(`[{"id": 2}, {"id": 3}, {"id": 5}, {"id": 50}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 10},
			reposExpected: []github.Repo{{ID: 2}, {ID: 3}, {ID: 5}},
		},
		{
			name:          "4 repos, one with id higher than max id, and next to the last same as max id",
			payload:       payload{body: []byte(`[{"id": 2}, {"id": 3}, {"id": 10}, {"id": 50}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 10},
			reposExpected: []github.Repo{{ID: 2}, {ID: 3}, {ID: 10}},
		},
		{
			name:          "3 repos, last one with id same as max",
			payload:       payload{body: []byte(`[{"id": 3}, {"id": 10}, {"id": 30}]`), status: http.StatusOK},
			paging:        github.Paging{Since: 1, MaxID: 30},
			reposExpected: []github.Repo{{ID: 3}, {ID: 10}, {ID: 30}},
		},
	}

	for _, tc := range testCases {
		f := func(t *testing.T) {
			doFunc = mockResponse(tc.payload)
			repos, err := gh.QueryRepos(context.Background(), tc.paging)
			require.NoError(t, err, "failed to make a request")

			if tc.payload.status == http.StatusOK {
				require.Len(t, repos, len(tc.reposExpected))
				for k := range tc.reposExpected {
					require.Equal(t, tc.reposExpected[k], repos[k])
				}
			}
		}

		t.Run(tc.name, f)
	}
}

// mockResponse builds and returns a fake http response.
func mockResponse(payload payload) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		// Create a new reader with the JSON input.
		r := ioutil.NopCloser(bytes.NewReader([]byte(payload.body)))
		return &http.Response{
			Header: http.Header{
				"Accept":     {"application/vnd.github.v3+json"},
				"User-Agent": {"https://github.com/carlisia/ghinfo"},
			},
			StatusCode: payload.status,
			Body:       r,
		}, nil
	}
}

// doFunc is he mock client's response for the `Do` function.
var doFunc func(req *http.Request) (*http.Response, error)

type mockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do is the `Do` function for the mock client.
func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	return doFunc(req)
}
