package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/model"
	"github.com/alexandrmotologa/tofuforge/internal/parser"
)

func TestAnalyze_AWSProductionEcommerce(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "fixtures", "aws_production_ecommerce.json")
	file, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("failed to open fixture: %v", err)
	}
	defer file.Close()

	raw, err := parser.ParsePlan(file)
	if err != nil {
		t.Fatalf("failed to parse plan: %v", err)
	}

	analyzed := Analyze(raw)

	if analyzed.Stats.TotalResources != 8 {
		t.Errorf("expected 8 total resources, got %d", analyzed.Stats.TotalResources)
	}

	if analyzed.Stats.ReplaceCount != 1 {
		t.Errorf("expected 1 replacement (aws_db_instance), got %d", analyzed.Stats.ReplaceCount)
	}

	if analyzed.Stats.CriticalAlertsCount != 1 {
		t.Errorf("expected 1 critical destruction alert, got %d", analyzed.Stats.CriticalAlertsCount)
	}

	// Security findings: SSH port 22 open + S3 unencrypted
	if analyzed.Stats.SecurityIssuesCount < 2 {
		t.Errorf("expected at least 2 security issues, got %d", analyzed.Stats.SecurityIssuesCount)
	}

	// Verify blast radius of aws_db_instance.customer_orders_postgresql
	var dbAlert *model.DestructionAlert
	for _, a := range analyzed.DestructionAlerts {
		if a.ResourceAddress == "aws_db_instance.customer_orders_postgresql" {
			dbAlert = &a
			break
		}
	}

	if dbAlert == nil {
		t.Fatalf("expected alert for database instance")
	}

	if dbAlert.DownstreamCount != 2 {
		t.Errorf("expected 2 downstream resources (ecs_service, route53_record), got %d", dbAlert.DownstreamCount)
	}

	if dbAlert.Severity != model.SeverityCritical {
		t.Errorf("expected CRITICAL severity, got %s", dbAlert.Severity)
	}
}

func TestAnalyze_CleanInfrastructure(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "fixtures", "clean_infrastructure.json")
	file, err := os.Open(fixturePath)
	if err != nil {
		t.Fatalf("failed to open clean fixture: %v", err)
	}
	defer file.Close()

	raw, err := parser.ParsePlan(file)
	if err != nil {
		t.Fatalf("failed to parse clean plan: %v", err)
	}

	analyzed := Analyze(raw)

	if analyzed.Stats.CreateCount != 4 {
		t.Errorf("expected 4 created resources, got %d", analyzed.Stats.CreateCount)
	}

	if analyzed.Stats.DeleteCount != 0 || analyzed.Stats.ReplaceCount != 0 {
		t.Errorf("expected 0 deletions/replacements in clean plan")
	}

	if analyzed.Stats.CriticalAlertsCount != 0 {
		t.Errorf("expected 0 critical alerts in clean plan, got %d", analyzed.Stats.CriticalAlertsCount)
	}

	if analyzed.Stats.SecurityIssuesCount != 0 {
		t.Errorf("expected 0 security violations in clean plan, got %d", analyzed.Stats.SecurityIssuesCount)
	}
}
