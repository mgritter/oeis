package main

import (
	"fmt"
	"sync"

	"github.com/mgritter/oeis/a166755/combinations"
	"github.com/mgritter/oeis/a166755/equiv"
)

func hasTwoRegions(n int, grid map[int]int) bool {
	colors := make(map[equiv.Coord]int)

	for y := 1; y <= n; y++ {
		for x := 1; x <= n; x++ {
			colors[equiv.Coord{x, y}] = grid[(y-1)*n+x]
		}
	}
	// fmt.Printf("Grid: %v\n", colors)

	visited := make(map[equiv.Coord]bool)

	numComponents := 0
	for y := 1; y <= n; y++ {
		for x := 1; x <= n; x++ {
			component := equiv.ConnectedComponentDFS(n, colors, equiv.Coord{x, y}, visited)
			if len(component) > 0 {
				numComponents += 1
				if numComponents > 2 {
					return false
				}
			}
		}
	}
	return numComponents == 2
}

type Count struct {
	Valid    int
	NotValid int
}

func exhaustiveWorker(n int, inputs <-chan combinations.IndicatorMap, result chan<- Count) {
	valid := 0
	notValid := 0
	for grid := range inputs {
		if hasTwoRegions(n, grid) {
			valid += 1
		} else {
			notValid += 1
		}
	}
	result <- Count{valid, notValid}
}

func exhaustiveCount(n int) Count {
	cells := make([]combinations.SetGenerator, n*n)
	for i := 0; i < n*n; i++ {
		cells[i] = &combinations.FreeChoice{i + 1}
	}

	allGrids := make(chan combinations.IndicatorMap, *NumWorkers*2)

	// Divide up the grid by cell
	numGenerators := 2
	prefixLength := 1
	for numGenerators*2 <= *NumWorkers {
		numGenerators *= 2
		prefixLength += 1
	}
	var gg sync.WaitGroup
	prefixes := combinations.ProductList(cells[:prefixLength])
	if len(prefixes) != numGenerators {
		panic("prefixes are not a power of 2.")
	}
	for i := 0; i < numGenerators; i++ {
		gg.Add(1)
		go func(me int) {
			defer gg.Done()
			combinations.ProductWithPrefix(prefixes[me], cells[prefixLength:], allGrids)
		}(i)
	}

	go func() {
		gg.Wait()
		close(allGrids)
	}()

	var wg sync.WaitGroup
	results := make(chan Count, *NumWorkers)
	for i := 0; i < *NumWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			exhaustiveWorker(n, allGrids, results)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var total Count
	for c := range results {
		fmt.Printf("Worker: %v\n", c)
		total.Valid += c.Valid
		total.NotValid += c.NotValid
	}
	return total
}

func exhaustiveEnumeration(cases []int) {
	for _, n := range cases {
		total := exhaustiveCount(n)
		fmt.Printf("%d | %d | %d\n", n, total.Valid, total.NotValid)
	}
}
