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
			Name: "repeated face count in groups",
			Args: []string{"3d6+2d6"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
						}: 5,
					},
				},
			},
			ExpectedColCount: 5,
		},
		{
			Name: "constant",
			Args: []string{"d6+3"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
						}: 1,
						{
							Constant: true,
						}: 3,
					},
				},
			},
			ExpectedColCount: 2,
		},
		{
			Name: "merge constants",
			Args: []string{"d6+1+2"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
						}: 1,
						{
							Constant: true,
						}: 3,
					},
				},
			},
			ExpectedColCount: 2,
		},
		{
			Name: "negative constant",
			Args: []string{"d6-1"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
						}: 1,
						{
							Constant: true,
						}: -1,
					},
				},
			},
			ExpectedColCount: 2,
		},
		{
			Name: "merge negative constants",
			Args: []string{"d6+1-2"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
						}: 1,
						{
							Constant: true,
						}: -1,
					},
				},
			},
			ExpectedColCount: 2,
		},
		{
			Name: "fudge dice",
			Args: []string{"dF"},
			ExpectedThrows: []ThrowGroup{
				{
					Counts: map[FCount]int{
						{
							FaceCount: 6,
							Fudge:     true,
						}: 1,
					},
				},
			},
			ExpectedColCount: 1,
		},
		{
			Name: "invalid arg",
			Args: []string{"d6", "d0", "Hd2"},
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
		{
			Name: "Constants",
			Group: ThrowGroup{
				Counts: map[FCount]int{
					{
						FaceCount: 6,
					}: 1,
					{
						Constant: true,
					}: 3,
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
				{
					FCount: FCount{
						Constant:  true,
						FaceCount: 0,
					},
					Min: 3,
					Max: 3,
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

func TestSplitOnOperators(t *testing.T) {
	tests := []struct {
		Name     string
		Arg      string
		Expected []string
	}{
		{
			Name:     "empty",
			Arg:      "",
			Expected: []string{""},
		},
		{
			Name:     "single",
			Arg:      "d6",
			Expected: []string{"d6"},
		},
		{
			Name:     "plus",
			Arg:      "d6+d10",
			Expected: []string{"d6", "d10"},
		},
		{
			Name:     "minus",
			Arg:      "1-2",
			Expected: []string{"1", "-2"},
		},
		{
			Name:     "plus and minus",
			Arg:      "1+2-3",
			Expected: []string{"1", "2", "-3"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			grouped := splitOnOperators(test.Arg)

			assert.Equal(t, test.Expected, grouped)
		})
	}
}

func TestSortThrows(t *testing.T) {
	tests := []struct {
		Name     string
		Throws   []Throw
		NewOrder []Throw
	}{
		{
			Name:     "empty",
			Throws:   []Throw{},
			NewOrder: []Throw{},
		},
		{
			Name: "single",
			Throws: []Throw{
				NewThrow(6, 1),
			},
			NewOrder: []Throw{
				NewThrow(6, 1),
			},
		},
		{
			Name: "Bigger Face Counts First",
			Throws: []Throw{
				NewThrow(10, 1),
				NewThrow(20, 20),
			},
			NewOrder: []Throw{
				NewThrow(20, 20),
				NewThrow(10, 1),
			},
		},
		{
			Name: "Bigger Numbers First",
			Throws: []Throw{
				NewThrow(6, 1),
				NewThrow(6, 6),
				NewThrow(6, 3),
			},
			NewOrder: []Throw{
				NewThrow(6, 6),
				NewThrow(6, 3),
				NewThrow(6, 1),
			},
		},
		{
			Name: "Percentile Before 10",
			Throws: []Throw{
				NewThrow(10, 1),
				{
					FCount: FCount{
						FaceCount:  10,
						Percentile: true,
					},
					Number: 1,
				},
			},
			NewOrder: []Throw{
				{
					FCount: FCount{
						FaceCount:  10,
						Percentile: true,
					},
					Number: 1,
				},
				NewThrow(10, 1),
			},
		},
		{
			Name: "Fudge After 6",
			Throws: []Throw{
				{
					FCount: FCount{
						FaceCount: 6,
						Fudge:     true,
					},
					Number: 2,
				},
				NewThrow(6, 1),
			},
			NewOrder: []Throw{
				NewThrow(6, 1),
				{
					FCount: FCount{
						FaceCount: 6,
						Fudge:     true,
					},
					Number: 2,
				},
			},
		},
		{
			Name: "Constants After All",
			Throws: []Throw{
				{
					FCount: FCount{
						FaceCount: 10,
						Constant:  true,
					},
				},
				{
					FCount: FCount{
						FaceCount:  10,
						Percentile: true,
					},
					Number: 5,
				},
				{
					FCount: FCount{
						FaceCount: 6,
						Fudge:     true,
					},
					Number: 2,
				},
				NewThrow(10, 9),
				NewThrow(6, 1),
				NewThrow(6, 6),
			},
			NewOrder: []Throw{
				{
					FCount: FCount{
						FaceCount:  10,
						Percentile: true,
					},
					Number: 5,
				},
				NewThrow(10, 9),
				NewThrow(6, 6),
				NewThrow(6, 1),
				{
					FCount: FCount{
						FaceCount: 6,
						Fudge:     true,
					},
					Number: 2,
				},
				{
					FCount: FCount{
						FaceCount: 10,
						Constant:  true,
					},
				},
			},
		},
		{
			Name: "Sort Constants",
			Throws: []Throw{
				{
					FCount: FCount{
						FaceCount: 10,
						Constant:  true,
					},
				},
				{
					FCount: FCount{
						FaceCount: 20,
					},
					Number: 10,
				},
				{
					FCount: FCount{
						FaceCount: 5,
						Constant:  true,
					},
				},
			},
			NewOrder: []Throw{
				{
					FCount: FCount{
						FaceCount: 20,
					},
					Number: 10,
				},
				{
					FCount: FCount{
						FaceCount: 5,
						Constant:  true,
					},
				},
				{
					FCount: FCount{
						FaceCount: 10,
						Constant:  true,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			sortThrows(test.Throws)

			assert.Equal(t, test.NewOrder, test.Throws)
		})
	}
}
