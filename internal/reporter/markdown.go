package reporter

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// GenerateMarkdownReport produces a GitHub PR comment markdown summary.
func GenerateMarkdownReport(plan *model.AnalyzedPlan) string {
	var sb strings.Builder

	// Title & Badge status
	statusBadge := "✅ SAFE TO APPLY"
	if plan.Stats.CriticalAlertsCount > 0 {
		statusBadge = "🚨 CRITICAL DESTRUCTION RISK"
	} else if plan.Stats.ReplaceCount > 0 || plan.Stats.DeleteCount > 0 {
		statusBadge = "⚠️ DESTRUCTIVE CHANGES DETECTED"
	} else if plan.Stats.SecurityIssuesCount > 0 {
		statusBadge = "🛡️ SECURITY VIOLATIONS FOUND"
	}

	sb.WriteString("## TofuForge Plan Analysis Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Status:** `%s`\n\n", statusBadge))

	// Overview Metrics Table
	sb.WriteString("| Metric | Count | Details |\n")
	sb.WriteString("| :--- | :--- | :--- |\n")
	sb.WriteString(fmt.Sprintf("| **Resources to Create** | `+%d` | New resources added |\n", plan.Stats.CreateCount))
	sb.WriteString(fmt.Sprintf("| **Resources to Update** | `~%d` | In-place attribute changes |\n", plan.Stats.UpdateCount))
	sb.WriteString(fmt.Sprintf("| **Resources to Delete** | `-%d` | Permanent deletions |\n", plan.Stats.DeleteCount))
	sb.WriteString(fmt.Sprintf("| **Resources to Replace** | `±%d` | Destroy and recreate cycles |\n", plan.Stats.ReplaceCount))
	sb.WriteString(fmt.Sprintf("| **Total Blast Radius Score** | `%d` | Cumulative downstream risk |\n", plan.Stats.TotalBlastRadius))
	sb.WriteString(fmt.Sprintf("| **Critical Destruction Alerts** | `%d` | Stateful or cascading risks |\n", plan.Stats.CriticalAlertsCount))
	sb.WriteString(fmt.Sprintf("| **Security Findings** | `%d` | Pre-flight policy violations |\n\n", plan.Stats.SecurityIssuesCount))

	// Destruction Alerts Section
	if len(plan.DestructionAlerts) > 0 {
		sb.WriteString("### ⚠️ Destruction & Replacement Guard\n\n")
		sb.WriteString("The following resources undergo deletion or replacement. Stateful resources or those with downstream dependencies require explicit verification:\n\n")
		sb.WriteString("| Severity | Action | Resource Address | Type | Blast Radius | Downstream | Replacement Triggers |\n")
		sb.WriteString("| :--- | :--- | :--- | :--- | :--- | :--- | :--- |\n")

		for _, a := range plan.DestructionAlerts {
			sevIcon := "🔴"
			if a.Severity == model.SeverityHigh {
				sevIcon = "🟠"
			} else if a.Severity == model.SeverityMedium {
				sevIcon = "🟡"
			}

			actStr := string(a.Action)
			if a.Action == model.ActionReplace {
				actStr = "DESTROY & RECREATE"
			}

			triggers := "n/a"
			if len(a.ReplacementKeys) > 0 {
				triggers = strings.Join(a.ReplacementKeys, ", ")
			}

			sb.WriteString(fmt.Sprintf("| %s **%s** | `%s` | `%s` | `%s` | `%d` | `%d` | `%s` |\n",
				sevIcon, a.Severity, actStr, a.ResourceAddress, a.ResourceType, a.BlastRadius, a.DownstreamCount, triggers))
		}
		sb.WriteString("\n")

		// Downstream cascade details if any
		hasDownstream := false
		for _, a := range plan.DestructionAlerts {
			if len(a.DownstreamPaths) > 0 {
				hasDownstream = true
				break
			}
		}

		if hasDownstream {
			sb.WriteString("<details><summary><b>View Downstream Cascades</b></summary>\n\n")
			for _, a := range plan.DestructionAlerts {
				if len(a.DownstreamPaths) > 0 {
					sb.WriteString(fmt.Sprintf("**%s** impacts:\n", a.ResourceAddress))
					for _, dp := range a.DownstreamPaths {
						sb.WriteString(fmt.Sprintf("- `%s`\n", dp))
					}
					sb.WriteString("\n")
				}
			}
			sb.WriteString("</details>\n\n")
		}
	}

	// Security Findings Section
	if len(plan.SecurityFindings) > 0 {
		sb.WriteString("### 🛡️ Pre-Flight Security Audit\n\n")
		sb.WriteString("| Severity | Rule ID | Title | Resource | Remediation |\n")
		sb.WriteString("| :--- | :--- | :--- | :--- | :--- |\n")

		for _, f := range plan.SecurityFindings {
			sevIcon := "🔴"
			if f.Severity == model.SeverityHigh {
				sevIcon = "🟠"
			} else if f.Severity == model.SeverityMedium {
				sevIcon = "🟡"
			} else if f.Severity == model.SeverityLow {
				sevIcon = "🔵"
			}

			sb.WriteString(fmt.Sprintf("| %s **%s** | `%s` | %s | `%s` | %s |\n",
				sevIcon, f.Severity, f.RuleID, f.Title, f.Resource, f.Remediation))
		}
		sb.WriteString("\n")
	}

	// Detailed Resource Diffs Section
	var changedNodes []model.ResourceNode
	for _, n := range plan.Nodes {
		if n.Action != model.ActionNoOp && n.Action != model.ActionRead && len(n.AttributeDiffs) > 0 {
			changedNodes = append(changedNodes, n)
		}
	}

	if len(changedNodes) > 0 {
		sb.WriteString("### 📝 Resource Changes & Attribute Diffs\n\n")
		sb.WriteString("<details><summary><b>Inspect Attribute Details</b> (Click to expand)</summary>\n\n")

		for _, n := range changedNodes {
			sb.WriteString(fmt.Sprintf("#### `%s` (%s)\n\n", n.Address, n.Action))
			sb.WriteString("| Attribute | Before | After | Forces Recreate |\n")
			sb.WriteString("| :--- | :--- | :--- | :--- |\n")

			for _, d := range n.AttributeDiffs {
				recreateStr := "No"
				if d.ForcesReplacement {
					recreateStr = "🚨 **YES**"
				}
				beforeStr := formatVal(d.Before)
				afterStr := formatVal(d.After)
				sb.WriteString(fmt.Sprintf("| `%s` | `%s` | `%s` | %s |\n",
					d.Path, beforeStr, afterStr, recreateStr))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("</details>\n\n")
	}

	sb.WriteString("---\n")
	sb.WriteString("*Generated by [TofuForge](https://github.com/alexandrmotologa/tofuforge) plan visualizer and destruction guard.*\n")

	return sb.String()
}

func formatVal(v interface{}) string {
	if v == nil {
		return "(none)"
	}
	s := fmt.Sprintf("%v", v)
	if len(s) > 40 {
		return s[:37] + "..."
	}
	return s
}
