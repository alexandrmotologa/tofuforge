package remediation

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func TestGenerateHCLRemediation(t *testing.T) {
	alert := model.DestructionAlert{
		ResourceAddress: "aws_db_instance.customer_orders",
		ResourceType:    "aws_db_instance",
		Action:          model.ActionReplace,
		IsStateful:      true,
		ReplacementKeys: []string{"identifier"},
	}

	plan := GenerateHCLRemediation(alert)

	if !strings.Contains(plan.HCLSnippet, "prevent_destroy       = true") {
		t.Errorf("expected prevent_destroy in snippet")
	}

	if !strings.Contains(plan.HCLSnippet, "create_before_destroy = true") {
		t.Errorf("expected create_before_destroy in snippet")
	}

	if !strings.Contains(plan.HCLSnippet, "identifier") {
		t.Errorf("expected identifier mentioned in snippet comments")
	}
}
