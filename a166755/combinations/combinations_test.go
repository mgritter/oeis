package combinations

import (
	"testing"
)

func TestCombinations_Example(t *testing.T) {
	out := make(chan IndicatorMap)
	go EnumerateNonemptySubsets([]int{1, 2, 3, 4}, 0, 1, out)

	count := 0
	for x := range out {
		count += 1
		t.Logf("output %d: %v", count, x)
		if len(x) != 4 {
			t.Fatal("map too small")
		}
		foundNonzero := false
		for _, v := range x {
			switch v {
			case 1:
				foundNonzero = true
			case 0:
			default:
				t.Fatal("non-0/1 element in map")
			}
		}
		if !foundNonzero {
			t.Fatal("no nonzero element in map")
		}
		// This will break the test if the map is not a copy
		x[1] = 2
	}
	if count != 15 {
		t.Fatalf("expected 15 subsets, got %d", count)
	}
}

func TestProduct_Example(t *testing.T) {
	a := &MandatoryOne{[]int{-3, -2, -1}}
	b := &FreeChoice{0}
	c := &MandatoryZero{[]int{1, 2, 3}}

	out := make(chan IndicatorMap)
	go Product([]SetGenerator{a, b, c}, out)

	count := 0
	for x := range out {
		count += 1
		t.Logf("output %d: %v", count, x)

		if len(x) != 7 {
			t.Fatal("map too small")
		}

		if x[1] != 0 && x[2] != 0 && x[3] != 0 {
			t.Fatal("no zeros for indices 1, 2, 3")
		}
		if x[-1] != 1 && x[-2] != 1 && x[-3] != 1 {
			t.Fatal("no ones for indices -1, -2, -3")
		}
	}

	if count != 7*2*7 {
		t.Fatalf("bad count, got %v", count)
	}
}
