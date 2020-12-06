package unionfind

import ()

// Union-find optimized for two-row case
type RowUnionFind struct {
	width  int
	parent []int
	rank   []int
}

func NewRowUnionFind(width int) *RowUnionFind {
	ret := &RowUnionFind{
		width:  width,
		parent: make([]int, 2*width),
		rank:   make([]int, 2*width),
	}
	for i := range ret.parent {
		ret.parent[i] = i
		ret.rank[i] = 0
	}
	return ret
}

func (u *RowUnionFind) FindCell(y int, x int) int {
	return u.Find(y*u.width + x)
}

func (u *RowUnionFind) Find(key int) int {
	if u.parent[key] == key {
		return key
	}

	p := u.Find(u.parent[key])
	u.parent[key] = p
	return p
}

func (u *RowUnionFind) UnionCell(y1, x1, y2, x2 int) {
	u.Union(y1*u.width+x1,
		y2*u.width+x2)
}

func (u *RowUnionFind) Union(a int, b int) {
	// fmt.Printf("Union %d %d\n", a, b)
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
