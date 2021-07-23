package github

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_trimUpToMaxID(t *testing.T) {
	type args struct {
		maxID   int
		repoSet []Repos
	}

	tc := []struct {
		name string
		args args
		want []Repos
	}{
		{
			name: "set includes only the max ID",
			args: args{maxID: 10, repoSet: []Repos{{ID: 10}}},
			want: []Repos{{ID: 10}},
		},
		{
			name: "set includes only IDs below max ID",
			args: args{maxID: 10, repoSet: []Repos{{ID: 5}, {ID: 6}, {ID: 7}, {ID: 8}}},
			want: []Repos{{ID: 5}, {ID: 6}, {ID: 7}, {ID: 8}},
		},
		{
			name: "set includes only IDs above max ID",
			args: args{maxID: 10, repoSet: []Repos{{ID: 15}, {ID: 16}, {ID: 17}, {ID: 18}}},
			want: []Repos{},
		},
		{
			name: "set includes IDs starting at max and above",
			args: args{maxID: 10, repoSet: []Repos{{ID: 10}, {ID: 16}, {ID: 17}, {ID: 18}}},
			want: []Repos{{ID: 10}},
		},
		{
			name: "set includes IDs below and up to max",
			args: args{maxID: 10, repoSet: []Repos{{ID: 3}, {ID: 8}, {ID: 10}}},
			want: []Repos{{ID: 3}, {ID: 8}, {ID: 10}},
		},
		{
			name: "set includes IDs below and above max, and max id included in the set",
			args: args{maxID: 10, repoSet: []Repos{{ID: 3}, {ID: 5}, {ID: 10}, {ID: 20}, {ID: 30}}},
			want: []Repos{{ID: 3}, {ID: 5}, {ID: 10}},
		},
		{
			name: "set includes IDs below and above max, and max id not included in the set",
			args: args{maxID: 10, repoSet: []Repos{{ID: 3}, {ID: 5}, {ID: 50}}},
			want: []Repos{{ID: 3}, {ID: 5}},
		},
	}

	for _, tc := range tc {
		t.Run(tc.name, func(t *testing.T) {
			repos := trimUpToMaxID(tc.args.maxID, tc.args.repoSet)

			require.Len(t, repos, len(tc.want))
			for k := range tc.want {
				require.Equal(t, tc.want[k], repos[k])
			}
		})
	}
}
