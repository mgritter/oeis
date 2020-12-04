package combinations

import ()

// FIXME: making a copy each time might be expensive,
// as is merging with other maps, so maybe some sort of
// custom data structure is in order.
type IndicatorMap = map[int]int

func copyMap(m IndicatorMap) IndicatorMap {
	c := make(map[int]int, len(m))
	for k, v := range m {
		c[k] = v
	}
	return c
}

// Enumerate all nonempty subsets of s, to a generator
// The subsets are represented as indicator maps, using the zero
// and one inputs (to allow reversal of color, if desired.)
func EnumerateNonemptySubsets(elements []int, zero int, one int, out chan<- IndicatorMap) {
	var recurse func([]int, bool)
	chosen := make(IndicatorMap, len(elements))

	recurse = func(remaining []int, onePresent bool) {
		first := remaining[0]
		// Base case: if "one" has already been used, we may instead pick zero,
		// otherwise this is out last chance to include it.
		if len(remaining) == 1 {
			if onePresent {
				chosen[first] = zero
				out <- copyMap(chosen)
			}
			chosen[first] = one
			out <- copyMap(chosen)
			return
		}

		// Recursive case: try with, and without the first element.
		chosen[first] = zero
		recurse(remaining[1:], onePresent)

		chosen[first] = one
		recurse(remaining[1:], true)
	}

	recurse(elements, false)
	close(out)
}

type SetGenerator interface {
	Enumerate(out chan<- IndicatorMap)
}

type MandatoryOne struct {
	Set []int
}

func (m *MandatoryOne) Enumerate(out chan<- IndicatorMap) {
	EnumerateNonemptySubsets(m.Set, 0, 1, out)
}

type MandatoryZero struct {
	Set []int
}

func (m *MandatoryZero) Enumerate(out chan<- IndicatorMap) {
	EnumerateNonemptySubsets(m.Set, 1, 0, out)
}

type FreeChoice struct {
	Index int
}

func (f *FreeChoice) Enumerate(out chan<- IndicatorMap) {
	chosen := make(IndicatorMap)
	chosen[f.Index] = 0
	out <- copyMap(chosen)
	chosen[f.Index] = 1
	out <- copyMap(chosen)
	close(out)
}

func Product(sets []SetGenerator, out chan<- IndicatorMap) {
	chosen := make(IndicatorMap)
	ProductWithPrefix(chosen, sets, out)
	close(out)
}

func ProductList(sets []SetGenerator) []IndicatorMap {
	ch := make(chan IndicatorMap)
	result := make([]IndicatorMap, 0)
	go Product(sets, ch)
	for m := range ch {
		result = append(result, m)
	}
	return result
}

func ProductWithPrefix(chosen IndicatorMap, sets []SetGenerator, out chan<- IndicatorMap) {
	var recurse func(int)
	recurse = func(i int) {
		ch := make(chan IndicatorMap)
		go sets[i].Enumerate(ch)

		for subMap := range ch {
			for k, v := range subMap {
				chosen[k] = v
			}
			if i == 0 {
				out <- copyMap(chosen)
			} else {
				recurse(i - 1)
			}
		}
	}

	if len(sets) > 0 {
		recurse(len(sets) - 1)
	}
}
