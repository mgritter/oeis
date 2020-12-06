package equiv

// A DFS implementation for testing/brute forcing

type Coord struct {
	X int
	Y int
}

func (c Coord) ToCell() Cell {
	switch {
	case c.X == c.Y:
		return Cell{c.X, 0}
	case c.X > c.Y:
		return Cell{c.X, c.X - c.Y}
	case c.Y > c.X:
		// Negative values
		return Cell{c.Y, c.X - c.Y}
	}
	panic("bad coord")
}

func (curr Coord) Neighbors(size int) []Coord {
	nn := make([]Coord, 0)
	if curr.X > 1 {
		nn = append(nn, Coord{curr.X - 1, curr.Y})
	}
	if curr.X < size {
		nn = append(nn, Coord{curr.X + 1, curr.Y})
	}
	if curr.Y > 1 {
		nn = append(nn, Coord{curr.X, curr.Y - 1})
	}
	if curr.Y < size {
		nn = append(nn, Coord{curr.X, curr.Y + 1})
	}
	return nn
}

func (c Coord) Index(size int) int {
	return (c.Y-1)*size + (c.X - 1)
}

func ConnectedComponentDFS(size int, colors []int, start Coord, visited []bool) []Coord {
	component := make([]Coord, 0)

	var dfs func(Coord)
	dfs = func(curr Coord) {
		if visited[curr.Index(size)] {
			return
		}
		component = append(component, curr)
		visited[curr.Index(size)] = true

		for _, n := range curr.Neighbors(size) {
			if colors[n.Index(size)] == colors[curr.Index(size)] {
				dfs(n)
			}
		}
	}

	dfs(start)
	return component
}

func FlattenGrid(n int, squares [][]int) []int {
	colors := make([]int, n*n)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			colors[y*n+x] = squares[y][x]
		}
	}
	return colors
}

func EdgeClassForGrid(n int, squares [][]int) *GridBoundary {
	colors := FlattenGrid(n, squares)
	visited := make([]bool, n*n)

	ret := &GridBoundary{
		Size:       n,
		SolidColor: false,
		White:      EdgePartition{Sets: make([][]int, 0)},
		Black:      EdgePartition{Sets: make([][]int, 0)},
	}

	numComponents := 0

	addComponent := func(component []Coord, color int) {
		if len(component) == 0 {
			return
		}

		numComponents += 1

		edges := make([]int, 0)
		for _, c := range component {
			cell := c.ToCell()
			if cell.Layer == n {
				edges = append(edges, cell.Position)
			}
		}

		if color == 0 {
			ret.White.Sets = append(ret.White.Sets, edges)
		} else {
			ret.Black.Sets = append(ret.Black.Sets, edges)
		}
	}

	for x := 1; x <= n; x++ {
		addComponent(
			ConnectedComponentDFS(n, colors, Coord{x, n}, visited),
			colors[Coord{x, n}.Index(n)],
		)
	}

	for y := 1; y < n; y++ {
		addComponent(
			ConnectedComponentDFS(n, colors, Coord{n, y}, visited),
			colors[Coord{n, y}.Index(n)],
		)
	}

	if numComponents == 1 {
		ret.SolidColor = true
		for _, v := range colors {
			if v != colors[0] {
				ret.SolidColor = false
				break
			}
		}
	}

	return ret
}

func RectangleClassForGrid(width int, height int, squares [][]int) *GridRectangle {
	// To re-use the DFS code that assumes a square grid we will
	// pad the squares out.
	colors := make([]int, width*width)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			colors[y*width+x] = squares[y][x]
		}
	}
	for y := height; y < width; y++ {
		for x := 0; x < width; x++ {
			colors[y*width+x] = -1
		}
	}
	visited := make([]bool, width*width)

	ret := &GridRectangle{
		Width:      width,
		Height:     height,
		SolidColor: false,
		White:      EdgePartition{Sets: make([][]int, 0)},
		Black:      EdgePartition{Sets: make([][]int, 0)},
	}

	numComponents := 0

	addComponent := func(component []Coord, color int) {
		if len(component) == 0 {
			return
		}

		numComponents += 1

		edges := make([]int, 0)
		for _, c := range component {
			if c.Y == height {
				edges = append(edges, c.X-1)
			}
		}

		if color == 0 {
			ret.White.Sets = append(ret.White.Sets, edges)
		} else {
			ret.Black.Sets = append(ret.Black.Sets, edges)
		}
	}

	for x := 1; x <= width; x++ {
		addComponent(
			ConnectedComponentDFS(width, colors, Coord{x, height}, visited),
			squares[height-1][x-1],
		)
	}

	if numComponents == 1 {
		ret.SolidColor = true
		for _, v := range colors[:width*height] {
			if v != colors[0] {
				ret.SolidColor = false
				break
			}
		}
	}

	return ret
}
