package security

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/config"
	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// Rule represents a pre-flight IaC security check.
type Rule struct {
	ID       string
	Title    string
	Severity model.Severity
	Check    func(node *model.ResourceNode) *model.SecurityFinding
}

// Registry maintains all registered security rules across cloud providers.
var Registry = []Rule{}

// RegisterRule adds a rule to the global registry.
func RegisterRule(r Rule) {
	Registry = append(Registry, r)
}

// Lint runs all registered rules against resource nodes in the plan with default config.
func Lint(nodes map[string]*model.ResourceNode) []model.SecurityFinding {
	return LintWithConfig(nodes, config.DefaultConfig())
}

// LintWithConfig runs all registered rules and checks required tags according to policy config.
func LintWithConfig(nodes map[string]*model.ResourceNode, cfg *config.Config) []model.SecurityFinding {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	ignoreMap := make(map[string]bool)
	for _, ign := range cfg.Rules.Ignore {
		ignoreMap[ign] = true
	}

	findings := make([]model.SecurityFinding, 0)

	for _, node := range nodes {
		// Only evaluate resources being created, updated, or replaced
		if node.Action == model.ActionDelete || node.Action == model.ActionNoOp || node.Action == model.ActionRead {
			continue
		}

		// 1. Evaluate registered cloud rules
		for _, rule := range Registry {
			if ignoreMap[rule.ID] {
				continue
			}

			if finding := rule.Check(node); finding != nil {
				node.SecurityFindings = append(node.SecurityFindings, *finding)
				findings = append(findings, *finding)
			}
		}

		// 2. Evaluate mandatory tags enforcement if configured and action is create/replace
		if len(cfg.TagsEnforcement.Required) > 0 && (node.Action == model.ActionCreate || node.Action == model.ActionReplace) {
			if missing := checkMissingTags(node, cfg.TagsEnforcement.Required); len(missing) > 0 {
				tagFinding := model.SecurityFinding{
					RuleID:      "TAG-SEC-001",
					Title:       "Resource Missing Required Governance Tags",
					Severity:    model.SeverityMedium,
					Resource:    node.Address,
					Type:        node.Type,
					Description: fmt.Sprintf("Missing mandatory enterprise tags: %s", strings.Join(missing, ", ")),
					Remediation: fmt.Sprintf("Define tags = { %s } on this resource block.", formatTagRemediation(missing)),
					Attribute:   "tags",
				}
				node.SecurityFindings = append(node.SecurityFindings, tagFinding)
				findings = append(findings, tagFinding)
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

func checkMissingTags(node *model.ResourceNode, required []string) []string {
	after := node.RawAfter
	if after == nil {
		return required
	}

	// Check tags or labels map
	tagMap := make(map[string]bool)

	extractKeys := func(val interface{}) {
		if m, ok := val.(map[string]interface{}); ok {
			for k := range m {
				tagMap[strings.ToLower(k)] = true
			}
		}
	}

	if t, exists := after["tags"]; exists {
		extractKeys(t)
	}
	if l, exists := after["labels"]; exists {
		extractKeys(l)
	}
	if tAll, exists := after["tags_all"]; exists {
		extractKeys(tAll)
	}

	var missing []string
	for _, req := range required {
		if !tagMap[strings.ToLower(req)] {
			missing = append(missing, req)
		}
	}

	return missing
}

func formatTagRemediation(missing []string) string {
	var parts []string
	for _, m := range missing {
		parts = append(parts, fmt.Sprintf("%s = \"<value>\"", m))
	}
	return strings.Join(parts, ", ")
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
