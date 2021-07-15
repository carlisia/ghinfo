package github

type StarBuckets struct {
	Bucket1 int // 0-10
	Bucket2 int // 10-100
	Bucket3 int // 100-1000
	Bucket4 int // 1000-5000
	Bucket5 int // 5000-10000
	Bucket6 int // 10000+
}

// AggregateStarStats returns the statistics on number of stars on the repositories
// specified. return (bucket, # of repos) where bucket is one of:
// 0-10, 10-100, 100-1000, 1000-5000, 5000-10000, 10000+
func AggregateStarStats(stars []Star) StarBuckets {
	var bucket StarBuckets
	for _, star := range stars {
		switch {
		case star.Count <= 10:
			bucket.Bucket1++
		case star.Count <= 100:
			bucket.Bucket2++
		case star.Count <= 1000:
			bucket.Bucket3++
		case star.Count <= 5000:
			bucket.Bucket4++
		case star.Count <= 10000:
			bucket.Bucket5++
		default:
			bucket.Bucket6++
		}
	}
	return bucket
}
