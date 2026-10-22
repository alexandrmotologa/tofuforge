package reporter

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func TestGenerateStandaloneHTML(t *testing.T) {
	plan := &model.AnalyzedPlan{
		FormatVersion: "1.2",
		Stats: model.SummaryStats{
			TotalResources: 5,
			CreateCount:    5,
		},
		Nodes: []model.ResourceNode{
			{
				Address: "aws_vpc.production",
				Type:    "aws_vpc",
				Action:  model.ActionCreate,
			},
		},
	}

	html, err := GenerateStandaloneHTML(plan)
	if err != nil {
		t.Fatalf("GenerateStandaloneHTML failed: %v", err)
	}

	if !strings.Contains(html, "__TOFUFORGE_PRELOADED_PLAN__") {
		t.Errorf("expected HTML to contain __TOFUFORGE_PRELOADED_PLAN__")
	}

	if !strings.Contains(html, "aws_vpc.production") {
		t.Errorf("expected HTML to embed aws_vpc.production")
	}
}
