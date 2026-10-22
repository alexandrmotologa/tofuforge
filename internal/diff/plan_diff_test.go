package diff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/engine"
	"github.com/alexandrmotologa/tofuforge/internal/parser"
)

func TestComparePlans_ResolvedDatabaseRecreation(t *testing.T) {
	basePath := filepath.Join("..", "..", "fixtures", "aws_production_ecommerce.json")
	targetPath := filepath.Join("..", "..", "fixtures", "aws_production_fixed.json")

	baseFile, err := os.Open(basePath)
	if err != nil {
		t.Fatalf("failed to open base fixture: %v", err)
	}
	defer baseFile.Close()

	targetFile, err := os.Open(targetPath)
	if err != nil {
		t.Fatalf("failed to open target fixture: %v", err)
	}
	defer targetFile.Close()

	rawBase, _ := parser.ParsePlan(baseFile)
	rawTarget, _ := parser.ParsePlan(targetFile)

	analyzedBase := engine.Analyze(rawBase)
	analyzedTarget := engine.Analyze(rawTarget)

	comp := ComparePlans(analyzedBase, analyzedTarget)

	if len(comp.ResolvedDestruction) != 1 {
		t.Fatalf("expected 1 resolved destruction (aws_db_instance), got %d", len(comp.ResolvedDestruction))
	}

	if comp.ResolvedDestruction[0].ResourceAddress != "aws_db_instance.customer_orders_postgresql" {
		t.Errorf("expected resolved destruction for customer_orders_postgresql, got %s", comp.ResolvedDestruction[0].ResourceAddress)
	}

	// Blast radius should be reduced
	if comp.BlastRadiusDelta >= 0 {
		t.Errorf("expected negative blast radius delta (reduction), got %+d", comp.BlastRadiusDelta)
	}

	// Security findings resolved: SSH port 22 open + S3 encryption resolved
	if len(comp.ResolvedSecurity) < 2 {
		t.Errorf("expected at least 2 resolved security findings, got %d", len(comp.ResolvedSecurity))
	}

	// Verify markdown output
	md := GenerateMarkdownDiff(comp)
	if !strings.Contains(md, "Resolved Destructive Risks") {
		t.Errorf("expected markdown diff to contain 'Resolved Destructive Risks'")
	}
	if !strings.Contains(md, "aws_db_instance.customer_orders_postgresql") {
		t.Errorf("expected markdown diff to mention customer_orders_postgresql")
	}
}
