package unionfind

import (
	"math"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

func TestUnionFind_Example(t *testing.T) {
	//   AB
	//   BB
	// ABCC
	// AAAC
	uf := NewUnionFind()
	for i := -2; i <= 2; i++ {
		uf.MakeSet(Cell{3, i})
	}
	for i := -3; i <= 3; i++ {
		uf.MakeSet(Cell{4, i})
	}

	// A
	uf.Union(Cell{3, -2}, Cell{3, 2})
	uf.Union(Cell{4, -3}, Cell{4, -2})
	uf.Union(Cell{4, -1}, Cell{4, -2})
	uf.Union(Cell{4, -3}, Cell{3, -2})

	// B
	uf.Union(Cell{3, -1}, Cell{3, 1})
	uf.Union(Cell{3, 1}, Cell{4, 2})
	uf.Union(Cell{4, 3}, Cell{4, 2})

	// C
	uf.Union(Cell{4, 0}, Cell{4, 1})
	uf.Union(Cell{3, 0}, Cell{4, 1})

	checkCells := func(set []Cell) {
		t.Helper()
		origin := uf.Find(set[0])
		for i := 1; i < len(set); i++ {
			if uf.Find(set[i]) != origin {
				t.Errorf("%v not in %v", set[i], origin)
			}
		}
	}

	checkCells([]Cell{
		Cell{3, 2},
		Cell{3, -2},
		Cell{4, -3},
		Cell{4, -2},
		Cell{4, -1},
	})

	checkCells([]Cell{
		Cell{4, 0},
		Cell{3, 0},
		Cell{4, 1},
	})

	checkCells([]Cell{
		Cell{3, -1},
		Cell{3, 1},
		Cell{4, 2},
		Cell{4, 3},
	})
}

func TestUnionFind_DoubleRing(t *testing.T) {
	// Break the elements into two pieces
	// Shuffle each half and join A to A+1 in each forming
	// two cycles.
	// Then bridge the two cycles to form the final set.
	// Do this twice to get two independent sets.

	n := 25

	nextCell := func(c Cell) Cell {
		return Cell{
			Layer:    c.Layer,
			Position: (c.Position + 1) % n,
		}
	}

	// shuffle should be 4n+4 numbers
	invariant := func(shuffle []int) bool {
		a := make([]Cell, n)
		b := make([]Cell, n)
		c := make([]Cell, n)
		d := make([]Cell, n)
		uf := NewUnionFind()
		for i := range a {
			a[i] = Cell{1, i}
			b[i] = Cell{2, i}
			c[i] = Cell{3, i}
			d[i] = Cell{4, i}
			uf.MakeSet(a[i])
			uf.MakeSet(b[i])
			uf.MakeSet(c[i])
			uf.MakeSet(d[i])
		}

		for i := 0; i < n; i++ {
			remaining := n - i
			a_choice := shuffle[i*4+0] % remaining
			b_choice := shuffle[i*4+1] % remaining
			c_choice := shuffle[i*4+0] % remaining
			d_choice := shuffle[i*4+0] % remaining

			// Move the one we've selected to the end of the list
			uf.Union(a[a_choice], nextCell(a[a_choice]))
			a[remaining-1], a[a_choice] = a[a_choice], a[remaining-1]

			uf.Union(b[b_choice], nextCell(b[b_choice]))
			b[remaining-1], b[b_choice] = b[b_choice], b[remaining-1]

			uf.Union(c[c_choice], nextCell(c[c_choice]))
			c[remaining-1], c[c_choice] = c[c_choice], c[remaining-1]

			uf.Union(d[d_choice], nextCell(d[d_choice]))
			d[remaining-1], d[d_choice] = d[d_choice], d[remaining-1]
		}

		uf.Union(a[shuffle[n*4+1]%n], b[shuffle[n*4+1]%n])
		uf.Union(c[shuffle[n*4+1]%n], d[shuffle[n*4+1]%n])

		aRoot := uf.Find(a[0])
		for _, cell := range a {
			if uf.Find(cell) != aRoot {
				t.Logf("%v is not in %v", cell, aRoot)
				return false
			}
		}
		for _, cell := range b {
			if uf.Find(cell) != aRoot {
				t.Logf("%v is not in %v", cell, aRoot)
				return false
			}
		}

		cRoot := uf.Find(c[0])
		for _, cell := range c {
			if uf.Find(cell) != cRoot {
				t.Logf("%v is not in %v", cell, aRoot)
				return false
			}
		}
		for _, cell := range d {
			if uf.Find(cell) != cRoot {
				t.Logf("%v is not in %v", cell, aRoot)
				return false
			}
		}

		if aRoot == cRoot {
			t.Logf("roots are the same: %v", aRoot)
			return false
		}
		return true
	}

	properties := gopter.NewProperties(nil)
	properties.Property("shuffle order",
		prop.ForAll(invariant,
			gen.SliceOfN(4*n+4, gen.IntRange(0, math.MaxInt32)),
		))
	properties.TestingRun(t)
}
