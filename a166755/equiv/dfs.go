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

func ConnectedComponentDFS(size int, colors map[Coord]int, start Coord, visited map[Coord]bool) []Coord {
	component := make([]Coord, 0)

	var dfs func(Coord)
	dfs = func(curr Coord) {
		if visited[curr] {
			return
		}
		component = append(component, curr)
		visited[curr] = true

		for _, n := range curr.Neighbors(size) {
			if colors[n] == colors[curr] {
				dfs(n)
			}
		}
	}

	dfs(start)
	return component
}

func GridToMap(n int, squares [][]int) map[Coord]int {
	colors := make(map[Coord]int, n*n)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			colors[Coord{x + 1, y + 1}] = squares[y][x]
		}
	}
	return colors
}

func EdgeClassForGrid(n int, squares [][]int) *GridBoundary {
	colors := GridToMap(n, squares)
	visited := make(map[Coord]bool)

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
			colors[Coord{x, n}],
		)
	}

	for y := 1; y < n; y++ {
		addComponent(
			ConnectedComponentDFS(n, colors, Coord{n, y}, visited),
			colors[Coord{n, y}],
		)
	}

	if numComponents == 1 {
		ret.SolidColor = true
	}

	return ret
}
