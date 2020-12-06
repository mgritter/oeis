package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"sort"
	"strconv"

	"net/http"
	_ "net/http/pprof"
)

var NumWorkers = flag.Int("numworkers", 8, "number of worker goroutines")
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var RunExhaustive = flag.Bool("exhaustive", false, "use exhaustive enumeration")
var RunSquare = flag.Bool("square", false, "use expanding squres")
var Verbose = flag.Bool("verbose", false, "verbose output")

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	go http.ListenAndServe("localhost:5432", nil)

	cases := []int{2, 3, 4, 5, 6, 7, 8, 9, 10}
	if len(flag.Args()) > 0 {
		cases = []int{}
		for _, txt := range flag.Args() {
			n, err := strconv.ParseInt(txt, 10, 32)
			if err != nil {
				fmt.Printf("couldn't parse argument %v: %v\n", txt, err)
				return
			}
			cases = append(cases, int(n))
		}
	}

	if *RunExhaustive {
		exhaustiveEnumeration(cases)
		return
	}

	if *RunSquare {
		// This is a bit silly, we have to generate all smaller cases anyway.
		sort.Ints(cases)
		equivalenceClassEnumeration(cases)
		return
	}

	rectangleEnumeration(cases)
}
