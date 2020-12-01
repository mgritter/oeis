package unionfind

type UnionFind struct {
	parent map[string]string
	rank   map[string]int
}

func NewUnionFind() *UnionFind {
	return &UnionFind{
		parent: make(map[string]string),
		rank:   make(map[string]int),
	}
}

func (u *UnionFind) MakeSet(key string) {
	u.parent[key] = key
	u.rank[key] = 0
}

func (u *UnionFind) Find(key string) string {
	if u.parent[key] == key {
		return key
	}

	p := u.Find(u.parent[key])
	u.parent[key] = p
	return p
}

func (u *UnionFind) Union(a string, b string) {
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
