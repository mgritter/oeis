package unionfind

// Identify a cell by its layer (size of the grid)
// and position (counting from lower-left to upper-right
// around the border.)
type Cell struct {
	Layer    int
	Position int
}

type UnionFind struct {
	parent map[Cell]Cell
	rank   map[Cell]int
}

func NewUnionFind() *UnionFind {
	return &UnionFind{
		parent: make(map[Cell]Cell),
		rank:   make(map[Cell]int),
	}
}

func (u *UnionFind) MakeSet(key Cell) {
	u.parent[key] = key
	u.rank[key] = 0
}

func (u *UnionFind) Find(key Cell) Cell {
	if u.parent[key] == key {
		return key
	}

	p := u.Find(u.parent[key])
	u.parent[key] = p
	return p
}

func (u *UnionFind) Union(a Cell, b Cell) {
	aRoot := u.Find(a)
	bRoot := u.Find(b)
	aRank := u.rank[aRoot]
	bRank := u.rank[bRoot]

	switch {
	case aRoot == bRoot:
		break
	case aRank > bRank:
		u.parent[bRoot] = aRoot
	case aRank < bRank:
		u.parent[aRoot] = bRoot
	case aRank == bRank:
		u.parent[bRoot] = aRoot
		u.rank[aRoot] = aRank + 1
	}
}
