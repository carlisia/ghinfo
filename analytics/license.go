package analytics

import (
	"context"
	"fmt"
	"sort"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"

	"github.com/carlisia/ghinfo/github"
)

type LicenseTypeReport struct {
	ParamOptions ParamOptions
	report       report
	aggregate    []aggregateLicense
}

type aggregateLicense struct {
	license   string
	repoCount int
}

func (l *LicenseTypeReport) Run(ctx context.Context, gh *github.Github) error {
	repos, err := queryRepos(ctx, gh, l.report.query)
	if err != nil {
		return err
	}
	l.report.repoCount = len(repos)

	fmt.Print("Getting license type information for each repository found...\n\n")

	repoInfo := gh.QueryLicenses(ctx, repos)

	i := 0
	l.aggregate = make([]aggregateLicense, len(repoInfo))
	for k, v := range repoInfo {
		l.aggregate[i].license = k
		l.aggregate[i].repoCount = v
		i++
	}

	l.sort()

	return nil
}

func (l *LicenseTypeReport) Count() int {
	return l.report.repoCount
}

func (l *LicenseTypeReport) Name() string {
	return l.report.name
}

func (l *LicenseTypeReport) sort() {
	licenses := l.aggregate
	l.ParamOptions.Column = columnOptions()(LicenseReportType, l.ParamOptions.Column)
	sort.Slice(licenses, func(i, j int) bool {
		var res bool
		switch l.ParamOptions.Column {
		case repoCol:
			res = licenses[i].repoCount < licenses[j].repoCount
		default:
			res = licenses[i].license < licenses[j].license
		}

		if !l.ParamOptions.Asc {
			return !res
		}
		return res
	})
}

func (l *LicenseTypeReport) PrintStats() {
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"license type", "#repos"})

	fmt.Println("Printing a license report...")
	fmt.Println("Ordering by column: ", l.ParamOptions.Column)
	fmt.Printf("Sorting by asc?: %v\n\n", l.ParamOptions.Asc)

	var allLicensessRepoCount int
	for _, license := range l.aggregate {
		allLicensessRepoCount += license.repoCount
		tw.AppendRows([]table.Row{
			{license, license.repoCount},
		})
	}

	tw.AppendFooter(table.Row{"total", allLicensessRepoCount})
	tw.SetStyle(table.StyleRounded)
	tw.Style().Format.Header = text.FormatLower
	tw.Style().Format.Row = text.FormatLower
	tw.Style().Format.Footer = text.FormatLower
	fmt.Printf("Report of total number of repositories per license:\n%s", tw.Render())
}
