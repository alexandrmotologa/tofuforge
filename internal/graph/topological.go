package graph

import (
	"sort"
)

// TopologicalSort computes a valid linear execution order for resources in the DAG.
// Uses Kahn's algorithm with cycle tolerance.
func (g *Graph) TopologicalSort() ([]string, bool) {
	inDegree := make(map[string]int)
	for addr := range g.Nodes {
		inDegree[addr] = 0
	}

	for _, edge := range g.Edges {
		inDegree[edge.Target]++
	}

	// Queue nodes with in-degree 0
	var queue []string
	for addr, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, addr)
		}
	}
	sort.Strings(queue)

	var order []string

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		order = append(order, curr)

		for _, neighbor := range g.OutEdges[curr] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
		sort.Strings(queue)
	}

	// If order contains all nodes, no cycles exist
	hasCycle := len(order) < len(g.Nodes)

	// If there were cycles, append remaining unvisited nodes
	if hasCycle {
		visited := make(map[string]bool)
		for _, addr := range order {
			visited[addr] = true
		}
		var remaining []string
		for addr := range g.Nodes {
			if !visited[addr] {
				remaining = append(remaining, addr)
			}
		}
		sort.Strings(remaining)
		order = append(order, remaining...)
	}

	return order, !hasCycle
}
