package equiv

import (
	"github.com/mgritter/oeis/a166755/unionfind"
)

// ExpandSet returns the squares adjacent to the given edge set,
// in the next-larger grid.
//
//        2 -> 3
//        1 -> 2
// -2 -1  0 -> 1
//  |  |  |
//  V  V  V
// -3 -2 -1    0 (no image)
func ExpandSet(s []int) []int {
	// FIMXE: always make room, or expand only when 0 is part of the set?
	ret := make([]int, 0, len(s)+1)
	for _, pos := range s {
		switch {
		case pos == 0:
			ret = append(ret, -1)
			ret = append(ret, 1)
		case pos < 0:
			ret = append(ret, pos-1)
		case pos > 0:
			ret = append(ret, pos+1)
		}
	}
	return ret
}

// Expand translates each set into the adjacent set in the
// next-larger grid edges, using ExpandSet.
func (e *EdgePartition) Expand() EdgePartition {
	ret := EdgePartition{
		Sets: make([][]int, len(e.Sets)),
	}
	for i := range ret.Sets {
		ret.Sets[i] = ExpandSet(e.Sets[i])
	}
	return ret
}

type Cell = unionfind.Cell

// Expand the border and return the normalized GridBoundary.
// the argument is a map of position -> 0 for white, 1 for black
func (g *GridBoundary) Expand(newBorder []int) *GridBoundary {
	if len(newBorder) != 2*g.Size+1 {
		panic("incomplete border")
	}

	// Convert position to slice offset
	// New border runs from -g.Size to +g.Size
	getBorder := func(i int) int {
		return newBorder[i+g.Size]
	}

	// use union-find to track partitions
	uf := unionfind.NewUnionFind()

	// FIXME: keep this in the GridBoundary object?
	// FIXME: convert to slice?
	colorMap := make(map[int]int)

	for i := -(g.Size - 1); i <= g.Size-1; i++ {
		uf.MakeSet(Cell{g.Size, i})
	}

	for i := -(g.Size); i <= g.Size; i++ {
		uf.MakeSet(Cell{g.Size + 1, i})
	}

	// Previous white cells
	for _, s := range g.White.Sets {
		firstElement := Cell{g.Size, s[0]}
		colorMap[s[0]] = 0

		for i := 1; i < len(s); i++ {
			uf.Union(firstElement, Cell{g.Size, s[i]})
			colorMap[s[i]] = 0
		}
	}

	// Previous black cells
	for _, s := range g.Black.Sets {
		firstElement := Cell{g.Size, s[0]}
		colorMap[s[0]] = 1
		for i := 1; i < len(s); i++ {
			uf.Union(firstElement, Cell{g.Size, s[i]})
			colorMap[s[i]] = 1
		}
	}

	// New bottom cells
	for i := -g.Size; i < 0; i++ {
		// Same color as cell above?
		if getBorder(i) == colorMap[i+1] {
			uf.Union(Cell{g.Size + 1, i},
				Cell{g.Size, i + 1})
		}
		// Same color as cell to the right?
		// (Including corner.)
		if getBorder(i) == getBorder(i+1) {
			uf.Union(Cell{g.Size + 1, i},
				Cell{g.Size + 1, i + 1})
		}
	}

	// New right cells
	for i := 1; i <= g.Size; i++ {
		// Same color as cell to the left?
		if getBorder(i) == colorMap[i-1] {
			uf.Union(Cell{g.Size + 1, i},
				Cell{g.Size, i - 1})
		}
		// Same color as cell below?
		// (Including corner.)
		if getBorder(i) == getBorder(i-1) {
			uf.Union(Cell{g.Size + 1, i},
				Cell{g.Size + 1, i - 1})
		}
	}

	whiteMap := make(map[Cell][]int)
	blackMap := make(map[Cell][]int)

	for i := -g.Size; i <= g.Size; i++ {
		r := uf.Find(Cell{g.Size + 1, i})
		if getBorder(i) == 0 {
			whiteMap[r] = append(whiteMap[r], i)
		} else {
			blackMap[r] = append(blackMap[r], i)
		}
	}

	white := make([]EdgeSet, 0, len(whiteMap))
	black := make([]EdgeSet, 0, len(blackMap))

	for _, s := range whiteMap {
		white = append(white, s)
	}

	for _, s := range blackMap {
		black = append(black, s)
	}

	// If there's a solid border it should not be partitioned into
	// more than one set.
	if len(white) == 0 && len(black) != 1 {
		panic("bad solid border partition")
	}
	if len(black) == 0 && len(white) != 1 {
		panic("bad solid border partition")
	}

	ret := &GridBoundary{
		Size:       g.Size + 1,
		SolidColor: false,
		White:      EdgePartition{Sets: white},
		Black:      EdgePartition{Sets: black},
	}

	if g.SolidColor && len(black) == 0 {
		ret.SolidColor = true
	}

	ret.MakeCanonical()
	return ret
}
