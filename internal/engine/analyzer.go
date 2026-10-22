package engine

import (
	"sort"

	"github.com/alexandrmotologa/tofuforge/internal/graph"
	"github.com/alexandrmotologa/tofuforge/internal/model"
	"github.com/alexandrmotologa/tofuforge/internal/security"
)

// Analyze processes a parsed raw plan through DAG construction, blast radius calculation, and security linting.
func Analyze(raw *model.RawPlan) *model.AnalyzedPlan {
	g := graph.BuildGraph(raw)
	alerts := g.ComputeBlastRadius()
	order, _ := g.TopologicalSort()
	findings := security.Lint(g.Nodes)

	// Collect nodes slice sorted by address
	nodes := make([]model.ResourceNode, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		nodes = append(nodes, *n)
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Address < nodes[j].Address
	})

	// Compute summary statistics
	stats := model.SummaryStats{
		TotalResources:      len(nodes),
		SecurityIssuesCount: len(findings),
	}

	for _, n := range nodes {
		switch n.Action {
		case model.ActionCreate:
			stats.CreateCount++
		case model.ActionUpdate:
			stats.UpdateCount++
		case model.ActionDelete:
			stats.DeleteCount++
		case model.ActionReplace:
			stats.ReplaceCount++
		case model.ActionNoOp:
			stats.NoOpCount++
		case model.ActionRead:
			stats.ReadCount++
		}
		stats.TotalBlastRadius += n.BlastRadiusScore
	}

	for _, a := range alerts {
		if a.Severity == model.SeverityCritical {
			stats.CriticalAlertsCount++
		}
	}

	return &model.AnalyzedPlan{
		TerraformVersion:  raw.TerraformVersion,
		FormatVersion:     raw.FormatVersion,
		Stats:             stats,
		Nodes:             nodes,
		Edges:             g.Edges,
		DestructionAlerts: alerts,
		SecurityFindings:  findings,
		TopologicalOrder:  order,
	}
}
