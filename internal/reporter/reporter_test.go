package reporter

import (
	"bytes"
	"strings"
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func TestGenerateMarkdownReport(t *testing.T) {
	plan := &model.AnalyzedPlan{
		TerraformVersion: "1.8.0",
		FormatVersion:    "1.2",
		Stats: model.SummaryStats{
			TotalResources:      3,
			CreateCount:         1,
			UpdateCount:         1,
			ReplaceCount:        1,
			TotalBlastRadius:    35,
			CriticalAlertsCount: 1,
			SecurityIssuesCount: 1,
		},
		Nodes: []model.ResourceNode{
			{
				Address: "aws_db_instance.main",
				Type:    "aws_db_instance",
				Action:  model.ActionReplace,
				AttributeDiffs: []model.AttributeDiff{
					{
						Path:              "allocated_storage",
						Before:            20,
						After:             50,
						ForcesReplacement: true,
					},
				},
			},
		},
		DestructionAlerts: []model.DestructionAlert{
			{
				ResourceAddress: "aws_db_instance.main",
				ResourceType:    "aws_db_instance",
				Severity:        model.SeverityCritical,
				Action:          model.ActionReplace,
				BlastRadius:     35,
				DownstreamCount: 2,
				DownstreamPaths: []string{"aws_ecs_service.app", "aws_route53_record.db"},
				ReplacementKeys: []string{"allocated_storage"},
				Message:         "Stateful database will be destroyed and recreated.",
			},
		},
		SecurityFindings: []model.SecurityFinding{
			{
				RuleID:      "AWS-SEC-001",
				Title:       "Open SSH Ingress",
				Severity:    model.SeverityCritical,
				Resource:    "aws_security_group_rule.ssh",
				Description: "Exposes port 22 to 0.0.0.0/0",
				Remediation: "Restrict CIDR",
			},
		},
	}

	md := GenerateMarkdownReport(plan)

	if !strings.Contains(md, "CRITICAL DESTRUCTION RISK") {
		t.Errorf("expected markdown to contain 'CRITICAL DESTRUCTION RISK'")
	}
	if !strings.Contains(md, "aws_db_instance.main") {
		t.Errorf("expected markdown to mention resource aws_db_instance.main")
	}
	if !strings.Contains(md, "AWS-SEC-001") {
		t.Errorf("expected markdown to mention rule AWS-SEC-001")
	}
	if !strings.Contains(md, "allocated_storage") {
		t.Errorf("expected markdown to mention attribute allocated_storage")
	}
}

func TestRenderTerminal(t *testing.T) {
	plan := &model.AnalyzedPlan{
		Stats: model.SummaryStats{
			CreateCount: 1,
		},
	}

	var buf bytes.Buffer
	RenderTerminal(&buf, plan, true)

	out := buf.String()
	if !strings.Contains(out, "TOFUFORGE PLAN INSPECTION REPORT") {
		t.Errorf("expected terminal output to contain header")
	}
}
