package main

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type ConstantSource struct {
	seed byte
}

func (d *ConstantSource) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = d.seed
	}

	return len(p), nil
}

type DeterministicSource struct {
	seed byte
}

func (d *DeterministicSource) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = d.seed
		d.seed += 3
	}

	return len(p), nil
}

type PanickySource struct{}

func (d *PanickySource) Read(p []byte) (n int, err error) {
	return 0, errors.New("something went wrong")
}

func TestMain(m *testing.M) {
	randSource = &ConstantSource{seed: 1}

	m.Run()
}

func captureOutput(f func()) string {
	randSource = &DeterministicSource{seed: 1}
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
	os.Args = []string{"dice", "d6", "8D20", "2d4+2d6", "6d%+6d10", "2d6+10", "1d6-1", "6dF", "8-4"}
	output := captureOutput(main)

	expectedOutput := `
 d6(  2 )                                                                                                                               =   2
d20( 20 ) + d20( 17 ) + d20( 14 ) + d20( 11 ) + d20(  8 ) + d20(  6 ) + d20( 5 ) + d20( 3 )                                             =  84
 d6(  5 ) +  d6(  2 ) +  d4(  4 ) +  d4(  1 )                                                                                           =  12
 d%( 90 ) +  d%( 70 ) +  d%( 60 ) +  d%( 30 ) +  d%( 20 ) +  d%( 00 ) + d10( 9 ) + d10( 8 ) + d10( 6 ) + d10( 5 ) + d10( 2 ) + d10( 1 ) = 301
 d6(  4 ) +  d6(  2 ) +        10                                                                                                       =  16
 d6(  5 ) -         1                                                                                                                   =   4
 dF(  1 ) +  dF(  1 ) +  dF(  0 ) +  dF(  0 ) +  dF( -1 ) +  dF( -1 )                                                                   =   0
        4                                                                                                                               =   4
                                                                                                                                        = 423
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

	assert.Panics(t, func() {
		main()
	})
}

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
				NewThrow(6, 2),
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
				NewThrow(10, 2),
				NewThrow(10, 2),
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
			Expected: []Throw{
				{
					FCount: FCount{
						FaceCount:  10,
						Percentile: true,
					},
					Number: 10,
				},
				NewThrow(10, 2),
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
			Expected: []Throw{
				NewThrow(64, 2),
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
			Expected: []Throw{
				NewThrow(6, 2),
				{
					FCount: FCount{
						Constant: true,
					},
					Number: 3,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			outcomes := test.Group.Roll()
			sortThrows(outcomes)

			assert.Equal(t, len(test.Expected), len(outcomes))

			for i, out := range outcomes {
				assert.Equal(t, test.Expected[i].FCount, out.FCount)
				assert.Equal(t, test.Expected[i].Percentile, out.Percentile)
				assert.Equal(t, test.Expected[i].Number, out.Number)
			}
		})
	}

	t.Run("Panicky Source", func(t *testing.T) {
		randSource = &PanickySource{}
		defer func() {
			randSource = &ConstantSource{seed: 1}
		}()

		group := ThrowGroup{
			Counts: map[FCount]int{
				{
					FaceCount: 6,
				}: 1,
			},
		}

		assert.Panics(t, func() {
			group.Roll()
		})
	})
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
