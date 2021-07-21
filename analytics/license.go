package analytics

import (
	"context"
	"fmt"

	"github.com/carlisia/ghinfo/github"
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
)

type LicenseTypeReport struct {
	ParamOptions ParamOptions
	licenseInfo  map[string]map[string]int
	report       report
	sortColumn   sortColumn
}

func (l *LicenseTypeReport) Run(ctx context.Context, gh *github.Github) error {
	repos, err := queryRepos(ctx, gh, l.report.query)
	if err != nil {
		return err
	}
	l.report.repoCount = len(repos)

	s := l.ParamOptions.SortColumn
	switch s {
	case "1":
		l.sortColumn = sortColumn{
			asc:    true,
			column: licenseCol,
		}
	case "2":
		l.sortColumn = sortColumn{
			asc:    false,
			column: licenseCol,
		}
	case "3":
		l.sortColumn = sortColumn{
			asc:    true,
			column: repoCol,
		}
	case "4":
		l.sortColumn = sortColumn{
			asc:    false,
			column: repoCol,
		}
	}

	fmt.Print("Getting license type information for each repository found...\n\n")

	repoInfo := gh.QueryLicenses(ctx, repos)
	l.licenseInfo = repoInfo
	return nil
}

func (l *LicenseTypeReport) Count() int {
	return l.report.repoCount
}

func (l *LicenseTypeReport) Name() string {
	return l.report.name
}

func (l *LicenseTypeReport) PrintStats() {
	asc := l.sortColumn.asc
	orderColumn := l.sortColumn.column

	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"license type", "#repos"})

	var allLicensessRepoCount int
	licenses := l.licenseInfo

	fmt.Println("Printing a license report...")
	fmt.Println("Ordering by column: ", orderColumn)
	fmt.Printf("Sorting by asc?: %v\n\n", asc)

	// TODO: Implement sorting by repo count
	if orderColumn == repoCol {
		fmt.Println("Sorting by repo column not yet implemented. Sorting by license type.")
	}
	keysLicenses := reportLicenseKeys(licenses)
	keysLicenses = sortStringKeys(keysLicenses, asc)

	for _, licenseName := range keysLicenses {
		licenseTypeRepoCount := licenses[licenseName]

		var licenseRepoCount int
		for _, v := range licenseTypeRepoCount {
			licenseRepoCount += v
		}

		allLicensessRepoCount += licenseRepoCount
		if orderColumn == licenseCol {
			tw.AppendRows([]table.Row{
				{licenseName, licenseRepoCount},
			})
		}
	}

	tw.AppendFooter(table.Row{"total", allLicensessRepoCount})
	tw.SetStyle(table.StyleRounded)
	tw.Style().Format.Header = text.FormatLower
	tw.Style().Format.Row = text.FormatLower
	tw.Style().Format.Footer = text.FormatLower
	fmt.Printf("Report of total number of repositories per license:\n%s", tw.Render())
}
