package main

import (
	"fmt"
	//"sync"

	"github.com/mgritter/oeis/a166755/combinations"
	"github.com/mgritter/oeis/a166755/equiv"
)

var Verbose = false

type EdgeClass struct {
	Key   string
	Class *equiv.GridRectangle
	Count uint64
}

type SuccessorMap struct {
	Width int

	// Map from key to all the successors of the edge class
	SuccessorCounts map[string][]EdgeClass

	// Newly introduced classes for which no successor is known
	NewClasses []*equiv.GridRectangle

	// For the current height, which classes are present
	// and how many members do they have?
	CountByClass map[string]EdgeClass
}

func EnumerateRectangleChildren(gr *equiv.GridRectangle) []EdgeClass {
	config := combinations.IndicatorConfig{
		Size:   gr.Width,
		Offset: 0,
	}
	white := gr.White.Sets
	black := gr.Black.Sets

	byKey := make(map[string]EdgeClass)
	add := func(gr *equiv.GridRectangle) {
		key := gr.Key()
		if exist, ok := byKey[key]; ok {
			byKey[key] = EdgeClass{
				Key:   exist.Key,
				Class: exist.Class,
				Count: exist.Count + 1,
			}
		} else {
			byKey[key] = EdgeClass{key, gr, 1}
		}
	}

	// If the existing border is a single color, but the grid is
	// not monochromatic, then there is only one way to expand it.
	if len(black) == 0 && !gr.SolidColor {
		allZeros := make([]int, gr.Width)
		expansion := gr.Expand(allZeros)
		return []EdgeClass{EdgeClass{expansion.Key(), expansion, 1}}
	}

	// If there is a single group of one color, we can create a solid
	// border of the other color.
	if len(white) == 1 {
		allOnes := make([]int, gr.Width)
		for i := range allOnes {
			allOnes[i] = 1
		}
		add(gr.Expand(allOnes))
	}

	if len(black) == 1 {
		allZeros := make([]int, gr.Width)
		add(gr.Expand(allZeros))
	}

	// Otherwise, the new border should preserve any existing
	// single-colored region, to avoid leaving it disconnected.
	// At least one square should be extended to the new border,
	// for each set.
	gens := make([]combinations.SetGenerator, 0, len(white)+len(black))

	for _, s := range white {
		gens = append(gens, &combinations.MandatoryZero{config, s})
	}
	for _, s := range black {
		gens = append(gens, &combinations.MandatoryOne{config, s})
	}

	ch := make(chan combinations.IndicatorMap)
	go combinations.Product(gens, ch)

	for boundary := range ch {
		add(gr.Expand(boundary.Values))
	}

	result := make([]EdgeClass, 0, len(byKey))
	for _, v := range byKey {
		result = append(result, v)
	}
	return result
}

func (s *SuccessorMap) Iterate(height int) {
	moreNewClasses := make(map[string]*equiv.GridRectangle)

	for _, c := range s.NewClasses {
		// Placeholder so that we don't trigger NEW again
		s.SuccessorCounts[c.Key()] = []EdgeClass{}
	}

	for _, c := range s.NewClasses {
		cKey := c.Key()
		if Verbose {
			fmt.Printf("Expanding %v %v\n", c.Plot(), cKey)
		}
		expansions := EnumerateRectangleChildren(c)
		s.SuccessorCounts[cKey] = expansions

		for _, e := range expansions {
			if _, ok := s.SuccessorCounts[e.Key]; !ok {
				if e.Class.Height != height {
					panic("new class lacks correct height")
				}
				moreNewClasses[e.Key] = e.Class
				if Verbose {
					fmt.Printf(" %v %d %v NEW\n", e.Class.Plot(), e.Count, e.Key)
				}
			} else {
				if Verbose {
					fmt.Printf(" %v %d %v\n", e.Class.Plot(), e.Count, e.Key)
				}
			}
		}
	}

	s.NewClasses = make([]*equiv.GridRectangle, 0, len(moreNewClasses))
	for _, v := range moreNewClasses {
		s.NewClasses = append(s.NewClasses, v)
	}

	newCounts := make(map[string]EdgeClass)
	for k, startClass := range s.CountByClass {
		for _, successor := range s.SuccessorCounts[k] {
			k2 := successor.Key
			if exist, ok := newCounts[k2]; ok {
				newCounts[k2] = EdgeClass{
					Class: exist.Class,
					Count: exist.Count + startClass.Count*successor.Count,
				}
			} else {
				newCounts[k2] = EdgeClass{
					Class: successor.Class,
					Count: startClass.Count * successor.Count,
				}
			}
		}
	}

	s.CountByClass = newCounts
	if Verbose {
		fmt.Printf("\nCounts at height %d:\n", height)
		for key, val := range newCounts {
			fmt.Printf(" %v %d %v\n", val.Class.Plot(), val.Count, key)
		}
		fmt.Printf("\n")
	}
}

func startingClasses(width int) map[string]EdgeClass {
	// Only
	//  wwwwwwwww / bbbbbbbbb
	// or
	//  wwwwwbbbb
	// or
	//  wbbbbbwww
	// can be be a valid first row. If there are more than three
	// groups, then not all of them can be connected later (without
	// violating planarity!)

	byKey := make(map[string]EdgeClass)
	add := func(gr *equiv.GridRectangle) {
		key := gr.Key()
		if exist, ok := byKey[key]; ok {
			byKey[key] = EdgeClass{
				Key:   key,
				Class: exist.Class,
				Count: exist.Count + 1,
			}
		} else {
			byKey[key] = EdgeClass{key, gr, 1}
		}
	}

	// Counting separately to make it easier to avoid double-counting.
	// First case; solid color
	set1 := make([]int, 0, width)
	for i := 0; i < width; i++ {
		set1 = append(set1, i)
	}
	a := &equiv.GridRectangle{
		Width:      width,
		Height:     1,
		SolidColor: true,
		White:      equiv.EdgePartition{[][]int{set1}},
		Black:      equiv.EdgePartition{[][]int{}},
	}
	a.MakeCanonical()
	byKey[a.Key()] = EdgeClass{a.Key(), a, 2}

	// Second case: two colors
	// Some of these map to the same class, i.e., aab and bba
	for right := 1; right < width; right++ {
		set1 := make([]int, right)
		set2 := make([]int, width-right)
		for i := 0; i < right; i++ {
			set1[i] = i
		}
		for i := right; i < width; i++ {
			set2[i-right] = i
		}

		a := &equiv.GridRectangle{
			Width:      width,
			Height:     1,
			SolidColor: false,
			White:      equiv.EdgePartition{[][]int{set1}},
			Black:      equiv.EdgePartition{[][]int{set2}},
		}
		a.MakeCanonical()
		add(a)

		b := &equiv.GridRectangle{
			Width:      width,
			Height:     1,
			SolidColor: false,
			White:      equiv.EdgePartition{[][]int{set2}},
			Black:      equiv.EdgePartition{[][]int{set1}},
		}
		b.MakeCanonical()
		add(b)
	}

	// Third case: white-black-white or black-white-black
	// these cannot alias under color inversion, but might under mirror image.
	for left := 1; left < width-1; left++ {
		for right := left + 1; right < width; right++ {
			set1 := make([]int, left)
			set2 := make([]int, right-left)
			set3 := make([]int, width-right)
			for i := 0; i < left; i++ {
				set1[i] = i
			}
			for i := left; i < right; i++ {
				set2[i-left] = i
			}
			for i := right; i < width; i++ {
				set3[i-right] = i
			}
			a := &equiv.GridRectangle{
				Width:      width,
				Height:     1,
				SolidColor: false,
				White:      equiv.EdgePartition{[][]int{set1, set3}},
				Black:      equiv.EdgePartition{[][]int{set2}},
			}
			a.MakeCanonical()
			add(a)

			b := &equiv.GridRectangle{
				Width:      width,
				Height:     1,
				SolidColor: false,
				White:      equiv.EdgePartition{[][]int{set2}},
				Black:      equiv.EdgePartition{[][]int{set1, set3}},
			}
			b.MakeCanonical()
			add(b)
		}
	}

	if Verbose {
		fmt.Printf("Initial row:")
		for key, val := range byKey {
			fmt.Printf(" %v %d %v\n", val.Class.Plot(), val.Count, key)
		}
	}

	return byKey
}

func (s *SuccessorMap) ValidCount() uint64 {
	var total uint64 = 0

	for k, v := range s.CountByClass {
		numPartitions := len(v.Class.White.Sets) + len(v.Class.Black.Sets)
		if numPartitions == 2 || (numPartitions == 1 && !v.Class.SolidColor) {
			if Verbose {
				fmt.Printf("Valid: %d %v\n", v.Count, k)
			}
			total += v.Count
		}
	}
	return total
}

func rectangleEnumeration(cases []int) {
	for _, width := range cases {

		firstRow := startingClasses(width)
		s := &SuccessorMap{
			Width:           width,
			SuccessorCounts: make(map[string][]EdgeClass),
			NewClasses:      make([]*equiv.GridRectangle, 0, len(firstRow)),
			CountByClass:    firstRow,
		}

		for _, v := range firstRow {
			s.NewClasses = append(s.NewClasses, v.Class)
		}
		for height := 2; height <= width; height++ {
			s.Iterate(height)
		}

		count := s.ValidCount()
		fmt.Printf("**** N=%v | grids=%v | classes = %v \n\n", width, count, len(s.CountByClass))
	}
}
