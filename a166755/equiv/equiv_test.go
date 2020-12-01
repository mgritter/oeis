package equiv

import (
	"reflect"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/mitchellh/copystructure"
)

func TestGrid_MakeCanonical(t *testing.T) {
	testCases := []struct {
		Input    GridBoundary
		Expected GridBoundary
	}{
		{
			// ..x
			// ..x
			// YYx
			GridBoundary{
				3,
				false,
				EdgePartition{ // white
					[][]int{
						[]int{1, 2, 0},
					},
				},
				EdgePartition{ // black
					[][]int{
						[]int{-1, -2},
					},
				},
			},
			GridBoundary{
				3,
				false,
				EdgePartition{ // white
					[][]int{
						[]int{-2, -1},
					},
				},
				EdgePartition{ // black
					[][]int{
						[]int{0, 1, 2},
					},
				},
			},
		},
		{
			// ..x
			// ..Y
			// xYZ
			GridBoundary{
				3,
				false,
				EdgePartition{ // white
					[][]int{
						[]int{2, -2},
						[]int{0},
					},
				},
				EdgePartition{ // black
					[][]int{
						[]int{-1, 1},
					},
				},
			},
			GridBoundary{
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
			},
		},
		{
			// ...c
			// ...c
			// ...D
			// aBaD
			GridBoundary{
				4,
				false,
				EdgePartition{ // white
					[][]int{
						[]int{-3, -1},
						[]int{2, 3},
					},
				},
				EdgePartition{ // black
					[][]int{
						[]int{-2},
						[]int{1, 0},
					},
				},
			},
			GridBoundary{
				4,
				false,
				EdgePartition{ // white
					[][]int{
						[]int{-3, -2},
						[]int{1, 3},
					},
				},
				EdgePartition{ // black
					[][]int{
						[]int{-1, 0},
						[]int{2},
					},
				},
			},
		},
		{
			// ...a
			// ...D
			// ...a
			// aBaC
			GridBoundary{
				4,
				false,
				EdgePartition{ // white
					[][]int{
						[]int{3, 1, -3, -1},
					},
				},
				EdgePartition{ // black
					[][]int{
						[]int{2},
						[]int{0},
						[]int{-2},
					},
				},
			},
			GridBoundary{
				4,
				false,
				EdgePartition{ // white
					[][]int{
						[]int{-3, -1, 1, 3},
					},
				},
				EdgePartition{ // black
					[][]int{
						[]int{-2},
						[]int{0},
						[]int{2},
					},
				},
			},
		},
	}

	for i, tc := range testCases {
		t.Logf("Test case %v", i)
		t.Logf("Original: %v", tc.Input.Key())
		tc.Input.MakeCanonical()
		t.Logf("Canonical: %v", tc.Input.Key())
		t.Logf("Plot:\n%v", tc.Input.Plot())
		t.Logf("Expected: %v", tc.Expected.Key())
		if !reflect.DeepEqual(tc.Input, tc.Expected) {
			t.Error("Not deep equal.")
		}
	}
}

func TestGrid_CanonicalIsInvariant(t *testing.T) {
	properties := gopter.NewProperties(nil)

	makeCopy := func(g *GridBoundary) *GridBoundary {
		t.Helper()
		raw, err := copystructure.Copy(g)
		if err != nil {
			t.Fatal(err)
		}
		return raw.(*GridBoundary)
	}

	n := 5
	invariant := func(membership []int) bool {
		// Interpret the 2n-1 integers as tags indicating membership
		// in a particular set for the edges. Some of the possible
		// partitions are impossible, but we should arrive at a correct
		// canonical representation anyway.
		//
		// Even numbers are assigned to white,
		// Odd numbers are assigned to black.

		members := make(map[int][]int)
		for i, v := range membership {
			pos := i - (n - 1)
			members[v] = append(members[v], pos)
		}
		white := make([][]int, 0)
		black := make([][]int, 0)
		for tag, edges := range members {
			if tag%2 == 0 {
				white = append(white, edges)
			} else {
				black = append(black, edges)
			}
		}

		orig := GridBoundary{
			n,
			false,
			EdgePartition{white},
			EdgePartition{black},
		}

		t.Logf("Testing %v", orig.Key())
		t.Logf("Plot\n%v", orig.Plot())
		unmodified := makeCopy(&orig)
		unmodified.MakeCanonical()

		colorSwap := makeCopy(&orig)
		colorSwap.White, colorSwap.Black = colorSwap.Black, colorSwap.White
		colorSwap.MakeCanonical()
		if !reflect.DeepEqual(unmodified, colorSwap) {
			t.Logf("color swap failed: %v != %v",
				unmodified.Key(), colorSwap.Key())
			return false
		}

		diagSwap := makeCopy(&orig)
		diagSwap.White.DiagonalFlip()
		diagSwap.Black.DiagonalFlip()
		diagSwap.MakeCanonical()
		if !reflect.DeepEqual(unmodified, diagSwap) {
			t.Logf("diagonal swap failed: %v != %v",
				unmodified.Key(), diagSwap.Key())
			return false
		}

		bothSwap := makeCopy(&orig)
		bothSwap.White, bothSwap.Black = colorSwap.Black.DiagonalFlip(), colorSwap.White.DiagonalFlip()
		bothSwap.MakeCanonical()
		if !reflect.DeepEqual(unmodified, bothSwap) {
			t.Logf("both swap failed: %v != %v",
				unmodified.Key(), bothSwap.Key())
			return false
		}

		for i := 0; i < len(orig.White.Sets)-1; i++ {
			reorder := makeCopy(&orig)
			reorder.White.Sets[i], reorder.White.Sets[i+1] =
				reorder.White.Sets[i+1], reorder.White.Sets[i]
			reorder.MakeCanonical()
			if !reflect.DeepEqual(unmodified, reorder) {
				t.Logf("set transposition failed: %v != %v",
					unmodified.Key(), reorder.Key())
				return false
			}
		}

		for i := 0; i < len(orig.White.Sets); i++ {
			reorder := makeCopy(&orig)
			if len(reorder.White.Sets[i]) >= 2 {
				s := reorder.White.Sets[i]
				s[0], s[1] = s[1], s[0]
				reorder.MakeCanonical()
				if !reflect.DeepEqual(unmodified, reorder) {
					t.Logf("set transposition failed: %v != %v",
						unmodified.Key(), reorder.Key())
					return false
				}
			}
		}
		return true

	}
	properties.Property("invariant under transformations",
		prop.ForAll(invariant,
			gen.SliceOfN(2*n-1, gen.IntRange(0, 2*n-1)),
		))
	properties.TestingRun(t)

}
