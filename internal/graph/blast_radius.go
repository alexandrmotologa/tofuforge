package graph

import (
	"fmt"
	"sort"

	"github.com/alexandrmotologa/tofuforge/internal/model"
	"github.com/alexandrmotologa/tofuforge/internal/parser"
)

// ComputeBlastRadius evaluates downstream cascades and risk scores for all nodes in the graph.
func (g *Graph) ComputeBlastRadius() []model.DestructionAlert {
	alerts := make([]model.DestructionAlert, 0)

	// Step 1: Compute transitive downstream reachable nodes for each node
	for addr, node := range g.Nodes {
		downstream := g.findTransitiveDownstream(addr)
		node.DownstreamIDs = downstream
		node.DownstreamCount = len(downstream)

		// Calculate blast radius score
		if parser.IsDestructive(node.Action) {
			score := calculateNodeScore(node, g)
			node.BlastRadiusScore = score

			// Create destruction alert
			alert := g.createDestructionAlert(node)
			alerts = append(alerts, alert)
		} else {
			node.BlastRadiusScore = 0
		}
	}

	// Sort alerts by severity and blast radius descending
	sort.Slice(alerts, func(i, j int) bool {
		if alerts[i].Severity != alerts[j].Severity {
			return severityRank(alerts[i].Severity) > severityRank(alerts[j].Severity)
		}
		return alerts[i].BlastRadius > alerts[j].BlastRadius
	})

	return alerts
}

// findTransitiveDownstream performs BFS forward from startAddr to find all reachable downstream nodes.
func (g *Graph) findTransitiveDownstream(startAddr string) []string {
	visited := make(map[string]bool)
	queue := []string{startAddr}
	visited[startAddr] = true

	downstream := make([]string, 0)

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for _, neighbor := range g.OutEdges[curr] {
			if !visited[neighbor] {
				visited[neighbor] = true
				downstream = append(downstream, neighbor)
				queue = append(queue, neighbor)
			}
		}
	}

	sort.Strings(downstream)
	return downstream
}

func calculateNodeScore(node *model.ResourceNode, g *Graph) int {
	score := 0

	// Base score based on statefulness
	if node.IsStateful {
		score += 25
	} else {
		score += 5
	}

	// Replacement penalty vs simple deletion
	if node.Action == model.ActionReplace {
		score += 5
	}

	// Downstream impact
	for _, downID := range node.DownstreamIDs {
		downNode, exists := g.Nodes[downID]
		if exists && downNode.IsStateful {
			score += 15
		} else {
			score += 4
		}
	}

	return score
}

func (g *Graph) createDestructionAlert(node *model.ResourceNode) model.DestructionAlert {
	var sev model.Severity
	if node.IsStateful || node.BlastRadiusScore >= 25 || node.DownstreamCount >= 3 {
		sev = model.SeverityCritical
	} else if node.BlastRadiusScore >= 10 || node.DownstreamCount >= 1 {
		sev = model.SeverityHigh
	} else {
		sev = model.SeverityMedium
	}

	// Gather replacement keys
	var replacementKeys []string
	for _, diff := range node.AttributeDiffs {
		if diff.ForcesReplacement {
			replacementKeys = append(replacementKeys, diff.Path)
		}
	}

	var msg string
	if node.Action == model.ActionReplace {
		if node.IsStateful {
			msg = fmt.Sprintf("CRITICAL: Stateful %s (%s) is marked for DESTROY AND RECREATE. Potential persistent data loss. Impacting %d downstream resources.",
				node.Category, node.Address, node.DownstreamCount)
		} else {
			msg = fmt.Sprintf("Resource %s will be replaced (destroyed and recreated). Impacting %d downstream resources.",
				node.Address, node.DownstreamCount)
		}
	} else {
		if node.IsStateful {
			msg = fmt.Sprintf("CRITICAL: Stateful %s (%s) is marked for DELETION. Impacting %d downstream resources.",
				node.Category, node.Address, node.DownstreamCount)
		} else {
			msg = fmt.Sprintf("Resource %s will be deleted. Impacting %d downstream resources.",
				node.Address, node.DownstreamCount)
		}
	}

	return model.DestructionAlert{
		ResourceAddress: node.Address,
		ResourceType:    node.Type,
		Category:        node.Category,
		Action:          node.Action,
		Severity:        sev,
		IsStateful:      node.IsStateful,
		BlastRadius:     node.BlastRadiusScore,
		DownstreamCount: node.DownstreamCount,
		DownstreamPaths: node.DownstreamIDs,
		Message:         msg,
		ReplacementKeys: replacementKeys,
	}
}

func severityRank(s model.Severity) int {
	switch s {
	case model.SeverityCritical:
		return 4
	case model.SeverityHigh:
		return 3
	case model.SeverityMedium:
		return 2
	case model.SeverityLow:
		return 1
	default:
		return 0
	}
}
