package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Name             string
		Args             []string
		ExpectedThrows   []ThrowGroup
		ExpectedColCount int
	}{
		{
			Name:             "no args",
			Args:             []string{},
			ExpectedThrows:   nil,
			ExpectedColCount: 0,
		},
		{
			Name: "single arg",
			Args: []string{"d6"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
						}: 1,
					},
				},
			},
			ExpectedColCount: 1,
		},
		{
			Name: "multiple args",
			Args: []string{"d6", "2d10", "d%"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
						}: 1,
					},
				},
				{
					Counts: map[FCount]int{
						{
							FaceCount: 10,
						}: 2,
					},
				},
				{
					Counts: map[FCount]int{
						{
							FaceCount:  10,
							Percentile: true,
						}: 1,
					},
				},
			},
			ExpectedColCount: 2,
		},
		{
			Name: "invalid arg",
			Args: []string{"d6", "d0"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
						}: 1,
					},
				},
			},
			ExpectedColCount: 1,
		},
		{
			Name: "negative arg",
			Args: []string{"-1d2", "d6"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
						}: 1,
					},
				},
			},
			ExpectedColCount: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			groups, colCount := parseArgs(test.Args)

			assert.Equal(t, test.ExpectedThrows, groups)
			assert.Equal(t, test.ExpectedColCount, colCount)
		})
	}
}

func TestRoll(t *testing.T) {
	t.Parallel()

	tests := []*struct {
		Name     string
		Group    ThrowGroup
		Expected []struct {
			FCount
			Min int
			Max int
		}
	}{
		{
			Name: "d6",
			Group: ThrowGroup{
				Counts: map[FCount]int{
					{
						FaceCount: 6,
					}: 1,
				},
			},
			Expected: []struct {
				FCount
				Min int
				Max int
			}{
				{
					FCount: FCount{
						FaceCount: 6,
					},
					Min: 1,
					Max: 6,
				},
			},
		},
		{
			Name: "2d10",
			Group: ThrowGroup{
				Counts: map[FCount]int{
					{
						FaceCount: 10,
					}: 2,
				},
			},
			Expected: []struct {
				FCount
				Min int
				Max int
			}{
				{
					FCount: FCount{
						FaceCount: 10,
					},
					Min: 1,
					Max: 10,
				},
				{
					FCount: FCount{
						FaceCount: 10,
					},
					Min: 1,
					Max: 10,
				},
			},
		},
		{
			Name: "d%+d10",
			Group: ThrowGroup{
				Counts: map[FCount]int{
					{
						FaceCount:  10,
						Percentile: true,
					}: 1,
					{
						FaceCount: 10,
					}: 1,
				},
			},
			Expected: []struct {
				FCount
				Min int
				Max int
			}{
				{
					FCount: FCount{
						FaceCount:  10,
						Percentile: true,
					},
					Min: 0,
					Max: 90,
				},
				{
					FCount: FCount{
						FaceCount: 10,
					},
					Min: 1,
					Max: 10,
				},
			},
		},
		{
			Name: "d64",
			Group: ThrowGroup{
				Counts: map[FCount]int{
					{
						FaceCount: 64,
					}: 1,
				},
			},
			Expected: []struct {
				FCount
				Min int
				Max int
			}{
				{
					FCount: FCount{
						FaceCount: 64,
					},
					Min: 1,
					Max: 64,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			runs := 100

			if testing.Short() {
				runs = 10
			}

			for range runs {
				outcomes := test.Group.Roll()
				sortThrows(outcomes)

				assert.Equal(t, len(test.Expected), len(outcomes))

				for i, out := range outcomes {
					assert.Equal(t, test.Expected[i].FCount, out.FCount)
					assert.Equal(t, test.Expected[i].Percentile, out.Percentile)
					assert.LessOrEqual(t, out.Number, test.Expected[i].Max, "run #%d / %d: expected %d to be less than or equal to %d", runs, i, out.Number, test.Expected[i].Max)
					assert.GreaterOrEqual(t, out.Number, test.Expected[i].Min, "run #%d / %d: expected %d to be greater than or equal to %d", runs, i, out.Number, test.Expected[i].Min)
				}
			}
		})
	}
}
