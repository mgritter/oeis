package main

import (
	"fmt"
	"math/big"
	"sync"

	"github.com/mgritter/oeis/a166755/combinations"
	"github.com/mgritter/oeis/a166755/equiv"
)

type EdgeClass struct {
	Key   string
	Class *equiv.GridRectangle
	Count *big.Int
}

func NewEdgeClass1(key string, c *equiv.GridRectangle) EdgeClass {
	return EdgeClass{
		Key:   key,
		Class: c,
		Count: big.NewInt(1),
	}
}

func NewEdgeClassProduct(key string, c *equiv.GridRectangle, n1 *big.Int, n2 *big.Int) EdgeClass {
	z := big.NewInt(0)
	z.Mul(n1, n2)
	return EdgeClass{
		Key:   key,
		Class: c,
		Count: z,
	}
}

func (e EdgeClass) Inc1() EdgeClass {
	z := big.NewInt(1)
	z.Add(z, e.Count)
	return EdgeClass{
		Key:   e.Key,
		Class: e.Class,
		Count: z,
	}
}

func (e EdgeClass) IncProduct(n1 *big.Int, n2 *big.Int) EdgeClass {
	z := big.NewInt(0)
	z.Mul(n1, n2)
	z.Add(z, e.Count)
	return EdgeClass{
		Key:   e.Key,
		Class: e.Class,
		Count: z,
	}
}

type SuccessorMap struct {
	Width int

	// Map from key to all the successors of the edge class
	//SuccessorCounts map[string][]EdgeClass
	SuccessorCounts sync.Map

	// Newly introduced classes for which no successor is known
	NewClasses []*equiv.GridRectangle

	// next version of NewClasses
	NextClasses sync.Map

	// For the current height, which classes are present
	// and how many members do they have?
	CountByClass map[string]EdgeClass

	// Valid classes only
	ValidClasses sync.Map
}

func (s *SuccessorMap) CheckValid(key string, gr *equiv.GridRectangle) {
	numPartitions := len(gr.White.Sets) + len(gr.Black.Sets)
	if numPartitions == 2 || (numPartitions == 1 && !gr.SolidColor) {
		s.ValidClasses.Store(key, struct{}{})
	}
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
			byKey[key] = exist.Inc1()
		} else {
			byKey[key] = NewEdgeClass1(key, gr)
		}
	}

	// If the existing border is a single color, but the grid is
	// not monochromatic, then there is only one way to expand it.
	if len(black) == 0 && !gr.SolidColor {
		allZeros := make([]int, gr.Width)
		expansion := gr.Expand(allZeros)
		return []EdgeClass{NewEdgeClass1(expansion.Key(), expansion)}
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

// We could accumulate all the results (successor counts and new functions) and have
// the originator put them all in the map, but I think the map is good enough for the
// scale we're working at.
func (s *SuccessorMap) Worker(height int, workQueue <-chan *equiv.GridRectangle) {
	for c := range workQueue {
		cKey := c.Key()
		if *Verbose {
			fmt.Printf("Expanding %v %v\n", c.Plot(), cKey)
		}
		expansions := EnumerateRectangleChildren(c)

		for i, e := range expansions {
			if _, ok := s.SuccessorCounts.Load(e.Key); !ok {
				if e.Class.Height != height {
					panic("new class lacks correct height")
				}
				s.NextClasses.Store(e.Key, e.Class)
				s.CheckValid(e.Key, e.Class)

				if *Verbose {
					fmt.Printf(" %v %d %v NEW\n", e.Class.Plot(), e.Count, e.Key)
				}
			} else {
				if *Verbose {
					fmt.Printf(" %v %d %v\n", e.Class.Plot(), e.Count, e.Key)
				}
			}
			// Throw away the class itself, so that we're normalized on
			// what remains in NextClasses
			expansions[i].Class = nil
		}

		// Store the low-cost list
		s.SuccessorCounts.Store(c.Key(), expansions)

	}
}

func (s *SuccessorMap) Iterate(height int) {
	for _, c := range s.NewClasses {
		// Placeholder so that we don't trigger NEW again
		s.SuccessorCounts.Store(c.Key(), []EdgeClass{})
	}

	workQueue := make(chan *equiv.GridRectangle, 100)
	var wg sync.WaitGroup

	for i := 0; i < *NumWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Worker(height, workQueue)
		}()
	}

	for _, c := range s.NewClasses {
		workQueue <- c
	}
	close(workQueue)
	wg.Wait()

	s.NewClasses = make([]*equiv.GridRectangle, 0)
	s.NextClasses.Range(func(k, v interface{}) bool {
		s.NewClasses = append(s.NewClasses, v.(*equiv.GridRectangle))
		s.NextClasses.Delete(k)
		return true
	})

	newCounts := make(map[string]EdgeClass)
	for k, startClass := range s.CountByClass {
		successorsRaw, found := s.SuccessorCounts.Load(k)
		if !found {
			panic("Sucessors not found.")
		}
		successors := successorsRaw.([]EdgeClass)
		for _, successor := range successors {
			k2 := successor.Key
			if exist, ok := newCounts[k2]; ok {
				newCounts[k2] = exist.IncProduct(startClass.Count, successor.Count)
			} else {
				newCounts[k2] = NewEdgeClassProduct(k2, successor.Class, startClass.Count, successor.Count)
			}
		}
	}

	numClasses := 0
	s.SuccessorCounts.Range(func(k, v interface{}) bool {
		numClasses += 1
		return true
	})
	fmt.Printf(" Height=%d classes=%d+%d\n", height, numClasses, len(s.NewClasses))

	s.CountByClass = newCounts
	if *Verbose {
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
			byKey[key] = exist.Inc1()
		} else {
			byKey[key] = NewEdgeClass1(key, gr)
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
	byKey[a.Key()] = EdgeClass{a.Key(), a, big.NewInt(2)}

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

	if *Verbose {
		fmt.Printf("Initial row:\n")
		for key, val := range byKey {
			fmt.Printf(" %v %d %v\n", val.Class.Plot(), val.Count, key)
		}
	}

	return byKey
}

func (s *SuccessorMap) ValidCount() *big.Int {
	total := big.NewInt(0)

	for k, v := range s.CountByClass {
		if _, present := s.ValidClasses.Load(k); present {
			total.Add(total, v.Count)
		}
	}
	return total
}

func rectangleEnumeration(cases []int) {
	for _, width := range cases {

		firstRow := startingClasses(width)

		s := &SuccessorMap{
			Width:        width,
			NewClasses:   make([]*equiv.GridRectangle, 0, len(firstRow)),
			CountByClass: firstRow,
		}

		for _, v := range firstRow {
			s.NewClasses = append(s.NewClasses, v.Class)
			s.CheckValid(v.Key, v.Class)
		}
		for height := 2; height <= width; height++ {
			s.Iterate(height)
		}

		count := s.ValidCount()
		fmt.Printf("**** N=%v | grids=%v | classes = %v \n\n", width, count, len(s.CountByClass))
	}
}
