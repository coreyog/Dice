package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

// ThrowGroup holds multiple FaceCount die rolls thrown at one time, e.g. 2d6+1d10
type ThrowGroup struct {
	Counts map[FCount]int // map[face count]number of throws
}

// FCount represents a die with a number of faces, e.g. d6 or d10, notably d% for percentile dice
type FCount struct {
	FaceCount  int
	Percentile bool
}

// Throw represents a single die roll, with the number rolled and the FaceCount of the die
type Throw struct {
	FCount
	Number int
}

func main() {
	// parse input
	throws, colCount := parseArgs(os.Args[1:])

	// build output writer
	tab := tabwriter.NewWriter(os.Stdout, 1, 0, 0, ' ', tabwriter.AlignRight)

	// accumulate statistics
	bigTotal := 0

	for _, tg := range throws {
		// roll each group and actually pick the values
		outcomes := tg.Roll()

		// sort outcomes by face count (percentile over non-percentile), then by number
		sortThrows(outcomes)

		// accumulate the total for this throw
		total := 0

		// build the output with the writer
		for i, index := range outcomes {
			if i > 0 {
				_, _ = tab.Write([]byte(" + \t"))
			}

			// take special care of percentile dice's 0 state
			total += index.Number
			numDisplay := strconv.Itoa(index.Number)
			if numDisplay == "0" {
				numDisplay = "00"
			}

			faceDisplay := strconv.Itoa(index.FaceCount)
			if index.Percentile {
				faceDisplay = "%"
			}

			_, _ = tab.Write([]byte(fmt.Sprintf("d%s( \t%s )\t", faceDisplay, numDisplay)))
		}

		if len(throws) > 1 || len(outcomes) > 1 {
			// account for missing columns (each missing column is 3x tabs: after the
			// plus, before the roll value, and after the roll value)
			_, _ = tab.Write(bytes.Repeat([]byte("\t\t\t"), colCount-len(outcomes)))

			fmt.Fprintf(tab, " = \t%d\t", total)
		}

		bigTotal += total

		// prepare for next row of output
		_, _ = tab.Write([]byte("\n"))
	}

	// if more than 1 group input, print TOTAL total
	if len(throws) > 1 {
		count := (colCount)*3 - 1
		fmt.Fprint(tab, strings.Repeat("\t", count))

		fmt.Fprintf(tab, "= \t%d\t", bigTotal)
	}

	// flush the output or you don't see it
	err := tab.Flush()
	if err != nil {
		panic(err)
	}
}

func parseArgs(args []string) (groups []ThrowGroup, columnCount int) {
	for _, arg := range args {
		tg := ThrowGroup{
			Counts: map[FCount]int{},
		}

		// watch for x+y grouped args
		grouped := strings.Split(arg, "+")
		cols := 0
		for _, g := range grouped {
			// split a dice count from it's face count
			parts := strings.Split(strings.ToUpper(g), "D")

			// assume 1 die if no count is given
			if parts[0] == "" {
				parts[0] = "1"
			}

			num, err := strconv.Atoi(parts[0])
			if err != nil || num < 1 {
				fmt.Printf("invalid dice entry: %s\n", arg)
				continue
			}

			var faces int
			var percentile bool

			if parts[1] == "%" {
				// handle percentile special
				faces = 10
				percentile = true
			} else {
				faces, err = strconv.Atoi(parts[1])
				if err != nil || faces < 1 {
					fmt.Printf("invalid dice entry: %s\n", arg)
					continue
				}
			}

			// accumulate total columns for this group
			cols += num

			fc := FCount{
				FaceCount:  faces,
				Percentile: percentile,
			}

			_, exists := tg.Counts[fc]
			if exists {
				tg.Counts[fc] += num
			} else {
				tg.Counts[fc] = num
			}
		}

		// keep track of longest row for formatting
		columnCount = max(columnCount, cols)

		if len(tg.Counts) > 0 {
			groups = append(groups, tg)
		}
	}

	return groups, columnCount
}

func sortThrows(outcomes []Throw) {
	sort.Slice(outcomes, func(i, j int) bool {
		if outcomes[i].FaceCount == outcomes[j].FaceCount {
			if outcomes[i].Percentile != outcomes[j].Percentile {
				return outcomes[i].Percentile
			}
			return outcomes[i].Number > outcomes[j].Number
		}

		return outcomes[i].FaceCount > outcomes[j].FaceCount
	})
}

func (tg ThrowGroup) Roll() []Throw {
	results := make([]Throw, 0, len(tg.Counts))

	for faceCount, number := range tg.Counts {
		for range number {
			// pull the best random number you can get
			n, err := rand.Int(rand.Reader, big.NewInt(int64(faceCount.FaceCount)))
			if err != nil {
				panic(err)
			}

			// normalize
			selected := int(n.Int64())
			if faceCount.Percentile {
				selected *= 10
			} else {
				selected++
			}

			// accumulate results
			results = append(results, Throw{
				FCount: faceCount,
				Number: selected,
			})
		}
	}

	return results
}
