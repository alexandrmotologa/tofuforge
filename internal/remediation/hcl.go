package remediation

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// RemediationPlan provides contextual advice and HCL snippets to safeguard a resource.
type RemediationPlan struct {
	ResourceAddress string `json:"resource_address"`
	HCLSnippet      string `json:"hcl_snippet"`
	Guidance        string `json:"guidance"`
}

// GenerateHCLRemediation creates a practical HCL lifecycle block for destructive resources.
func GenerateHCLRemediation(alert model.DestructionAlert) RemediationPlan {
	var lines []string
	lines = append(lines, "  lifecycle {")

	if alert.IsStateful {
		lines = append(lines, "    prevent_destroy       = true")
	}

	if alert.Action == model.ActionReplace {
		lines = append(lines, "    create_before_destroy = true")
	}

	if len(alert.ReplacementKeys) > 0 {
		var quotedKeys []string
		for _, k := range alert.ReplacementKeys {
			quotedKeys = append(quotedKeys, k)
		}
		lines = append(lines, fmt.Sprintf("    # Attribute changes forcing recreation: %s", strings.Join(quotedKeys, ", ")))
	}

	lines = append(lines, "  }")
	snippet := strings.Join(lines, "\n")

	var guidance string
	if alert.IsStateful && alert.Action == model.ActionReplace {
		guidance = "Add prevent_destroy = true to block apply, or migrate data via cloud snapshot / blue-green deployment before modifying replacement attributes."
	} else if alert.Action == model.ActionReplace {
		guidance = "Add create_before_destroy = true to provision the new resource before deleting the old one, avoiding downtime."
	} else {
		guidance = "Verify that this resource deletion is intentional and all downstream services have been re-routed."
	}

	return RemediationPlan{
		ResourceAddress: alert.ResourceAddress,
		HCLSnippet:      snippet,
		Guidance:        guidance,
	}
}
