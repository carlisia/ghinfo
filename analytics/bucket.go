package analytics

import (
	"context"
	"fmt"
	"sort"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"

	"github.com/carlisia/ghinfo/github"
)

type BucketReport struct {
	ParamOptions ParamOptions
	report       report
	aggregate    []aggregateBucket
}

type aggregateBucket struct {
	bucket               string
	repoCount, starCount int
}

func (b *BucketReport) Run(ctx context.Context, gh *github.Github) error {
	repos, err := queryRepos(ctx, gh, b.report.query)
	if err != nil {
		return err
	}
	b.report.repoCount = len(repos)

	fmt.Print("Getting star gazers information for each repository found...\n\n")

	repoInfo, errs := gh.QueryStars(ctx, repos)
	// TODO: include a prompt asking the user if errors should be displayed)
	if len(errs) > 0 {
		b.report.aggregatedErrors = append(b.report.aggregatedErrors, errs...)
	}

	i := 0
	b.aggregate = make([]aggregateBucket, len(repoInfo))
	for k, v := range repoInfo {
		b.aggregate[i].bucket = k
		for kk, vv := range v {
			b.aggregate[i].repoCount = kk
			b.aggregate[i].starCount = vv
		}
		i++
	}

	b.sort()

	return nil
}

func (b *BucketReport) Count() int {
	return b.report.repoCount
}

func (b *BucketReport) Name() string {
	return b.report.name
}

func (b *BucketReport) sort() {
	buckets := b.aggregate
	b.ParamOptions.Column = columnOptions()(StarGazersReportType, b.ParamOptions.Column)
	sort.Slice(buckets, func(i, j int) bool {
		var res bool
		switch b.ParamOptions.Column {
		case repoCol:
			res = buckets[i].repoCount < buckets[j].repoCount
		case starCol:
			res = buckets[i].starCount < buckets[j].starCount
		default:
			res = buckets[i].bucket < buckets[j].bucket
		}

		if !b.ParamOptions.Asc {
			return !res
		}
		return res
	})
}

func (b *BucketReport) PrintStats() {
	tw := table.NewWriter()
	tw.AppendHeader(table.Row{"bucket", "#repos", "bucket total stars", "avg stars/repo"})

	fmt.Println("Printing a star bucket report...")
	fmt.Println("Ordering by column: ", b.ParamOptions.Column)
	fmt.Printf("Sorting by asc?: %v\n\n", b.ParamOptions.Asc)

	var allBucketsStarCount, allBucketsRepoCount int
	for _, bucket := range b.aggregate {
		var repoAverage float64
		if bucket.repoCount > 0 {
			repoAverage = float64(bucket.starCount) / float64(bucket.repoCount)
		}

		allBucketsStarCount = allBucketsStarCount + bucket.starCount
		allBucketsRepoCount += bucket.repoCount
		tw.AppendRows([]table.Row{
			{bucket.bucket, bucket.repoCount, bucket.starCount, repoAverage},
		})
	}

	tw.AppendFooter(table.Row{"total", allBucketsRepoCount, allBucketsStarCount, ""})
	tw.SetStyle(table.StyleRounded)
	tw.Style().Format.Header = text.FormatLower
	tw.Style().Format.Row = text.FormatLower
	tw.Style().Format.Footer = text.FormatLower
	fmt.Printf("Report of total number of repositories and stars per bucket:\n%s", tw.Render())
}
