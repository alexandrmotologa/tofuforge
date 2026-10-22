package engine

import (
	"sort"

	"github.com/alexandrmotologa/tofuforge/internal/config"
	"github.com/alexandrmotologa/tofuforge/internal/cost"
	"github.com/alexandrmotologa/tofuforge/internal/graph"
	"github.com/alexandrmotologa/tofuforge/internal/model"
	"github.com/alexandrmotologa/tofuforge/internal/security"
)

// Analyze processes a raw plan with default governance configuration.
func Analyze(raw *model.RawPlan) *model.AnalyzedPlan {
	return AnalyzeWithConfig(raw, config.DefaultConfig())
}

// AnalyzeWithConfig processes a raw plan applying user-defined configuration policies.
func AnalyzeWithConfig(raw *model.RawPlan, cfg *config.Config) *model.AnalyzedPlan {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	g := graph.BuildGraph(raw)
	alerts := g.ComputeBlastRadius()
	order, _ := g.TopologicalSort()
	findings := security.LintWithConfig(g.Nodes, cfg)

	// Collect nodes slice sorted by address and compute cost estimates
	nodes := make([]model.ResourceNode, 0, len(g.Nodes))
	for _, n := range g.Nodes {
		n.Cost = cost.EstimateNodeCost(n)
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

	// Compute financial cost metrics
	beforeCost, afterCost, deltaCost := cost.ComputePlanCostAggregates(nodes)
	stats.TotalMonthlyBefore = beforeCost
	stats.TotalMonthlyAfter = afterCost
	stats.TotalMonthlyDelta = deltaCost

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
