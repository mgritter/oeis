package combinations

import (
	"testing"
)

func TestCombinations_Example(t *testing.T) {
	out := make(chan IndicatorMap)
	go EnumerateNonemptySubsets(
		IndicatorConfig{4, 0},
		[]int{0, 1, 2, 3},
		0, 1,
		out)

	count := 0
	for x := range out {
		count += 1
		t.Logf("output %d: %v", count, x)
		if len(x.Values) != 4 {
			t.Fatal("map too small")
		}
		foundNonzero := false
		for _, v := range x.Values {
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
		x.Set(1, 2)
	}
	if count != 15 {
		t.Fatalf("expected 15 subsets, got %d", count)
	}
}

func TestProduct_Example(t *testing.T) {
	config := IndicatorConfig{7, 3}
	a := &MandatoryOne{config, []int{-3, -2, -1}}
	b := &FreeChoice{config, 0}
	c := &MandatoryZero{config, []int{1, 2, 3}}

	out := make(chan IndicatorMap)
	go Product([]SetGenerator{a, b, c}, out)

	count := 0
	for x := range out {
		count += 1
		t.Logf("output %d: %v", count, x)

		if len(x.Values) != 7 {
			t.Fatal("map too small")
		}

		if x.Get(1) != 0 && x.Get(2) != 0 && x.Get(3) != 0 {
			t.Fatal("no zeros for indices 1,2,3")
		}
		if x.Get(-1) != 1 && x.Get(-2) != 1 && x.Get(-3) != 1 {
			t.Fatal("no ones for indices -3,-2,-1")
		}
		for _, p := range x.Present {
			if !p {
				t.Fatal("value not present")
			}
		}
	}

	if count != 7*2*7 {
		t.Fatalf("bad count, got %v", count)
	}
}
