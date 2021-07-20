package analytics

import (
	"context"
	"fmt"

	"github.com/carlisia/ghinfo/github"
	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
)

type BucketReport struct {
	ParamOptions ParamOptions
	starInfo     map[string]map[int]int
	report       report
	sortColumn   sortColumn
}

func (b *BucketReport) Run(ctx context.Context, gh *github.Github) error {
	repos, err := queryRepos(ctx, gh, b.report.query)
	if err != nil {
		return err
	}

	s := b.ParamOptions.SortColumn
	switch s {
	case "1":
		b.sortColumn = sortColumn{
			asc:    true,
			column: bucketCol,
		}
	case "2":
		b.sortColumn = sortColumn{
			asc:    false,
			column: bucketCol,
		}
	case "3":
		b.sortColumn = sortColumn{
			asc:    true,
			column: repoCol,
		}
	case "4":
		b.sortColumn = sortColumn{
			asc:    false,
			column: repoCol,
		}
	}

	b.report.repoCount = len(repos)

	repoInfo, errs := gh.QueryStars(ctx, repos)
	// TODO: include a prompt asking the user if errors should be displayed)
	if len(errs) > 0 {
		b.report.aggregatedErrors = append(b.report.aggregatedErrors, errs...)
	}
	b.starInfo = repoInfo

	return nil
}

func (b *BucketReport) Count() int {
	return b.report.repoCount
}

func (b *BucketReport) PrintStats() {
	asc := b.sortColumn.asc
	orderColumn := b.sortColumn.column

	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"bucket", "#repos", "bucket total stars", "avg stars/repo"})

	var allBucketsStarCount, allBucketsRepoCount int
	buckets := b.starInfo

	fmt.Println("Printing a star bucket report...")
	fmt.Println("Ordering by column: ", orderColumn)
	fmt.Println("Sorting by asc?: ", asc)

	// TODO: Implement sorting by repo count
	if orderColumn == repoCol {
		fmt.Println("Sorting by repo column not yet implemented. Sorting by bucket count.")
	}
	keysBuckets := reportKeys(buckets)
	keysBuckets = sortStringKeys(keysBuckets, asc)

	for _, tier := range keysBuckets {
		tierRepoStarCount := buckets[tier]
		keysRepos := intIntKeys(tierRepoStarCount)

		var bucketNumRepos, bucketStarCount int
		for _, v := range keysRepos {
			bucketNumRepos += v
			bucketStarCount += tierRepoStarCount[v]
		}

		var repoAverage float64
		if bucketNumRepos > 0 {
			repoAverage = float64(bucketStarCount) / float64(bucketNumRepos)
		}

		allBucketsStarCount = allBucketsStarCount + bucketStarCount
		allBucketsRepoCount += bucketNumRepos
		tw.AppendRows([]table.Row{
			{tier, bucketNumRepos, bucketStarCount, repoAverage},
		})
	}

	tw.AppendFooter(table.Row{"total", allBucketsRepoCount, allBucketsStarCount, ""})
	tw.SetStyle(table.StyleRounded)
	tw.Style().Format.Header = text.FormatLower
	tw.Style().Format.Row = text.FormatLower
	tw.Style().Format.Footer = text.FormatLower
	fmt.Printf("Report of total number of repositories and stars per bucket:\n%s", tw.Render())
}
