package unionfind

import (
	"testing"
	/*
		"github.com/leanovate/gopter"
		"github.com/leanovate/gopter/gen"
		"github.com/leanovate/gopter/prop"
	*/)

func TestUnionFind_Example(t *testing.T) {
	labels := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "o", "u"}
	uf := NewUnionFind()
	for i := range labels {
		uf.MakeSet(labels[i])
	}

	uf.Union("a", "e")
	uf.Union("i", "o")
	uf.Union("o", "u")
	uf.Union("a", "u")

	uf.Union("b", "c")
	uf.Union("d", "f")
	uf.Union("f", "g")
	uf.Union("c", "h")
	uf.Union("j", "b")
	uf.Union("g", "j")

	vowelRoot := uf.Find("a")
	consonantRoot := uf.Find("b")
	for _, vowel := range []string{"a", "e", "i", "o", "u"} {
		if uf.Find(vowel) != vowelRoot {
			t.Fatalf("%q not in set %q", vowel, vowelRoot)
		}
	}
	for _, consonant := range []string{"b", "c", "d", "f", "g", "h", "j"} {
		if uf.Find(consonant) != consonantRoot {
			t.Fatalf("%q not in set %q", consonant, consonantRoot)
		}
	}
}
