package tests

import (
	"gotasker/src/graph"
	"reflect"
	"sort"
	"testing"
)

func TestNewGraph(t *testing.T) {
	g := graph.NewGraph()
	if g == nil {
		t.Error("NewGraph returned", g, "expected a graph")
	}
}

func TestDependOn(t *testing.T) {
	g := graph.NewGraph()
	err := g.DependOn("a", "b")
	if err != nil {
		t.Error("DependOn returned", err, "expected 'nil'")
	}
}

func TestDependsOn(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	expected := true
	result := g.DependsOn("a", "b")
	if result != expected {
		t.Error("DependsOn returned ", result, " expected ", expected)
	}
}

func TestHasDependent(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	expected := true
	result := g.HasDependent("b", "a")
	if result != expected {
		t.Error("HasDependent returned ", result, " expected ", expected)
	}
}

func TestDependOnSelf(t *testing.T) {
	g := graph.NewGraph()
	err := g.DependOn("a", "a")
	if err == nil {
		t.Error("DependOn returned", err, "expected 'error'")
	}
}

func TestDependOnCircular(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	err := g.DependOn("c", "a")
	if err == nil {
		t.Error("DependOn returned", err, "expected 'error'")
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
	expected := []string{"z"}
	leaves := g.Leaves()
	if !reflect.DeepEqual(leaves, expected) {
		t.Error("Leaves returned", leaves, "expected", expected)
	}
}

func TestTopSortedLayers(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("b", "d")
	g.DependOn("c", "e")
	g.DependOn("d", "e")

	// Sort the layers before comparing
	expected := [][]string{{"e"}, {"c", "d"}, {"b"}, {"a"}}
	for _, layer := range expected {
		sort.Strings(layer)
	}

	layers := g.TopSortedLayers()
	for _, layer := range layers {
		sort.Strings(layer)
	}

	if !reflect.DeepEqual(layers, expected) {
		t.Error("TopSortedLayers returned", layers, "expected", expected)
	}
}

func TestTopSortedLayersCircular(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("c", "a") // Circular is ignored
	expected := [][]string{{"c"}, {"b"}, {"a"}}
	layers := g.TopSortedLayers()

	if !reflect.DeepEqual(layers, expected) {
		t.Error("TopSortedLayers returned", layers, "expected", expected)
	}
}

// Test TopSorted()

func TestTopSorted(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("c", "d")
	expected := []string{"d", "c", "b", "a"}
	sorted := g.TopSorted()

	if !reflect.DeepEqual(sorted, expected) {
		t.Error("TopSorted returned", sorted, "expected", expected)
	}
}

func TestTopSortedCircular(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("c", "a") // Circular is ignored
	expected := []string{"c", "b", "a"}
	sorted := g.TopSorted()

	if !reflect.DeepEqual(sorted, expected) {
		t.Error("TopSorted returned", sorted, "expected", expected)
	}
}

func TestTopSortedMultiple(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("c", "d")
	g.DependOn("d", "e")
	expected := []string{"e", "d", "c", "b", "a"}
	sorted := g.TopSorted()
	if !reflect.DeepEqual(sorted, expected) {
		t.Error("TopSorted returned", sorted, "expected", expected)
	}
}

func TestTopSortedMultipleCircular(t *testing.T) {
	g := graph.NewGraph()
	g.DependOn("a", "b")
	g.DependOn("b", "c")
	g.DependOn("c", "a") // Circular is ignored
	expected := []string{"c", "b", "a"}
	sorted := g.TopSorted()
	if !reflect.DeepEqual(sorted, expected) {
		t.Error("TopSorted returned", sorted, "expected", expected)
	}
}
