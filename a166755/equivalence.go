package main

import (
	"fmt"
	"sync"

	"github.com/mgritter/oeis/a166755/combinations"
	"github.com/mgritter/oeis/a166755/equiv"
)

type EquivalenceClasses struct {
	Size                 int
	Classes              map[string]*equiv.GridBoundary
	CountByClass         map[string]uint64
	CountByPartitionSize map[int]uint64
	CountValid           uint64
}

func NewEquivalenceClasses(size int) *EquivalenceClasses {
	return &EquivalenceClasses{
		Size:                 size,
		Classes:              make(map[string]*equiv.GridBoundary),
		CountByClass:         make(map[string]uint64),
		CountByPartitionSize: make(map[int]uint64),
		CountValid:           0,
	}
}

func InitEquivalenceClasses() *EquivalenceClasses {
	gb := equiv.GridBoundary{
		Size:       1,
		SolidColor: true,
		White: equiv.EdgePartition{[][]int{
			[]int{0},
		}},
		Black: equiv.EdgePartition{[][]int{}},
	}
	return &EquivalenceClasses{
		Size: 1,
		Classes: map[string]*equiv.GridBoundary{
			gb.Key(): &gb,
		},
		CountByClass: map[string]uint64{
			gb.Key(): 2,
		},
		CountByPartitionSize: map[int]uint64{
			1: 2,
		},
		CountValid: 2,
	}
}

type EquivalentGrids struct {
	Boundary *equiv.GridBoundary
	Count    uint64
}

func (e *EquivalenceClasses) AddGrids(g EquivalentGrids) {
	classKey := g.Boundary.Key()
	// fmt.Printf("Key %v count %v\n", classKey, g.Count)
	// fmt.Printf("%v\n", g.Boundary.Plot())
	e.Classes[classKey] = g.Boundary
	e.CountByClass[classKey] += g.Count

	partKey := len(g.Boundary.White.Sets) + len(g.Boundary.Black.Sets)
	e.CountByPartitionSize[partKey] += g.Count
	if partKey == 2 || (partKey == 1 && !g.Boundary.SolidColor) {
		e.CountValid += g.Count
	}
}

func EnumerateChildren(gb *equiv.GridBoundary, count uint64, out chan<- EquivalentGrids) {
	whiteExpansion := gb.White.Expand()
	blackExpansion := gb.Black.Expand()

	// For a 3x3 grid we expand to 4x4 which has
	// -3 -2 -1 0 1 2 3
	config := combinations.IndicatorConfig{
		Size:   2*gb.Size + 1,
		Offset: gb.Size,
	}

	// If the existing border is a single color, but the grid is
	// not monochromatic, then there is only one way to expand it.
	if len(gb.Black.Sets) == 0 && !gb.SolidColor {
		allZeros := make([]int, 2*gb.Size+1)
		out <- EquivalentGrids{
			gb.Expand(allZeros),
			count,
		}
		return
	}

	// If there is a single group of one color, we can create a solid
	// border of the other color.
	if len(gb.White.Sets) == 1 {
		allOnes := make([]int, 2*gb.Size+1)
		for i := range allOnes {
			allOnes[i] = 1
		}
		out <- EquivalentGrids{
			gb.Expand(allOnes),
			count,
		}
	}

	if len(gb.Black.Sets) == 1 {
		allZeros := make([]int, 2*gb.Size+1)
		out <- EquivalentGrids{
			gb.Expand(allZeros),
			count,
		}
	}

	// Otherwise, the new border should preserve any existing
	// single-colored region, to avoid leaving it disconnected.
	// At least one square should be extended to the new border,
	// for each set.
	gens := make([]combinations.SetGenerator, 0,
		len(whiteExpansion.Sets)+len(blackExpansion.Sets)+1)

	for _, s := range whiteExpansion.Sets {
		gens = append(gens, &combinations.MandatoryZero{config, s})
	}
	gens = append(gens, &combinations.FreeChoice{config, 0})
	for _, s := range blackExpansion.Sets {
		gens = append(gens, &combinations.MandatoryOne{config, s})
	}

	ch := make(chan combinations.IndicatorMap)
	go combinations.Product(gens, ch)

	for boundary := range ch {
		out <- EquivalentGrids{
			gb.Expand(boundary.Values),
			count,
		}
	}
}

func equivalenceClassEnumerator(workQueue <-chan EquivalentGrids, results chan<- EquivalentGrids) {
	for g := range workQueue {
		EnumerateChildren(g.Boundary, g.Count, results)
	}
}

func equivalenceClassAccumulator(size int, workQueue <-chan EquivalentGrids, results chan<- *EquivalenceClasses) {
	ec := NewEquivalenceClasses(size)
	for child := range workQueue {
		ec.AddGrids(child)
	}
	results <- ec
}

func equivalenceClassEnumeration(cases []int) {
	max := cases[len(cases)-1]
	size := 1
	ec := InitEquivalenceClasses()

	for size < max {
		size += 1
		prevEc := ec

		prevClasses := make(chan EquivalentGrids, 100)
		newClasses := make(chan EquivalentGrids, 100)
		newResult := make(chan *EquivalenceClasses)
		var wg sync.WaitGroup
		for i := 0; i < *NumWorkers; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				equivalenceClassEnumerator(prevClasses, newClasses)
			}()
		}

		go func() {
			wg.Wait()
			close(newClasses)
		}()

		go func() {
			for key, boundary := range prevEc.Classes {
				count := prevEc.CountByClass[key]
				prevClasses <- EquivalentGrids{boundary, count}
			}
			close(prevClasses)
		}()

		go func() {
			equivalenceClassAccumulator(size, newClasses, newResult)
			close(newResult)
		}()

		ec = <-newResult

		fmt.Printf("\n N=%d | grids=%v | classes=%v \n\n", size, ec.CountValid, len(ec.Classes))
	}
}
