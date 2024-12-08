package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

type DeterministicSource struct {
	data []byte
	pos  int
	err  string
}

func NewDeterministicSource(len int) *DeterministicSource {
	return &DeterministicSource{
		data: bytes.Repeat([]byte{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5, 8, 9, 7, 9, 3, 2, 3, 8, 4, 6, 2, 6, 4, 3, 3, 8, 3, 2, 7, 9, 5, 0, 2, 8, 8, 4, 1, 9, 7, 1, 6, 9, 3, 9, 9, 3, 7, 5, 1, 0, 5, 8, 2, 0, 9, 7, 4, 9, 4, 4, 5, 9, 2, 3, 0, 7, 8, 1, 6, 4, 0, 6, 2, 8, 6, 2, 0, 8, 9, 9, 8, 6, 2, 8, 0, 3, 4, 8, 2, 5}, len),
		pos:  0,
		err:  "",
	}
}

func (d *DeterministicSource) Read(p []byte) (n int, err error) {
	if d.err != "" {
		return 0, errors.New(d.err)
	}

	if d.pos >= len(d.data) {
		return 0, io.EOF
	}

	n = copy(p, d.data[d.pos:])
	d.pos += n

	return n, nil
}

func setup() {
	randSource = NewDeterministicSource(5)
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
	setup()

	os.Args = []string{"dice", "d6", "8D20", "2d4+2d6", "6d%+6d10", "10d%", "1d6-1", "6dF", "8-4"}
	output := captureOutput(main)

	expectedOutput := `
 d6(  4 )                                                                                                                                   =   4
d20( 10 ) + d20(  7 ) + d20(  6 ) + d20(  6 ) + d20(  5 ) + d20(  3 ) + d20(  2 ) + d20(  2 )                                               =  41
 d6(  2 ) +  d6(  1 ) +  d4(  4 ) +  d4(  2 )                                                                                               =   9
 d%( 90 ) +  d%( 80 ) +  d%( 70 ) +  d%( 30 ) +  d%( 30 ) +  d%( 20 ) + d10(  7 ) + d10(  7 ) + d10(  5 ) + d10(  5 ) + d10( 4 ) + d10( 3 ) = 351
 d%( 90 ) +  d%( 80 ) +  d%( 80 ) +  d%( 70 ) +  d%( 50 ) +  d%( 30 ) +  d%( 30 ) +  d%( 20 ) +  d%( 20 ) +  d%( 00 )                       = 470
 d6(  1 ) -         1                                                                                                                       =   0
 dF(  1 ) +  dF(  0 ) +  dF( -1 ) +  dF( -1 ) +  dF( -1 ) +  dF( -1 )                                                                       =  -3
        4                                                                                                                                   =   4
                                                                                                                                            = 876
`[1:]

	assert.Equal(t, expectedOutput, output)
}

func TestMainOutputError(t *testing.T) {
	setup()

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
	setup()
	randSource.(*DeterministicSource).err = "error"

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
	setup()
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
			setup()
			throws, colCount := parseArgs(tt.Args)
			assert.Equal(t, tt.ExpectedThrows, throws)
			assert.Equal(t, tt.ExpectedColCount, colCount)
		})
	}
}

func TestRoll(t *testing.T) {
	setup()

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
				NewThrow(6, 4),
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
				NewThrow(10, 4),
				NewThrow(10, 2),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			setup()
			throws := tt.Group.Roll()
			assert.Equal(t, tt.Expected, throws)
		})
	}
}

func TestSortThrows(t *testing.T) {
	setup()

	throws := []Throw{
		NewConstant(4),
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
		NewConstant(2),
		NewConstant(3),
		NewConstant(4),
	}, throws)
}
