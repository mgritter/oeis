package combinations

import ()

// This structure represents a subset of 0...N-1
type IndicatorConfig struct {
	Size   int
	Offset int
}

type IndicatorMap struct {
	Present []bool
	Values  []int
	Config  IndicatorConfig
}

func NewIndicatorMap(c IndicatorConfig) IndicatorMap {
	return IndicatorMap{
		Present: make([]bool, c.Size),
		Values:  make([]int, c.Size),
		Config:  c,
	}
}

func (m *IndicatorMap) Set(i int, v int) {
	m.Present[i+m.Config.Offset] = true
	m.Values[i+m.Config.Offset] = v
}

func (m *IndicatorMap) SetRaw(i int, v int) {
	m.Present[i] = true
	m.Values[i] = v
}

func (m *IndicatorMap) Get(i int) int {
	return m.Values[i+m.Config.Offset]
}

func copyMap(m IndicatorMap) IndicatorMap {
	c := NewIndicatorMap(m.Config)
	for i := range c.Values {
		c.Present[i] = m.Present[i]
		if c.Present[i] {
			c.Values[i] = m.Values[i]
		}
	}
	return c
}

// Enumerate all nonempty subsets of elemeents, to a channel.
// The subsets are represented as indicator maps, using the zero
// and one inputs (to allow reversal of color, if desired.)
func EnumerateNonemptySubsets(config IndicatorConfig, elements []int, zero int, one int, out chan<- IndicatorMap) {
	var recurse func([]int, bool)
	chosen := NewIndicatorMap(config)

	recurse = func(remaining []int, onePresent bool) {
		first := remaining[0]
		// Base case: if "one" has already been used, we may instead pick zero,
		// otherwise this is out last chance to include it.
		if len(remaining) == 1 {
			if onePresent {
				chosen.Set(first, zero)
				out <- copyMap(chosen)
			}
			chosen.Set(first, one)
			out <- copyMap(chosen)
			return
		}

		// Recursive case: try with, and without the first element.
		chosen.Set(first, zero)
		recurse(remaining[1:], onePresent)

		chosen.Set(first, one)
		recurse(remaining[1:], true)
	}

	recurse(elements, false)
	close(out)
}

type SetGenerator interface {
	Config() IndicatorConfig
	Enumerate(out chan<- IndicatorMap)
}

type MandatoryOne struct {
	C   IndicatorConfig
	Set []int
}

func (m *MandatoryOne) Config() IndicatorConfig {
	return m.C
}

func (m *MandatoryOne) Enumerate(out chan<- IndicatorMap) {
	EnumerateNonemptySubsets(m.C, m.Set, 0, 1, out)
}

type MandatoryZero struct {
	C   IndicatorConfig
	Set []int
}

func (m *MandatoryZero) Enumerate(out chan<- IndicatorMap) {
	EnumerateNonemptySubsets(m.C, m.Set, 1, 0, out)
}

func (m *MandatoryZero) Config() IndicatorConfig {
	return m.C
}

type FreeChoice struct {
	C     IndicatorConfig
	Index int
}

func (f *FreeChoice) Enumerate(out chan<- IndicatorMap) {
	chosen := NewIndicatorMap(f.C)
	chosen.Set(f.Index, 0)
	out <- copyMap(chosen)
	chosen.Set(f.Index, 1)
	out <- copyMap(chosen)
	close(out)
}

func (f *FreeChoice) Config() IndicatorConfig {
	return f.C
}

func Product(sets []SetGenerator, out chan<- IndicatorMap) {
	chosen := NewIndicatorMap(sets[0].Config())
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

	config := sets[0].Config()

	recurse = func(i int) {
		ch := make(chan IndicatorMap)
		if sets[i].Config() != config {
			panic("mismatched set configs")
		}
		go sets[i].Enumerate(ch)

		for subMap := range ch {
			// Copy in just the values that were chosen.
			for i, p := range subMap.Present {
				if p {
					chosen.SetRaw(i, subMap.Values[i])
				}
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
