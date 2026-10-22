package security

import (
	"testing"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func TestSecurityLinter_AWSRules(t *testing.T) {
	nodes := map[string]*model.ResourceNode{
		"aws_security_group_rule.bad_ssh": {
			Address:  "aws_security_group_rule.bad_ssh",
			Type:     "aws_security_group_rule",
			Action:   model.ActionCreate,
			RawAfter: map[string]interface{}{
				"type":        "ingress",
				"from_port":   22,
				"to_port":     22,
				"cidr_blocks": []interface{}{"0.0.0.0/0"},
			},
		},
		"aws_s3_bucket.unencrypted": {
			Address:  "aws_s3_bucket.unencrypted",
			Type:     "aws_s3_bucket",
			Action:   model.ActionCreate,
			RawAfter: map[string]interface{}{
				"bucket": "customer-data",
			},
		},
		"aws_db_instance.public_db": {
			Address:  "aws_db_instance.public_db",
			Type:     "aws_db_instance",
			Action:   model.ActionUpdate,
			RawAfter: map[string]interface{}{
				"publicly_accessible": true,
				"storage_encrypted":   false,
			},
		},
	}

	findings := Lint(nodes)

	if len(findings) < 4 {
		t.Fatalf("expected at least 4 security findings, got %d", len(findings))
	}

	ruleMap := make(map[string]bool)
	for _, f := range findings {
		ruleMap[f.RuleID] = true
	}

	expectedRules := []string{"AWS-SEC-001", "AWS-SEC-003", "AWS-SEC-004", "AWS-SEC-005"}
	for _, r := range expectedRules {
		if !ruleMap[r] {
			t.Errorf("expected rule %s to be triggered, but was missing", r)
		}
	}
}

func TestSecurityLinter_GCPAndAzureRules(t *testing.T) {
	nodes := map[string]*model.ResourceNode{
		"google_compute_firewall.open": {
			Address: "google_compute_firewall.open",
			Type:    "google_compute_firewall",
			Action:  model.ActionCreate,
			RawAfter: map[string]interface{}{
				"direction":     "INGRESS",
				"source_ranges": []interface{}{"0.0.0.0/0"},
				"allow": []interface{}{
					map[string]interface{}{
						"protocol": "tcp",
						"ports":    []interface{}{"3389"},
					},
				},
			},
		},
		"azurerm_network_security_rule.bad_rdp": {
			Address: "azurerm_network_security_rule.bad_rdp",
			Type:    "azurerm_network_security_rule",
			Action:  model.ActionCreate,
			RawAfter: map[string]interface{}{
				"direction":              "Inbound",
				"access":                 "Allow",
				"source_address_prefix":  "*",
				"destination_port_range": "3389",
			},
		},
	}

	findings := Lint(nodes)

	if len(findings) != 2 {
		t.Fatalf("expected 2 findings for GCP and Azure, got %d", len(findings))
	}
}

func TestSecurityLinter_KubernetesRules(t *testing.T) {
	nodes := map[string]*model.ResourceNode{
		"kubernetes_pod.root": {
			Address: "kubernetes_pod.root",
			Type:    "kubernetes_pod",
			Action:  model.ActionCreate,
			RawAfter: map[string]interface{}{
				"spec": map[string]interface{}{
					"container": map[string]interface{}{
						"security_context": map[string]interface{}{
							"privileged":      true,
							"run_as_non_root": false,
						},
					},
				},
			},
		},
	}

	findings := Lint(nodes)
	if len(findings) < 2 {
		t.Fatalf("expected at least 2 findings (privileged, root), got %d", len(findings))
	}
}
