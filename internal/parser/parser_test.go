package parser

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func TestClassifyAction(t *testing.T) {
	tests := []struct {
		actions  []string
		expected model.Action
	}{
		{[]string{"no-op"}, model.ActionNoOp},
		{[]string{"create"}, model.ActionCreate},
		{[]string{"read"}, model.ActionRead},
		{[]string{"update"}, model.ActionUpdate},
		{[]string{"delete"}, model.ActionDelete},
		{[]string{"delete", "create"}, model.ActionReplace},
		{[]string{"create", "delete"}, model.ActionReplace},
		{[]string{}, model.ActionNoOp},
	}

	for _, tt := range tests {
		got := ClassifyAction(tt.actions)
		if got != tt.expected {
			t.Errorf("ClassifyAction(%v) = %v; want %v", tt.actions, got, tt.expected)
		}
	}
}

func TestIsDestructive(t *testing.T) {
	if !IsDestructive(model.ActionDelete) {
		t.Errorf("expected ActionDelete to be destructive")
	}
	if !IsDestructive(model.ActionReplace) {
		t.Errorf("expected ActionReplace to be destructive")
	}
	if IsDestructive(model.ActionCreate) {
		t.Errorf("expected ActionCreate NOT to be destructive")
	}
	if IsDestructive(model.ActionUpdate) {
		t.Errorf("expected ActionUpdate NOT to be destructive")
	}
}

func TestClassifyCategoryAndStatefulness(t *testing.T) {
	tests := []struct {
		resType    string
		expected   model.ResourceCategory
		isStateful bool
	}{
		{"aws_db_instance", model.CategoryDatabase, true},
		{"aws_s3_bucket", model.CategoryStorage, true},
		{"aws_vpc", model.CategoryNetwork, true},
		{"aws_security_group", model.CategoryNetwork, false},
		{"aws_instance", model.CategoryCompute, false},
		{"aws_iam_role", model.CategorySecurity, false},
		{"aws_route53_zone", model.CategoryDNS, false},
		{"google_sql_database_instance", model.CategoryDatabase, true},
		{"google_storage_bucket", model.CategoryStorage, true},
		{"azurerm_mssql_database", model.CategoryDatabase, true},
	}

	for _, tt := range tests {
		cat := ClassifyCategory(tt.resType)
		if cat != tt.expected {
			t.Errorf("ClassifyCategory(%s) = %v; want %v", tt.resType, cat, tt.expected)
		}
		stateful := IsStatefulResource(tt.resType)
		if stateful != tt.isStateful {
			t.Errorf("IsStatefulResource(%s) = %v; want %v", tt.resType, stateful, tt.isStateful)
		}
	}
}

func TestComputeAttributeDiffs(t *testing.T) {
	change := model.RawChange{
		Before: map[string]interface{}{
			"allocated_storage": 20,
			"instance_class":    "db.t3.micro",
			"identifier":        "prod-db",
		},
		After: map[string]interface{}{
			"allocated_storage": 50,
			"instance_class":    "db.t3.medium",
			"identifier":        "prod-db-v2",
			"deletion_protection": true,
		},
		ReplacePaths: [][]interface{}{
			{"identifier"},
		},
	}

	diffs := ComputeAttributeDiffs(change)
	if len(diffs) != 4 {
		t.Fatalf("expected 4 diffs, got %d", len(diffs))
	}

	// First diff should be the one forcing replacement
	if !diffs[0].ForcesReplacement || diffs[0].Path != "identifier" {
		t.Errorf("expected first diff to be replacement-forcing 'identifier', got %+v", diffs[0])
	}
}

func TestParsePlan(t *testing.T) {
	planJSON := `{
		"format_version": "1.2",
		"terraform_version": "1.8.0",
		"resource_changes": [
			{
				"address": "aws_s3_bucket.data",
				"type": "aws_s3_bucket",
				"name": "data",
				"provider_name": "registry.terraform.io/hashicorp/aws",
				"change": {
					"actions": ["create"],
					"after": {
						"bucket": "my-company-assets"
					}
				}
			}
		]
	}`

	plan, err := ParsePlan(strings.NewReader(planJSON))
	if err != nil {
		t.Fatalf("ParsePlan failed: %v", err)
	}

	if plan.FormatVersion != "1.2" {
		t.Errorf("expected format_version 1.2, got %s", plan.FormatVersion)
	}
	if len(plan.ResourceChanges) != 1 {
		t.Fatalf("expected 1 resource change, got %d", len(plan.ResourceChanges))
	}
	if NormalizeProvider(plan.ResourceChanges[0].ProviderName) != "aws" {
		t.Errorf("expected normalized provider 'aws', got %s", NormalizeProvider(plan.ResourceChanges[0].ProviderName))
	}
}
