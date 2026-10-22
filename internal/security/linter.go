package security

import (
	"sort"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// Rule represents a pre-flight IaC security check.
type Rule struct {
	ID          string
	Title       string
	Severity    model.Severity
	Check       func(node *model.ResourceNode) *model.SecurityFinding
}

// Registry maintains all registered security rules across cloud providers.
var Registry = []Rule{}

// RegisterRule adds a rule to the global registry.
func RegisterRule(r Rule) {
	Registry = append(Registry, r)
}

// Lint runs all registered rules against resource nodes in the plan.
func Lint(nodes map[string]*model.ResourceNode) []model.SecurityFinding {
	findings := make([]model.SecurityFinding, 0)

	for _, node := range nodes {
		// Only evaluate resources being created, updated, or replaced
		if node.Action == model.ActionDelete || node.Action == model.ActionNoOp || node.Action == model.ActionRead {
			continue
		}

		for _, rule := range Registry {
			if finding := rule.Check(node); finding != nil {
				node.SecurityFindings = append(node.SecurityFindings, *finding)
				findings = append(findings, *finding)
			}
		}
	}

	// Sort findings by severity (Critical first) and then Resource
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Severity != findings[j].Severity {
			return severityWeight(findings[i].Severity) > severityWeight(findings[j].Severity)
		}
		return findings[i].Resource < findings[j].Resource
	})

	return findings
}

func severityWeight(s model.Severity) int {
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
