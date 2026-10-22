package diff

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// ActionChange captures a transition in planned action between two plan revisions.
type ActionChange struct {
	Address    string       `json:"address"`
	BaseAction model.Action `json:"base_action"`
	TargetAction model.Action `json:"target_action"`
}

// PlanComparison represents the differential analysis between two plan revisions.
type PlanComparison struct {
	BaseStats            model.SummaryStats     `json:"base_stats"`
	TargetStats          model.SummaryStats     `json:"target_stats"`
	BlastRadiusDelta     int                    `json:"blast_radius_delta"`
	CostDelta            float64                `json:"cost_delta"`
	ResolvedDestruction  []model.DestructionAlert `json:"resolved_destruction"`
	NewDestruction       []model.DestructionAlert `json:"new_destruction"`
	PersistingDestruction []model.DestructionAlert `json:"persisting_destruction"`
	ResolvedSecurity     []model.SecurityFinding `json:"resolved_security"`
	NewSecurity          []model.SecurityFinding `json:"new_security"`
	AddedResources       []string               `json:"added_resources"`
	RemovedResources     []string               `json:"removed_resources"`
	ActionChanges        []ActionChange         `json:"action_changes"`
}

// ComparePlans evaluates the delta between a base plan and an updated target plan.
func ComparePlans(base, target *model.AnalyzedPlan) *PlanComparison {
	comp := &PlanComparison{
		BaseStats:             base.Stats,
		TargetStats:           target.Stats,
		BlastRadiusDelta:     target.Stats.TotalBlastRadius - base.Stats.TotalBlastRadius,
		CostDelta:            target.Stats.TotalMonthlyDelta - base.Stats.TotalMonthlyDelta,
		ResolvedDestruction:  make([]model.DestructionAlert, 0),
		NewDestruction:       make([]model.DestructionAlert, 0),
		PersistingDestruction: make([]model.DestructionAlert, 0),
		ResolvedSecurity:     make([]model.SecurityFinding, 0),
		NewSecurity:          make([]model.SecurityFinding, 0),
		AddedResources:       make([]string, 0),
		RemovedResources:     make([]string, 0),
		ActionChanges:        make([]ActionChange, 0),
	}

	// 1. Compare Destruction Alerts
	baseAlertMap := make(map[string]model.DestructionAlert)
	for _, a := range base.DestructionAlerts {
		baseAlertMap[a.ResourceAddress] = a
	}

	targetAlertMap := make(map[string]model.DestructionAlert)
	for _, a := range target.DestructionAlerts {
		targetAlertMap[a.ResourceAddress] = a
	}

	for addr, bAlert := range baseAlertMap {
		if _, exists := targetAlertMap[addr]; exists {
			comp.PersistingDestruction = append(comp.PersistingDestruction, bAlert)
		} else {
			comp.ResolvedDestruction = append(comp.ResolvedDestruction, bAlert)
		}
	}

	for addr, tAlert := range targetAlertMap {
		if _, exists := baseAlertMap[addr]; !exists {
			comp.NewDestruction = append(comp.NewDestruction, tAlert)
		}
	}

	// 2. Compare Security Findings
	baseSecKey := func(f model.SecurityFinding) string {
		return fmt.Sprintf("%s::%s", f.RuleID, f.Resource)
	}

	baseSecMap := make(map[string]model.SecurityFinding)
	for _, f := range base.SecurityFindings {
		baseSecMap[baseSecKey(f)] = f
	}

	targetSecMap := make(map[string]model.SecurityFinding)
	for _, f := range target.SecurityFindings {
		targetSecMap[baseSecKey(f)] = f
	}

	for k, bSec := range baseSecMap {
		if _, exists := targetSecMap[k]; !exists {
			comp.ResolvedSecurity = append(comp.ResolvedSecurity, bSec)
		}
	}

	for k, tSec := range targetSecMap {
		if _, exists := baseSecMap[k]; !exists {
			comp.NewSecurity = append(comp.NewSecurity, tSec)
		}
	}

	// 3. Compare Nodes & Action Changes
	baseNodeMap := make(map[string]model.ResourceNode)
	for _, n := range base.Nodes {
		baseNodeMap[n.Address] = n
	}

	targetNodeMap := make(map[string]model.ResourceNode)
	for _, n := range target.Nodes {
		targetNodeMap[n.Address] = n
	}

	for addr, bNode := range baseNodeMap {
		if tNode, exists := targetNodeMap[addr]; exists {
			if bNode.Action != tNode.Action {
				comp.ActionChanges = append(comp.ActionChanges, ActionChange{
					Address:      addr,
					BaseAction:   bNode.Action,
					TargetAction: tNode.Action,
				})
			}
		} else {
			comp.RemovedResources = append(comp.RemovedResources, addr)
		}
	}

	for addr := range targetNodeMap {
		if _, exists := baseNodeMap[addr]; !exists {
			comp.AddedResources = append(comp.AddedResources, addr)
		}
	}

	sort.Strings(comp.AddedResources)
	sort.Strings(comp.RemovedResources)
	sort.Slice(comp.ActionChanges, func(i, j int) bool {
		return comp.ActionChanges[i].Address < comp.ActionChanges[j].Address
	})

	return comp
}

// GenerateMarkdownDiff formats a GitHub pull request revision comparison summary.
func GenerateMarkdownDiff(comp *PlanComparison) string {
	var sb strings.Builder

	sb.WriteString("## TofuForge Plan Revision Comparison\n\n")

	// Overview summary table
	sb.WriteString("| Metric | Prior Revision | Current Revision | Delta |\n")
	sb.WriteString("| :--- | :--- | :--- | :--- |\n")
	sb.WriteString(fmt.Sprintf("| **Total Blast Radius** | `%d` | `%d` | `%+d` |\n",
		comp.BaseStats.TotalBlastRadius, comp.TargetStats.TotalBlastRadius, comp.BlastRadiusDelta))
	sb.WriteString(fmt.Sprintf("| **Critical Alerts** | `%d` | `%d` | `%+d` |\n",
		comp.BaseStats.CriticalAlertsCount, comp.TargetStats.CriticalAlertsCount,
		comp.TargetStats.CriticalAlertsCount-comp.BaseStats.CriticalAlertsCount))
	sb.WriteString(fmt.Sprintf("| **Security Flaws** | `%d` | `%d` | `%+d` |\n",
		comp.BaseStats.SecurityIssuesCount, comp.TargetStats.SecurityIssuesCount,
		comp.TargetStats.SecurityIssuesCount-comp.BaseStats.SecurityIssuesCount))
	sb.WriteString(fmt.Sprintf("| **Monthly Cost Delta** | `$%.2f/mo` | `$%.2f/mo` | `%+.2f/mo` |\n\n",
		comp.BaseStats.TotalMonthlyDelta, comp.TargetStats.TotalMonthlyDelta, comp.CostDelta))

	// Resolved Destructions
	if len(comp.ResolvedDestruction) > 0 {
		sb.WriteString("### ✅ Resolved Destructive Risks\n\n")
		sb.WriteString("The following critical resources are no longer scheduled for deletion or recreation:\n\n")
		for _, a := range comp.ResolvedDestruction {
			sb.WriteString(fmt.Sprintf("- ✅ **%s** (`%s`) - Former %s (Blast Radius: %d)\n",
				a.ResourceAddress, a.ResourceType, a.Action, a.BlastRadius))
		}
		sb.WriteString("\n")
	}

	// New Destructions
	if len(comp.NewDestruction) > 0 {
		sb.WriteString("### 🚨 Newly Introduced Destructive Risks\n\n")
		for _, a := range comp.NewDestruction {
			sb.WriteString(fmt.Sprintf("- ⚠️ **%s** (`%s`) - Action: `%s` (Blast Radius: %d)\n",
				a.ResourceAddress, a.ResourceType, a.Action, a.BlastRadius))
		}
		sb.WriteString("\n")
	}

	// Resolved Security Findings
	if len(comp.ResolvedSecurity) > 0 {
		sb.WriteString("### 🛡️ Resolved Security Findings\n\n")
		for _, s := range comp.ResolvedSecurity {
			sb.WriteString(fmt.Sprintf("- ✅ **%s**: %s on `%s`\n", s.RuleID, s.Title, s.Resource))
		}
		sb.WriteString("\n")
	}

	// Action Transitions
	if len(comp.ActionChanges) > 0 {
		sb.WriteString("### 🔄 Resource Action Transitions\n\n")
		sb.WriteString("| Resource | Prior Action | Revised Action |\n")
		sb.WriteString("| :--- | :--- | :--- |\n")
		for _, ac := range comp.ActionChanges {
			sb.WriteString(fmt.Sprintf("| `%s` | `%s` | **`%s`** |\n", ac.Address, ac.BaseAction, ac.TargetAction))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n*Generated by [TofuForge](https://github.com/alexandrmotologa/tofuforge) plan comparator.*\n")
	return sb.String()
}

// RenderTerminalDiff outputs ANSI colorized revision comparison.
func RenderTerminalDiff(w io.Writer, comp *PlanComparison, noColor bool) {
	c := func(code, text string) string {
		if noColor {
			return text
		}
		return code + text + "\033[0m"
	}

	fmt.Fprintln(w, c("\033[1m\033[36m", "================================================================================"))
	fmt.Fprintln(w, c("\033[1m\033[36m", "                    TOFUFORGE PLAN REVISION COMPARISON                          "))
	fmt.Fprintln(w, c("\033[1m\033[36m", "================================================================================"))

	fmt.Fprintf(w, "Blast Radius Score Delta:  %s (Prior: %d -> Current: %d)\n",
		c("\033[1m", fmt.Sprintf("%+d", comp.BlastRadiusDelta)),
		comp.BaseStats.TotalBlastRadius, comp.TargetStats.TotalBlastRadius)

	fmt.Fprintf(w, "Critical Alerts Delta:      %+d\n",
		comp.TargetStats.CriticalAlertsCount-comp.BaseStats.CriticalAlertsCount)
	fmt.Fprintf(w, "Security Flaws Delta:       %+d\n",
		comp.TargetStats.SecurityIssuesCount-comp.BaseStats.SecurityIssuesCount)

	if len(comp.ResolvedDestruction) > 0 {
		fmt.Fprintln(w, c("\033[1m\033[32m", "\n[RESOLVED DESTRUCTIONS]:"))
		for _, a := range comp.ResolvedDestruction {
			fmt.Fprintf(w, "  ✓ %s (%s) - no longer marked for %s\n",
				c("\033[32m", a.ResourceAddress), a.ResourceType, a.Action)
		}
	}

	if len(comp.NewDestruction) > 0 {
		fmt.Fprintln(w, c("\033[1m\033[31m", "\n[NEW DESTRUCTIONS INTRODUCED]:"))
		for _, a := range comp.NewDestruction {
			fmt.Fprintf(w, "  ! %s (%s) - action: %s (Blast: %d)\n",
				c("\033[31m", a.ResourceAddress), a.ResourceType, a.Action, a.BlastRadius)
		}
	}

	if len(comp.ResolvedSecurity) > 0 {
		fmt.Fprintln(w, c("\033[1m\033[32m", "\n[RESOLVED SECURITY FINDINGS]:"))
		for _, s := range comp.ResolvedSecurity {
			fmt.Fprintf(w, "  ✓ %s: %s (%s)\n", s.RuleID, s.Title, s.Resource)
		}
	}

	if len(comp.ActionChanges) > 0 {
		fmt.Fprintln(w, c("\033[1m", "\n[ACTION TRANSITIONS]:"))
		for _, ac := range comp.ActionChanges {
			fmt.Fprintf(w, "  • %s: %s -> %s\n", ac.Address, ac.BaseAction, ac.TargetAction)
		}
	}

	fmt.Fprintln(w, c("\033[1m\033[36m", "\n================================================================================"))
}
