package equiv

import (
	"fmt"
	"sort"
	"strings"
)

// Equivalence classes of nxn grids based on their lower edge
//
//  x   x   x
//  0   1   2
//
// The canonical label is obtained by:
//   * sort the positions within each edge set
//   * sort the edge sets within each color by their minimum element
//   * take the minimum over
//         * interchange black and write
//         * flip along the vertical line x = (n-1)/2
//
// Solid boundaries are colored White.

type GridRectangle struct {
	Width      int
	Height     int
	SolidColor bool
	White      EdgePartition
	Black      EdgePartition
}

func (e *EdgePartition) MidpointFlip(width int) EdgePartition {
	var ret EdgePartition
	ret.Sets = make([]EdgeSet, len(e.Sets))

	for i := range e.Sets {
		n := len(e.Sets[i])
		ret.Sets[i] = make([]int, n)
		// example: [ 1 2 4 ] width 5
		// midpoint is 2
		// becomes [ 0 2 1 ]
		for j := range e.Sets[i] {
			ret.Sets[i][j] = (width - 1) - e.Sets[i][n-1-j]
		}
	}

	sort.Slice(ret.Sets, func(i, j int) bool {
		return ret.Sets[i][0] < ret.Sets[j][0]
	})
	return ret

}

// MakeCanonical normalizes a grid rectangle to its canonical value,
// from which we can derive a label.
func (g *GridRectangle) MakeCanonical() {
	g.White.Sort()
	g.Black.Sort()

	// Smallest edge is 0
	if len(g.Black.Sets) > 0 && g.Black.Sets[0][0] == 0 {
		// Swap so that White has the smallest edge
		g.White, g.Black = g.Black, g.White
	}

	extremal := g.Width - 1
	// If the largest edge is in Black, then compare color change + flip
	if g.Black.Contains(extremal) {
		alt := g.Black.MidpointFlip(g.Width)
		if alt.Compare(&g.White) == -1 {
			g.Black = g.White.MidpointFlip(g.Width)
			g.White = alt
		}
	} else {
		// otherwise compare only flip
		alt := g.White.MidpointFlip(g.Width)
		switch alt.Compare(&g.White) {
		case -1:
			g.Black = g.Black.MidpointFlip(g.Width)
			g.White = alt
		case 0:
			// If flip(W) = W, then we might
			// have to tie-break on flip(B)
			altBlack := g.Black.MidpointFlip(g.Width)
			if altBlack.Compare(&g.Black) == -1 {
				g.Black = altBlack
				g.White = alt // equal, should not matter
			}
		}
	}
}

// Key is unique to the width, but not the height,
// so we can re-use it
func (g *GridRectangle) Key() string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%d", g.Width))
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

func (g *GridRectangle) Plot() string {
	whiteLetters := "abcdefghijklmnopqrstuvwxyz"
	blackLetters := "ZYXWVUTSRQPONMLKJIHGFEDCBA"
	lower := make([]byte, g.Width)

	for i := range g.White.Sets {
		for _, pos := range g.White.Sets[i] {
			lower[pos] = whiteLetters[i]
		}
	}
	for i := range g.Black.Sets {
		for _, pos := range g.Black.Sets[i] {
			lower[pos] = blackLetters[i]
		}
	}
	return string(lower)
}
