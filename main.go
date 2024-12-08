package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
)

var randSource = rand.Reader

// ThrowGroup holds multiple FaceCount die rolls thrown at one time, e.g. 2d6+1d10
type ThrowGroup struct {
	Counts map[FCount]int // map[face count]number of throws
}

// FCount represents a die with a number of faces, e.g. d6 or d10, notably d% for percentile dice
type FCount struct {
	FaceCount  int
	Percentile bool
	Fudge      bool
	Constant   bool
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
	tab := tabwriter.NewWriter(os.Stdout, 0, 0, 0, ' ', tabwriter.AlignRight)

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
			handledNegative := false

			if i > 0 {
				if index.Number >= 0 {
					_, _ = tab.Write([]byte(" + \t"))
				} else {
					_, _ = tab.Write([]byte(" - \t"))
					handledNegative = true
				}
			}

			// take special care of percentile dice's 0 state
			if !index.Fudge {
				total += index.Number
			}

			numDisplay := strconv.Itoa(index.Number)
			if numDisplay == "0" {
				numDisplay = "00"
			}

			faceDisplay := strconv.Itoa(index.FaceCount)
			if index.Percentile {
				faceDisplay = "%"
			}

			if index.Constant {
				if handledNegative {
					fmt.Fprintf(tab, "\t%d\t", int(math.Abs(float64(index.Number))))
				} else {
					fmt.Fprintf(tab, "\t%d\t", index.Number)
				}

				continue
			}

			if index.Fudge {
				faceDisplay = "F"
				switch index.Number {
				case 1, 2:
					numDisplay = "-1"
					total--
				case 3, 4:
					numDisplay = "0"
				default:
					numDisplay = "1"
					total++
				}
			}

			_, _ = fmt.Fprintf(tab, "d%s( \t%s )\t", faceDisplay, numDisplay)
		}

		if len(throws) > 1 || len(outcomes) > 1 {
			// account for missing columns (each missing column is 3x tabs: after the
			// plus, before the roll value, and after the roll value)
			tabSkips := (colCount - len(outcomes)) * 3

			_, _ = tab.Write(bytes.Repeat([]byte("\t"), tabSkips))

			fmt.Fprintf(tab, " = \t%d\t", total)
		}

		bigTotal += total

		// prepare for next row of output
		_, _ = tab.Write([]byte("\n"))
	}

	// if more than 1 group input, print TOTAL total
	if len(throws) > 1 {
		tabSkips := (colCount)*3 - 1
		fmt.Fprint(tab, strings.Repeat("\t", tabSkips))

		fmt.Fprintf(tab, "= \t%d\t\n", bigTotal)
	}

	// flush the output or you don't see it
	err := tab.Flush()
	if err != nil {
		panic(err)
	}
}

func parseArgs(args []string) (groups []ThrowGroup, columnCount int) {
argLoop:
	for _, arg := range args {
		tg := ThrowGroup{
			Counts: map[FCount]int{},
		}

		grouped := splitOnOperators(arg)

		cols := 0
	groupLoop:
		for _, g := range grouped {
			if g == "" {
				continue groupLoop
			}
			// split a dice count from it's face count
			parts := strings.Split(strings.ToUpper(g), "D")

			// assume 1 die if no count is given
			if parts[0] == "" {
				parts[0] = "1"
			}

			num, err := strconv.Atoi(parts[0])
			if err != nil {
				fmt.Printf("invalid dice entry: %s\n", arg)
				continue argLoop
			}

			if len(parts) == 1 {
				// Constant value
				fc := FCount{Constant: true}

				_, exists := tg.Counts[fc]
				if exists {
					tg.Counts[fc] += num
				} else {
					tg.Counts[fc] = num
					cols++
				}

				continue groupLoop
			}

			if num < 1 {
				fmt.Printf("invalid dice entry: %s\n", arg)

				continue groupLoop
			}

			var faces int
			var percentile, fudge bool

			if parts[1] == "%" {
				// handle percentile special
				faces = 10
				percentile = true
			} else if parts[1] == "F" {
				// fudge dice
				faces = 6
				fudge = true
			} else {
				faces, err = strconv.Atoi(parts[1])
				if err != nil || faces < 1 {
					fmt.Printf("invalid dice entry: %s\n", arg)
					continue argLoop
				}
			}

			// accumulate total columns for this group
			cols += num

			fc := FCount{
				FaceCount:  faces,
				Percentile: percentile,
				Fudge:      fudge,
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

func splitOnOperators(arg string) []string {
	grouped := []string{}
	start := 0

	for i, r := range arg {
		if r == '+' {
			grouped = append(grouped, arg[start:i])
			start = i + 1
		}
		if r == '-' {
			grouped = append(grouped, arg[start:i])
			start = i
		}
	}

	grouped = append(grouped, arg[start:])

	return grouped
}

func sortThrows(outcomes []Throw) {
	sort.Slice(outcomes, func(i, j int) bool {
		// Move constants to the end
		if outcomes[i].Constant && !outcomes[j].Constant {
			return false
		}
		if !outcomes[i].Constant && outcomes[j].Constant {
			return true
		}

		// If both are constants, sort by FaceCount (constant value)
		if outcomes[i].Constant && outcomes[j].Constant {
			return outcomes[i].FaceCount < outcomes[j].FaceCount
		}

		// Existing sorting logic for non-constants
		if outcomes[i].FaceCount == outcomes[j].FaceCount {
			if outcomes[i].Percentile != outcomes[j].Percentile {
				return outcomes[i].Percentile
			}

			if outcomes[i].Fudge != outcomes[j].Fudge {
				return outcomes[j].Fudge
			}

			return outcomes[i].Number > outcomes[j].Number
		}

		return outcomes[i].FaceCount > outcomes[j].FaceCount
	})
}

func (tg ThrowGroup) Roll() []Throw {
	results := make([]Throw, 0, len(tg.Counts))

	for faceCount, number := range tg.Counts {
		if faceCount.Constant {
			results = append(results, Throw{
				FCount: faceCount,
				Number: number,
			})
			continue
		}
		for range number {
			// pull the best random number you can get
			n, err := rand.Int(randSource, big.NewInt(int64(faceCount.FaceCount)))
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
