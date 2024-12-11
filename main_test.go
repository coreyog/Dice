package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	randNum = func(n int) int {
		return 0
	}
}

func NewThrow(f int, n int) Throw {
	return Throw{
		FCount: FCount{
			FaceCount: f,
		},
		Number: n,
	}
}

func NewConstant(n int) Throw {
	return Throw{
		FCount: FCount{
			FaceCount: n,
			Constant:  true,
		},
	}
}

func NewPercentile(n int) Throw {
	if n < 10 {
		n *= 10
	}

	return Throw{
		FCount: FCount{
			FaceCount:  10,
			Percentile: true,
		},
		Number: n,
	}
}

func NewFudge(n int) Throw {
	return Throw{
		FCount: FCount{
			FaceCount: 6,
			Fudge:     true,
		},
		Number: n,
	}
}

func captureOutput(f func()) string {
	r, w, _ := os.Pipe()
	old := os.Stdout
	os.Stdout = w

	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = buf.ReadFrom(r)
		outC <- buf.String()
	}()

	f()
	_ = w.Close()
	os.Stdout = old
	return <-outC
}

func TestMainFunc(t *testing.T) {
	orig := randNum
	defer func() {
		randNum = orig
	}()

	index := 0
	arr := []int{1, 6, 4, 5, 1, 3, 0, 8, 1, 3, 0, 2, 4}
	randNum = func(n int) int {
		ret := arr[index]
		index++

		return ret
	}

	os.Args = []string{"dice", "d6", "2D20", "2d4+2d6", "d%+d10", "1d6-1", "3dF", "8-4"}
	output := captureOutput(main)

	expectedOutput := `
 d6(  2 )                                 =   2
d20(  7 ) + d20( 5 )                      =  12
 d6(  4 ) +  d6( 1 ) + d4(  6 ) + d4( 2 ) =  13
 d%( 80 ) + d10( 2 )                      =  82
 d6(  4 ) -        1                      =   3
 dF(  1 ) +  dF( 0 ) + dF( -1 )           =   0
        4                                 =   4
                                          = 116
`[1:]

	assert.Equal(t, expectedOutput, output)
}

func TestMainOutputError(t *testing.T) {
	old := os.Stdout
	defer func() {
		os.Stdout = old
	}()

	os.Args = []string{"dice", "d6", "8D20", "2d4+2d6", "6d%+6d10", "2d6+10", "1d6-1", "6dF", "8-4", "d0"}

	_, w, _ := os.Pipe()
	os.Stdout = w
	w.Close()

	assert.Panics(t, main)
}

func TestMainRandSourceError(t *testing.T) {
	orig := randNum
	randNum = func(n int) int {
		panic("something went wrong")
	}
	defer func() {
		randNum = orig
	}()

	old := os.Stdout
	defer func() {
		os.Stdout = old
	}()

	os.Args = []string{"dice", "d6", "8D20", "2d4+2d6", "6d%+6d10", "2d6+10", "1d6-1", "6dF", "8-4", "d0"}

	_, w, _ := os.Pipe()
	os.Stdout = w

	assert.Panics(t, main)
}

func TestParseArgs(t *testing.T) {
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
			Args: []string{"d6+3", "2+d4"},
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
				{
					Counts: map[FCount]int{
						{
							FaceCount: 4,
						}: 1,
						{
							Constant: true,
						}: 2,
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
			Args: []string{"d6", "d0", "Hd2", "-1d6"},
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

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			throws, colCount := parseArgs(tt.Args)
			assert.Equal(t, tt.ExpectedThrows, throws)
			assert.Equal(t, tt.ExpectedColCount, colCount)
		})
	}
}

func TestRoll(t *testing.T) {
	tests := []*struct {
		Name     string
		Group    ThrowGroup
		Expected []Throw
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
			Expected: []Throw{
				NewThrow(6, 1),
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
			Expected: []Throw{
				NewThrow(10, 1),
				NewThrow(10, 1),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			throws := tt.Group.Roll()
			assert.Equal(t, tt.Expected, throws)
		})
	}
}

func TestSortThrows(t *testing.T) {
	throws := []Throw{
		NewConstant(4),
		NewConstant(1),
		NewThrow(4, 2),
		NewThrow(4, 3),
		NewThrow(10, 5),
		NewPercentile(30),
		NewConstant(3),
		NewConstant(2),
		NewThrow(6, 4),
		NewFudge(-1),
		NewFudge(1),
		NewThrow(4, 1),
		NewThrow(6, 2),
	}

	sortThrows(throws)

	assert.Equal(t, []Throw{
		NewPercentile(30),
		NewThrow(10, 5),
		NewThrow(6, 4),
		NewThrow(6, 2),
		NewFudge(1),
		NewFudge(-1),
		NewThrow(4, 3),
		NewThrow(4, 2),
		NewThrow(4, 1),
		NewConstant(1),
		NewConstant(2),
		NewConstant(3),
		NewConstant(4),
	}, throws)
}

func TestSortFCounts(t *testing.T) {
	faces := []FCount{
		{FaceCount: 6},
		{FaceCount: 2, Constant: true},
		{FaceCount: 1, Constant: true},
		{FaceCount: 10, Percentile: true},
		{FaceCount: 6, Fudge: true},
		{FaceCount: 6},
		{FaceCount: 10},
		{FaceCount: 10, Percentile: true},
		{FaceCount: 3, Constant: true},
	}

	SortFCounts(faces)

	assert.Equal(t, []FCount{
		{FaceCount: 10, Percentile: true},
		{FaceCount: 10, Percentile: true},
		{FaceCount: 10},
		{FaceCount: 6},
		{FaceCount: 6},
		{FaceCount: 6, Fudge: true},
		{FaceCount: 1, Constant: true},
		{FaceCount: 2, Constant: true},
		{FaceCount: 3, Constant: true},
	}, faces)
}
