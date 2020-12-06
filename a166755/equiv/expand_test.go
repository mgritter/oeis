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

	boundary := []int{0, 1, 0, 1, 0, 1, 0}
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
		boundary := make([]int, 9)
		for x := 0; x <= 4; x++ {
			boundary[x] = grid[4][x]
			// 0 1 2 3 4
		}
		for y := 0; y < 4; y++ {
			boundary[8-y] = grid[y][4]
			//         8
			//         7
			//         6
			//         5
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

func TestRectangle_5x3(t *testing.T) {
	// Take a random 5x5 grid
	// Compute the equivalence classes for the 5x4 subset
	// and the 5x5, then ensure the expansion matches.
	invariant := func(grid [][]int) bool {
		class4 := RectangleClassForGrid(5, 4, grid)
		class5 := RectangleClassForGrid(5, 5, grid)
		boundary := make([]int, 5)
		for x := 0; x <= 4; x++ {
			boundary[x] = grid[4][x]
		}
		t.Logf("Grid:\n%v", ShowGrid(5, grid))
		t.Logf("4x5 class:\n%v", class4.Plot())
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

func TestRectangle_ExpandSolid(t *testing.T) {
	firstRow := &GridRectangle{
		Width:      4,
		Height:     1,
		SolidColor: true,
		White:      EdgePartition{Sets: [][]int{[]int{0, 1, 2, 3}}},
		Black:      EdgePartition{Sets: make([][]int, 0)},
	}

	singleBoundaries := [][]int{
		[]int{0, 0, 0, 0},
		[]int{0, 0, 0, 1},
		[]int{0, 0, 1, 0},
		[]int{0, 0, 1, 1},
		[]int{0, 1, 0, 0},
		[]int{0, 1, 1, 0},
		[]int{0, 1, 1, 1},
		[]int{1, 0, 0, 0},
		[]int{1, 1, 0, 0},
		[]int{1, 1, 1, 0},
		[]int{1, 1, 1, 1},
	}
	for _, b := range singleBoundaries {
		expansion := firstRow.Expand(b)
		white := expansion.White.Sets
		black := expansion.Black.Sets
		if len(black)+len(white) > 2 {
			t.Fatalf("expansion %v created too many clases %v", b, expansion)
		}
	}

	doubleBoundaries := [][]int{
		[]int{0, 1, 0, 1},
		[]int{1, 0, 0, 1},
		[]int{1, 0, 1, 0},
		[]int{1, 0, 1, 1},
		[]int{1, 1, 0, 1},
	}
	for _, b := range doubleBoundaries {
		expansion := firstRow.Expand(b)
		white := expansion.White.Sets
		black := expansion.Black.Sets
		if len(black)+len(white) != 3 {
			t.Fatalf("expansion %v created too few clases %v", b, expansion)
		}
	}

}
