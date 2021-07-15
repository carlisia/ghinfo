package github

import (
	"net/http"
	"regexp"
)

var linkRE = regexp.MustCompile(`<([^>]+)>;\s*rel="([^"]+)"`)

func findNextPage(resp *http.Response) (string, bool) {
	for _, m := range linkRE.FindAllStringSubmatch(resp.Header.Get("Link"), -1) {
		if len(m) >= 2 && m[2] == "next" {
			return m[1], true
		}
	}
	return "", false
}
