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
	Run(context.Context, *github.Github) error
	PrintStats()
	Count() int
	Name() string
}

type ParamOptions struct {
	Column string
	Asc    bool
	MaxID  int
	Since  int
}

type report struct {
	name             string
	query            github.Query
	repoCount        int
	aggregatedErrors []error
}

const (
	StarGazersReportType = "1"
	LicenseReportType    = "2"

	starGazersReportName   = "StarGazers Report"
	licenseTypesReportName = "License Types Report"

	bucketCol  = "bucket"
	starCol    = "stars"
	repoCol    = "repos"
	licenseCol = "license type"
)

func columnOptions() func(string, string) string {
	bucket := map[string]string{
		"1": bucketCol,
		"2": repoCol,
		"3": starCol,
	}

	license := map[string]string{
		"1": bucketCol,
		"2": repoCol,
		"3": starCol,
	}

	return func(reportType, key string) string {
		switch reportType {
		case LicenseReportType:
			return license[key]
		default:
			return bucket[key]
		}
	}
}

func NewReport(reportType string, opts ParamOptions) (StatsReport, error) {
	if err := validateIDRange(opts.Since, opts.MaxID); err != nil {
		return nil, err
	}

	switch reportType {
	case StarGazersReportType:
		return &BucketReport{
			ParamOptions: opts,
			report: report{
				name:  starGazersReportName,
				query: github.Query{Since: opts.Since, MaxID: opts.MaxID},
			},
		}, nil
	case LicenseReportType:
		return &LicenseTypeReport{
			ParamOptions: opts,
			report: report{
				name:  licenseTypesReportName,
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
