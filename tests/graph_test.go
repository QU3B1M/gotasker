package tests

import (
	"gotasker/src/graph"
	"reflect"
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := graph.NewGraph()
	if g == nil {
		t.Error("NewGraph returned nil")
	}
}

func TestDependOn(t *testing.T) {
	g := graph.NewGraph()
	err := g.DependOn("a", "b")
	if err != nil {
		t.Error("DependOn returned an error")
	}
}

func TestDependsOn(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	if !g.DependsOn("a", "b") {
		t.Error("DependsOn returned false")
	}
}

func TestHasDependent(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	if !g.HasDependent("b", "a") {
		t.Error("HasDependent returned false")
	}
}

func TestDependOnSelf(t *testing.T) {
	g := graph.NewGraph()
	err := g.DependOn("a", "a")
	if err == nil {
		t.Error("DependOn did not return an error")
	}
}

func TestDependOnCircular(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	err := g.DependOn("c", "a")
	if err == nil {
		t.Error("DependOn did not return an error")
	}
}

func TestLeaves(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("c", "d")
	g.DependOn("d", "e")
	g.DependOn("e", "f")
	g.DependOn("f", "g")
	g.DependOn("g", "h")
	g.DependOn("h", "i")
	g.DependOn("i", "j")
	g.DependOn("j", "k")
	g.DependOn("k", "l")
	g.DependOn("l", "m")
	g.DependOn("m", "n")
	g.DependOn("n", "o")
	g.DependOn("o", "p")
	g.DependOn("p", "q")
	g.DependOn("q", "r")
	g.DependOn("r", "s")
	g.DependOn("s", "t")
	g.DependOn("t", "u")
	g.DependOn("u", "v")
	g.DependOn("v", "w")
	g.DependOn("w", "x")
	g.DependOn("x", "y")
	g.DependOn("y", "z")
	leaves := g.Leaves()
	if !reflect.DeepEqual(leaves, []string{"z"}) {
		t.Error("Leaves did not return the expected result")
	}
}

func TestTopSortedLayers(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("b", "d")
	g.DependOn("c", "e")
	g.DependOn("d", "e")
	layers := g.TopSortedLayers()
	if !reflect.DeepEqual(layers, [][]string{{"e"}, {"c", "d"}, {"b"}, {"a"}}) {
		t.Error("TopSortedLayers did not return the expected result")
	}
}

func TestTopSortedLayersCircular(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("c", "a") // Circular is ignored
	layers := g.TopSortedLayers()
	if !reflect.DeepEqual(layers, [][]string{{"c"}, {"b"}, {"a"}}) {
		t.Error("TopSortedLayers did not return the expected result")
	}
}

// Test TopSorted()

func TestTopSorted(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("b", "d")
	g.DependOn("c", "e")
	g.DependOn("d", "e")
	sorted := g.TopSorted()
	if !reflect.DeepEqual(sorted, []string{"e", "c", "d", "b", "a"}) {
		t.Error("TopSorted did not return the expected result")
	}
}

func TestTopSortedCircular(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("c", "a") // Circular is ignored
	sorted := g.TopSorted()
	if !reflect.DeepEqual(sorted, []string{"c", "b", "a"}) {
		t.Error("TopSorted did not return the expected result")
	}
}

func TestTopSortedMultiple(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("b", "d")
	g.DependOn("c", "e")
	g.DependOn("d", "e")
	sorted := g.TopSorted()
	if !reflect.DeepEqual(sorted, []string{"e", "c", "d", "b", "a"}) {
		t.Error("TopSorted did not return the expected result")
	}
}

func TestTopSortedMultipleCircular(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("c", "a") // Circular is ignored
	sorted := g.TopSorted()
	if !reflect.DeepEqual(sorted, []string{"c", "b", "a"}) {
		t.Error("TopSorted did not return the expected result")
	}
}
