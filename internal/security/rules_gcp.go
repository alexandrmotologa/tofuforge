package security

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func init() {
	RegisterRule(Rule{
		ID:       "GCP-SEC-001",
		Title:    "GCP Firewall Allows Public Ingress on Sensitive Ports",
		Severity: model.SeverityCritical,
		Check:    checkGCPFirewallSensitiveIngress,
	})

	RegisterRule(Rule{
		ID:       "GCP-SEC-002",
		Title:    "Cloud SQL Instance Has Public IP With Open Ingress",
		Severity: model.SeverityCritical,
		Check:    checkGCPCloudSQLPublicIP,
	})

	RegisterRule(Rule{
		ID:       "GCP-SEC-003",
		Title:    "IAM Policy Grants Role to All Public Users",
		Severity: model.SeverityCritical,
		Check:    checkGCPIAMPublicMembers,
	})

	RegisterRule(Rule{
		ID:       "GCP-SEC-004",
		Title:    "Cloud Storage Uniform Bucket-Level Access Disabled",
		Severity: model.SeverityMedium,
		Check:    checkGCPStorageUniformAccess,
	})
}

func checkGCPFirewallSensitiveIngress(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "google_compute_firewall" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	dir, _ := after["direction"].(string)
	if dir != "" && !strings.EqualFold(dir, "INGRESS") {
		return nil
	}

	// Check source_ranges contains 0.0.0.0/0
	ranges, ok := after["source_ranges"].([]interface{})
	if !ok {
		return nil
	}

	hasOpenSource := false
	for _, r := range ranges {
		if rStr, ok := r.(string); ok && (rStr == "0.0.0.0/0" || rStr == "::/0") {
			hasOpenSource = true
			break
		}
	}
	if !hasOpenSource {
		return nil
	}

	// Check allowed ports
	if allows, ok := after["allow"].([]interface{}); ok {
		for _, a := range allows {
			aMap, ok := a.(map[string]interface{})
			if !ok {
				continue
			}
			proto, _ := aMap["protocol"].(string)
			if proto == "all" {
				return &model.SecurityFinding{
					RuleID:      "GCP-SEC-001",
					Title:       "GCP Firewall Allows All Public Inbound Traffic",
					Severity:    model.SeverityCritical,
					Resource:    node.Address,
					Type:        node.Type,
					Description: "Firewall rule allows protocol 'all' from 0.0.0.0/0.",
					Remediation: "Restrict allowed protocols and target tags.",
					Attribute:   "allow",
				}
			}

			if ports, ok := aMap["ports"].([]interface{}); ok {
				for _, p := range ports {
					pStr := fmt.Sprintf("%v", p)
					for sPort, sName := range sensitivePorts {
						if pStr == fmt.Sprintf("%d", sPort) {
							return &model.SecurityFinding{
								RuleID:      "GCP-SEC-001",
								Title:       fmt.Sprintf("GCP Firewall Exposes Public %s (Port %d)", sName, sPort),
								Severity:    model.SeverityCritical,
								Resource:    node.Address,
								Type:        node.Type,
								Description: fmt.Sprintf("Firewall rule permits 0.0.0.0/0 ingress to port %d (%s).", sPort, sName),
								Remediation: "Scope source ranges to designated bastion or internal subnets.",
								Attribute:   "allow.ports",
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func checkGCPCloudSQLPublicIP(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "google_sql_database_instance" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	settings, ok := after["settings"].([]interface{})
	if !ok || len(settings) == 0 {
		return nil
	}

	settingMap, ok := settings[0].(map[string]interface{})
	if !ok {
		return nil
	}

	ipConfig, ok := settingMap["ip_configuration"].([]interface{})
	if !ok || len(ipConfig) == 0 {
		return nil
	}

	ipMap, ok := ipConfig[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if ipv4, ok := ipMap["ipv4_enabled"].(bool); ok && ipv4 {
		// Check authorized networks
		if authNets, ok := ipMap["authorized_networks"].([]interface{}); ok {
			for _, an := range authNets {
				anMap, ok := an.(map[string]interface{})
				if !ok {
					continue
				}
				if val, ok := anMap["value"].(string); ok && (val == "0.0.0.0/0" || val == "::/0") {
					return &model.SecurityFinding{
						RuleID:      "GCP-SEC-002",
						Title:       "Cloud SQL Instance Has Public IPv4 With 0.0.0.0/0 Authorized Network",
						Severity:    model.SeverityCritical,
						Resource:    node.Address,
						Type:        node.Type,
						Description: "Database instance is configured with a public IPv4 address and accepts connections from any internet IP.",
						Remediation: "Use private IP with Cloud SQL Auth Proxy or Service Networking.",
						Attribute:   "settings.ip_configuration",
					}
				}
			}
		}
	}

	return nil
}

func checkGCPIAMPublicMembers(node *model.ResourceNode) *model.SecurityFinding {
	if !strings.Contains(node.Type, "google_") || !strings.Contains(node.Type, "_iam_") {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	// Check member attribute
	if mem, ok := after["member"].(string); ok {
		if mem == "allUsers" || mem == "allAuthenticatedUsers" {
			return &model.SecurityFinding{
				RuleID:      "GCP-SEC-003",
				Title:       "GCP IAM Grants Role To Public Internet (allUsers)",
				Severity:    model.SeverityCritical,
				Resource:    node.Address,
				Type:        node.Type,
				Description: fmt.Sprintf("IAM assignment grants access to %s.", mem),
				Remediation: "Remove public identities from IAM role bindings.",
				Attribute:   "member",
			}
		}
	}

	// Check members list
	if mems, ok := after["members"].([]interface{}); ok {
		for _, m := range mems {
			if mStr, ok := m.(string); ok && (mStr == "allUsers" || mStr == "allAuthenticatedUsers") {
				return &model.SecurityFinding{
					RuleID:      "GCP-SEC-003",
					Title:       "GCP IAM Grants Role To Public Internet (allUsers)",
					Severity:    model.SeverityCritical,
					Resource:    node.Address,
					Type:        node.Type,
					Description: fmt.Sprintf("IAM assignment grants access to %s.", mStr),
					Remediation: "Remove public identities from IAM role bindings.",
					Attribute:   "members",
				}
			}
		}
	}

	return nil
}

func checkGCPStorageUniformAccess(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "google_storage_bucket" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	if ubla, ok := after["uniform_bucket_level_access"].(bool); ok && !ubla {
		return &model.SecurityFinding{
			RuleID:      "GCP-SEC-004",
			Title:       "Cloud Storage Uniform Bucket-Level Access Disabled",
			Severity:    model.SeverityMedium,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "Uniform bucket-level access is disabled, allowing object-level ACL overrides.",
			Remediation: "Set uniform_bucket_level_access = true for consolidated IAM governance.",
			Attribute:   "uniform_bucket_level_access",
		}
	}

	return nil
}
