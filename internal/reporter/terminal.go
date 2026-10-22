package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// ANSI color codes
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
)

// RenderTerminal writes a colorized analysis report to the provided writer.
func RenderTerminal(w io.Writer, plan *model.AnalyzedPlan, noColor bool) {
	c := func(code, text string) string {
		if noColor {
			return text
		}
		return code + text + colorReset
	}

	fmt.Fprintln(w, c(colorBold+colorCyan, "================================================================================"))
	fmt.Fprintln(w, c(colorBold+colorCyan, "                      TOFUFORGE PLAN INSPECTION REPORT                          "))
	fmt.Fprintln(w, c(colorBold+colorCyan, "================================================================================"))

	// Status Header
	statusStr := c(colorBold+colorGreen, "SAFE TO APPLY")
	if plan.Stats.CriticalAlertsCount > 0 {
		statusStr = c(colorBold+colorRed, "CRITICAL DESTRUCTION RISK")
	} else if plan.Stats.ReplaceCount > 0 || plan.Stats.DeleteCount > 0 {
		statusStr = c(colorBold+colorYellow, "DESTRUCTIVE CHANGES DETECTED")
	} else if plan.Stats.SecurityIssuesCount > 0 {
		statusStr = c(colorBold+colorMagenta, "SECURITY VIOLATIONS DETECTED")
	}

	fmt.Fprintf(w, "Status: %s\n\n", statusStr)

	// Summary Statistics
	fmt.Fprintln(w, c(colorBold, "PLAN SUMMARY:"))
	fmt.Fprintf(w, "  + Create:       %s\n", c(colorGreen, fmt.Sprintf("%d", plan.Stats.CreateCount)))
	fmt.Fprintf(w, "  ~ Update:       %s\n", c(colorYellow, fmt.Sprintf("%d", plan.Stats.UpdateCount)))
	fmt.Fprintf(w, "  - Delete:       %s\n", c(colorRed, fmt.Sprintf("%d", plan.Stats.DeleteCount)))
	fmt.Fprintf(w, "  ± Replace:      %s\n", c(colorMagenta, fmt.Sprintf("%d", plan.Stats.ReplaceCount)))
	fmt.Fprintf(w, "  Total Blast Radius Score:  %s\n", c(colorBold, fmt.Sprintf("%d", plan.Stats.TotalBlastRadius)))
	fmt.Fprintf(w, "  Critical Alerts:           %s\n", c(colorRed, fmt.Sprintf("%d", plan.Stats.CriticalAlertsCount)))
	fmt.Fprintf(w, "  Security Violations:       %s\n\n", c(colorYellow, fmt.Sprintf("%d", plan.Stats.SecurityIssuesCount)))

	// Destruction Alerts
	if len(plan.DestructionAlerts) > 0 {
		fmt.Fprintln(w, c(colorBold+colorRed, "--------------------------------------------------------------------------------"))
		fmt.Fprintln(w, c(colorBold+colorRed, "                   DESTRUCTION & RECREATION ALERTS                              "))
		fmt.Fprintln(w, c(colorBold+colorRed, "--------------------------------------------------------------------------------"))

		for _, a := range plan.DestructionAlerts {
			var sevColor string
			switch a.Severity {
			case model.SeverityCritical:
				sevColor = colorBold + colorRed
			case model.SeverityHigh:
				sevColor = colorYellow
			default:
				sevColor = colorGray
			}

			fmt.Fprintf(w, "[%s] %s (%s)\n", c(sevColor, string(a.Severity)), c(colorBold, a.ResourceAddress), a.ResourceType)
			fmt.Fprintf(w, "  Action:       %s\n", string(a.Action))
			fmt.Fprintf(w, "  Stateful:     %v\n", a.IsStateful)
			fmt.Fprintf(w, "  Blast Radius: %d (Downstream: %d resources)\n", a.BlastRadius, a.DownstreamCount)

			if len(a.ReplacementKeys) > 0 {
				fmt.Fprintf(w, "  Triggers:     %s\n", strings.Join(a.ReplacementKeys, ", "))
			}
			if len(a.DownstreamPaths) > 0 {
				fmt.Fprintf(w, "  Impacts:      %s\n", strings.Join(a.DownstreamPaths, ", "))
			}
			fmt.Fprintf(w, "  Note:         %s\n\n", a.Message)
		}
	}

	// Security Findings
	if len(plan.SecurityFindings) > 0 {
		fmt.Fprintln(w, c(colorBold+colorYellow, "--------------------------------------------------------------------------------"))
		fmt.Fprintln(w, c(colorBold+colorYellow, "                   PRE-FLIGHT SECURITY AUDIT                                    "))
		fmt.Fprintln(w, c(colorBold+colorYellow, "--------------------------------------------------------------------------------"))

		for _, f := range plan.SecurityFindings {
			var sevColor string
			switch f.Severity {
			case model.SeverityCritical:
				sevColor = colorBold + colorRed
			case model.SeverityHigh:
				sevColor = colorYellow
			case model.SeverityMedium:
				sevColor = colorCyan
			default:
				sevColor = colorGray
			}

			fmt.Fprintf(w, "[%s] %s - %s\n", c(sevColor, string(f.Severity)), f.RuleID, c(colorBold, f.Title))
			fmt.Fprintf(w, "  Resource:    %s (%s)\n", f.Resource, f.Type)
			fmt.Fprintf(w, "  Description: %s\n", f.Description)
			fmt.Fprintf(w, "  Remediation: %s\n\n", c(colorGreen, f.Remediation))
		}
	}

	fmt.Fprintln(w, c(colorBold+colorCyan, "================================================================================"))
}
