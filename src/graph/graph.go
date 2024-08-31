// Package graph provides a simple directed acyclic graph (DAG) implementation.
// Using the code from github.com/kendru/darwin/blob/main/go/depgraph/ as base
package graph

import (
	"errors"
)

// Nodeset is a map of nodes, in this graph a node is just a string,
// keys are the parent nodes.
type Nodeset map[string]struct{}

// depmap tracks the nodes that have some dependency relationship to
// some other node, represented by the key of the map.
type depmap map[string]Nodeset

// DependencyGraph represents a directed graph with dependencies and dependents.
type DependencyGraph struct {
	nodes        Nodeset
	dependencies depmap // `dependencies` tracks child -> parents.
	dependents   depmap // `dependents` tracks parent -> children.
}

// NewGraph creates a new DependencyGraph instance.
func NewGraph() *DependencyGraph {
	return &DependencyGraph{
		dependencies: make(depmap),
		dependents:   make(depmap),
		nodes:        make(Nodeset),
	}
}

// DependOn adds a dependency relationship where 'child' depends on 'parent'.
// Returns an error if the relationship is self-referential or creates a circular dependency.
func (g *DependencyGraph) DependOn(child, parent string) error {
	if child == parent {
		return errors.New("self-referential dependencies not allowed")
	}

	if g.DependsOn(parent, child) {
		return errors.New("circular dependencies not allowed")
	}

	// Add nodes.
	g.nodes[parent] = struct{}{}
	g.nodes[child] = struct{}{}

	// Add edges.
	addNodeToNodeset(g.dependents, parent, child)
	addNodeToNodeset(g.dependencies, child, parent)

	return nil
}

// DependsOn checks if 'child' depends on 'parent'.
func (g *DependencyGraph) DependsOn(child, parent string) bool {
	deps := g.Dependencies(child)
	_, ok := deps[parent]
	return ok
}

// HasDependent checks if 'parent' has 'child' as a dependent.
func (g *DependencyGraph) HasDependent(parent, child string) bool {
	deps := g.Dependents(parent)
	_, ok := deps[child]
	return ok
}

// Leaves returns a list of nodes that have no dependencies.
func (g *DependencyGraph) Leaves() []string {
	leaves := make([]string, 0)

	for node := range g.nodes {
		if _, ok := g.dependencies[node]; !ok {
			leaves = append(leaves, node)
		}
	}

	return leaves
}

// TopSortedLayers returns the nodes of the graph sorted in layers, where each layer contains nodes with no dependencies.
func (g *DependencyGraph) TopSortedLayers() [][]string {
	layers := [][]string{}

	// Copy the graph
	shrinkingGraph := g.clone()
	for {
		leaves := shrinkingGraph.Leaves()
		if len(leaves) == 0 {
			break
		}

		layers = append(layers, leaves)
		for _, leafNode := range leaves {
			shrinkingGraph.remove(leafNode)
		}
	}

	return layers
}

// removeFromDepmap removes a node from the dependency map.
func removeFromDepmap(dm depmap, key, node string) {
	nodes := dm[key]
	if len(nodes) == 1 {
		// The only element in the Nodeset must be `node`, so we
		// can delete the entry entirely.
		delete(dm, key)
	} else {
		// Otherwise, remove the single node from the Nodeset.
		delete(nodes, node)
	}
}

// remove removes a node and all its edges from the graph.
func (g *DependencyGraph) remove(node string) {
	// Remove edges from things that depend on `node`.
	for dependent := range g.dependents[node] {
		removeFromDepmap(g.dependencies, dependent, node)
	}
	delete(g.dependents, node)

	// Remove all edges from node to the things it depends on.
	for dependency := range g.dependencies[node] {
		removeFromDepmap(g.dependents, dependency, node)
	}
	delete(g.dependencies, node)

	// Finally, remove the node itself.
	delete(g.nodes, node)
}

// TopSorted returns all the nodes in the graph is topological sort order.
// See also `DependencyGraph.TopSortedLayers()`.
func (g *DependencyGraph) TopSorted() []string {
	nodeCount := 0
	layers := g.TopSortedLayers()
	for _, layer := range layers {
		nodeCount += len(layer)
	}

	allNodes := make([]string, 0, nodeCount)
	for _, layer := range layers {
		allNodes = append(allNodes, layer...)
	}

	return allNodes
}

// Dependencies returns all transitive dependencies of the given child node.
func (g *DependencyGraph) Dependencies(child string) Nodeset {
	return g.buildTransitive(child, g.immediateDependencies)
}

// immediateDependencies returns the immediate dependencies of the given node.
func (g *DependencyGraph) immediateDependencies(node string) Nodeset {
	return g.dependencies[node]
}

// Dependents returns all transitive dependents of the given parent node.
func (g *DependencyGraph) Dependents(parent string) Nodeset {
	return g.buildTransitive(parent, g.immediateDependents)
}

// immediateDependents returns the immediate dependents of the given node.
func (g *DependencyGraph) immediateDependents(node string) Nodeset {
	return g.dependents[node]
}

// clone creates a deep copy of the DependencyGraph.
func (g *DependencyGraph) clone() *DependencyGraph {
	return &DependencyGraph{
		dependencies: copyDepmap(g.dependencies),
		dependents:   copyDepmap(g.dependents),
		nodes:        copyNodeset(g.nodes),
	}
}

// buildTransitive builds a transitive closure of nodes starting from the root node.
func (g *DependencyGraph) buildTransitive(root string, nextFn func(string) Nodeset) Nodeset {
	if _, ok := g.nodes[root]; !ok {
		// The root node is not in the graph, so there are no dependencies.
		return nil
	}

	out := make(Nodeset)
	searchNext := []string{root}
	for len(searchNext) > 0 {
		// List of new nodes from this layer of the dependency graph. This is
		// assigned to `searchNext` at the end of the outer "discovery" loop.
		discovered := []string{}
		for _, node := range searchNext {
			// For each node to discover, find the next nodes.
			for nextNode := range nextFn(node) {
				// If we have not seen the node before, add it to the output as well
				// as the list of nodes to traverse in the next iteration.
				if _, ok := out[nextNode]; !ok {
					out[nextNode] = struct{}{}
					discovered = append(discovered, nextNode)
				}
			}
		}
		searchNext = discovered
	}

	return out
}

// copyNodeset creates a deep copy of a Nodeset.
func copyNodeset(s Nodeset) Nodeset {
	out := make(Nodeset, len(s))
	for k, v := range s {
		out[k] = v
	}
	return out
}

// copyDepmap creates a deep copy of a depmap.
func copyDepmap(m depmap) depmap {
	out := make(depmap, len(m))
	for k, v := range m {
		out[k] = copyNodeset(v)
	}
	return out
}

// addNodeToNodeset adds a node to the Nodeset in the depmap.
func addNodeToNodeset(dm depmap, key, node string) {
	nodes, ok := dm[key]
	if !ok {
		nodes = make(Nodeset)
		dm[key] = nodes
	}
	nodes[node] = struct{}{}
}
