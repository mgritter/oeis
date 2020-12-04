package equiv

// Equivalence classes of nxn grids based on their boundary.
//
//  x   x   2
//  x   x   1
// -2  -1   0
//
// We create one set for the elements of the boundary that
// are connected by a path, and a list of such sets for the two
// colors white (0) and black (1).
//
// The canonical label is obtained by:
//   * sort the positions within each edge set
//   * sort the edge sets within each color by their minimum element
//   * take the minimum over
//         * interchange black and write (i.e., partition sets)
//         * flip along the diagonal axes (multiply all numbers by -1)
//
// For example, if we have (-3 -1 0) | (-2) (1 2 3)
//    flip color: (-2) (1 2 3)   | (-3 -1 0)
//    flip diag:  (0 1 3)        | (-3 -2 -1) (2)
//    flip both:  (-3 -2 -1) (2) | (0 1 3)
// The last option is the lexicographically smallest vector of integers.
// We might have to compare several edge sets, although only the ones
// starting with -N are the possible minimums.
// for example if we had (-3 2 3) as a set we would only flip diagnonally,
// not by color.
//
// Solid boundaries are colored White, which is a little inconsistent but
// more intuitive.
//
// To use the canonical label as a key, we need to make a string out of it.
//
// We also need to track for the signature (-2, -1, 0, 1, 2) whether
// it is all-white (and could be expanded with some black squares) or
// two-color and thus can only be expanded with white squares.

import (
	"fmt"
	"sort"
	"strings"
)

type EdgeSet = []int

type EdgePartition struct {
	Sets []EdgeSet
}

func (e *EdgePartition) Sort() {
	// Sort each set in numerical order
	for i := range e.Sets {
		sort.Ints(e.Sets[i])
	}
	// Sort by first (least) element
	sort.Slice(e.Sets, func(i, j int) bool {
		return e.Sets[i][0] < e.Sets[j][0]
	})
}

// DiagonalFlip returns an EdgePartition derived by flippping this
// partition along the diagonal.
func (e *EdgePartition) DiagonalFlip() EdgePartition {
	var ret EdgePartition
	ret.Sets = make([]EdgeSet, len(e.Sets))
	// Negate each element of the set, which exactly reverses the order
	for i := range e.Sets {
		n := len(e.Sets[i])
		ret.Sets[i] = make([]int, n)
		for j := range e.Sets[i] {
			ret.Sets[i][j] = -(e.Sets[i][n-1-j])
		}
	}
	// Sort the remaining elements, no good way to predict this?
	sort.Slice(ret.Sets, func(i, j int) bool {
		return ret.Sets[i][0] < ret.Sets[j][0]
	})
	return ret
}

func SetLexicographicCompare(a []int, b []int) int {
	for i := range a {
		if i >= len(b) {
			return 1 // a > b, b is a prefix of a
		}
		if a[i] < b[i] {
			return -1 // a < b
		}
		if a[i] > b[i] { // a > b
			return 1
		}
	}
	if len(a) == len(b) {
		return 0
	}
	return -1 // a < b, a is a prefix of b
}

func (e *EdgePartition) Compare(f *EdgePartition) int {
	for i := range e.Sets {
		if i >= len(f.Sets) {
			// e is a suffix of f
			return 1
		}
		switch SetLexicographicCompare(e.Sets[i], f.Sets[i]) {
		case -1:
			return -1
		case 1:
			return 1
		default:
		}
	}
	// All equal (this should be unlikely?), e is first if
	// if it a prefix of f.
	if len(e.Sets) < len(f.Sets) {
		return -1
	}
	return 0
}

// Contains checkes whether any set contains the given edge.
func (e *EdgePartition) Contains(x int) bool {
	for i := range e.Sets {
		for _, v := range e.Sets[i] {
			if v == x {
				return true
			}
		}
	}
	return false
}

type GridBoundary struct {
	Size       int
	SolidColor bool
	White      EdgePartition
	Black      EdgePartition
}

// MakeCanonical normalizes a grid boundary to its canonical value,
// form which we can derive a label.
func (g *GridBoundary) MakeCanonical() {
	g.White.Sort()
	g.Black.Sort()

	extremal := g.Size - 1
	if len(g.Black.Sets) > 0 && g.Black.Sets[0][0] == -extremal {
		// Swap so that White has the smallest edge
		tmp := g.White
		g.White = g.Black
		g.Black = tmp
	}

	// If the largest edge is in Black, then compare color change + diagonal
	if g.Black.Contains(extremal) {
		alt := g.Black.DiagonalFlip()
		if alt.Compare(&g.White) == -1 {
			g.Black = g.White.DiagonalFlip()
			g.White = alt
		}
	} else {
		// otherwise compare only diagonal flip
		alt := g.White.DiagonalFlip()
		switch alt.Compare(&g.White) {
		case -1:
			g.Black = g.Black.DiagonalFlip()
			g.White = alt
		case 0:
			// If color-swap(W) = W, then we might
			// have to tie-break on color-swap(B)
			altBlack := g.Black.DiagonalFlip()
			if altBlack.Compare(&g.Black) == -1 {
				g.Black = altBlack
				g.White = alt // equal, should not matter
			}
		}
	}
}

// Key creates a string unique to this GridBoundary, so it can be
// put in a map.
func (g *GridBoundary) Key() string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%d", g.Size))
	if len(g.Black.Sets) == 0 {
		if g.SolidColor {
			buf.WriteString(":solid")
		} else {
			buf.WriteString(":bicolor")
		}
		return buf.String()
	}

	for i := range g.White.Sets {
		buf.WriteString(":W")
		s := g.White.Sets[i]
		for j := range s {
			buf.WriteString(fmt.Sprintf(",%d", s[j]))
		}
	}

	for i := range g.Black.Sets {
		buf.WriteString(":B")
		s := g.Black.Sets[i]
		for j := range s {
			buf.WriteString(fmt.Sprintf(",%d", s[j]))
		}
	}
	return buf.String()
}

// Plot shows a graphical representation of the congruence class.
func (g *GridBoundary) Plot() string {
	whiteLetters := "abcdefghijklmnopqrstuvwxyz"
	blackLetters := "ZYXWVUTSRQPONMLKJIHGFEDCBA"
	right := make([]byte, g.Size)
	lower := make([]byte, g.Size)
	var zero byte

	for i := range g.White.Sets {
		for _, pos := range g.White.Sets[i] {
			switch {
			case pos == 0:
				zero = whiteLetters[i]
			case pos < 0:
				lower[-pos] = whiteLetters[i]
			case pos > 0:
				right[pos] = whiteLetters[i]
			}
		}
	}
	for i := range g.Black.Sets {
		for _, pos := range g.Black.Sets[i] {
			switch {
			case pos == 0:
				zero = blackLetters[i]
			case pos < 0:
				lower[-pos] = blackLetters[i]
			case pos > 0:
				right[pos] = blackLetters[i]
			}
		}
	}

	var buf strings.Builder
	for y := 0; y < g.Size-1; y++ {
		for x := 0; x < g.Size-1; x++ {
			buf.WriteRune('.')
		}
		buf.WriteByte(right[g.Size-1-y])
		buf.WriteRune('\n')
	}

	for x := 0; x < g.Size-1; x++ {
		buf.WriteByte(lower[g.Size-1-x])
	}
	buf.WriteByte(zero)
	buf.WriteRune('\n')
	return buf.String()
}
