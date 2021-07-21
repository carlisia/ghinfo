// Package analytics creates and formats reports.
// It is a tier that accesses the data and transforms it into
// reports.

package analytics

import (
	"context"
	"errors"
	"fmt"

	"github.com/carlisia/ghinfo/github"
)

type StatsReport interface {
	PrintStats()
	Run(context.Context, *github.Github) error
	Count() int
	Name() string
}

// Sorter to be implemented for sorting the different types of reports.
// TODO
type Sorter interface {
	Sort()
}

type ParamOptions struct {
	SortColumn string
	MaxID      int
	Since      int
}

type report struct {
	name             string
	query            github.Query
	repoCount        int
	aggregatedErrors []error
}

type sortColumn struct {
	asc    bool
	column string
}

const (
	StarGazers  = "1"
	LicenseType = "2"

	starGazersReport   = "StarGazers Report"
	licenseTypesReport = "License Types Report"
	bucketCol          = "bucket"
	licenseCol         = "license type"
	repoCol            = "repos"
	asc                = "asc"
	desc               = "desc"
)

func NewReport(reportType string, opts ParamOptions) (StatsReport, error) {
	if err := validateIDRange(opts.Since, opts.MaxID); err != nil {
		return nil, err
	}

	switch reportType {
	case StarGazers:
		return &BucketReport{
			ParamOptions: opts,
			report: report{
				name:  starGazersReport,
				query: github.Query{Since: opts.Since, MaxID: opts.MaxID},
			},
		}, nil
	case LicenseType:
		return &LicenseTypeReport{
			ParamOptions: opts,
			report: report{
				name:  licenseTypesReport,
				query: github.Query{Since: opts.Since, MaxID: opts.MaxID},
			},
		}, nil
	default:
		return nil, errors.New("no report type was selected")
	}
}

func queryRepos(ctx context.Context, gh *github.Github, query github.Query) ([]github.Repos, error) {
	repos, err := gh.QueryRepos(ctx, query)
	if err != nil {
		return nil, err
	}

	return repos, nil
}

func validateIDRange(since, max int) error {
	const maxNumIDs = 500

	if max < since {
		msg := fmt.Sprintf("the `maxID` value (%d) cannot be smaller than the `since` value (%d)", max, since)
		return errors.New(msg)
	}

	numIDs := max - since
	if numIDs > maxNumIDs {
		msg := fmt.Sprintf("the number of IDs (%d)  has exceeded the limit (%d)", numIDs, maxNumIDs)
		return errors.New(msg)
	}

	return nil
}
