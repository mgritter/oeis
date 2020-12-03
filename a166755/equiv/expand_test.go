package equiv

import (
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"reflect"
	"strings"
	"testing"
)

func TestGrid_ExpandExample(t *testing.T) {
	// ...
	// .XX
	// .X.
	before := GridBoundary{
		3,
		false,
		EdgePartition{ // white
			[][]int{
				[]int{-2, 2},
				[]int{0},
			},
		},
		EdgePartition{ // black
			[][]int{
				[]int{-1, 1},
			},
		},
	}

	t.Logf("before plot:\n%v", before.Plot())

	// ....
	// .XXX
	// .X..
	// .X.X

	boundary := map[int]int{
		-3: 0,
		-2: 1,
		-1: 0,
		0:  1,
		1:  0,
		2:  1,
		3:  0,
	}

	after := GridBoundary{
		4,
		false,
		EdgePartition{ // white
			[][]int{
				[]int{-3, 3},
				[]int{-1, 1},
			},
		},
		EdgePartition{ // black
			[][]int{
				[]int{-2, 2},
				[]int{0},
			},
		},
	}

	t.Logf("after plot:\n%v", after.Plot())

	actual := before.Expand(boundary)

	t.Logf("Actual: %v", actual.Key())
	t.Logf("Plot:\n%v", actual.Plot())

	if !reflect.DeepEqual(*actual, after) {
		t.Error("expansion does not match expected value.")
	}
}

func ShowGrid(n int, grid [][]int) string {
	var buf strings.Builder
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			switch grid[y][x] {
			case 0:
				buf.WriteString(".")
			case 1:
				buf.WriteString("X")
			default:
				buf.WriteString("?")
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

func TestGrid_5x5(t *testing.T) {
	// Take a random 5x5 grid
	// Compute the equivalence classes for the 4x4 subset
	// and the 5x5, then ensure the expansion matches.
	invariant := func(grid [][]int) bool {
		class4 := EdgeClassForGrid(4, grid)
		class5 := EdgeClassForGrid(5, grid)
		boundary := make(map[int]int)
		for x := 0; x <= 4; x++ {
			boundary[x-4] = grid[4][x]
		}
		for y := 0; y < 4; y++ {
			boundary[4-y] = grid[y][4]
		}
		t.Logf("Grid:\n%v", ShowGrid(5, grid))
		t.Logf("4x4 class:\n%v", class4.Plot())
		t.Logf("5x5 class:\n%v", class5.Plot())
		t.Logf("Boundary: %v", boundary)

		actual := class4.Expand(boundary)
		t.Logf("Actual: %v\n%v", actual, actual.Plot())
		class5.MakeCanonical()
		if !reflect.DeepEqual(actual, class5) {
			t.Logf("Expected: %v\n%v", class5, class5.Plot())
			return false
		}
		return true
	}

	properties := gopter.NewProperties(nil)
	properties.Property("expansion matches DFS",
		prop.ForAll(invariant,
			gen.SliceOfN(5, gen.SliceOfN(5, gen.IntRange(0, 1))),
		))
	properties.TestingRun(t)

}
