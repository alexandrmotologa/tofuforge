package security

import (
	"fmt"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func init() {
	RegisterRule(Rule{
		ID:       "AZR-SEC-001",
		Title:    "Network Security Group Inbound Allows Internet to Sensitive Ports",
		Severity: model.SeverityCritical,
		Check:    checkAzureNSGSensitiveIngress,
	})

	RegisterRule(Rule{
		ID:       "AZR-SEC-002",
		Title:    "Storage Account Allows Insecure HTTP Traffic",
		Severity: model.SeverityHigh,
		Check:    checkAzureStorageHTTPSOnly,
	})

	RegisterRule(Rule{
		ID:       "AZR-SEC-003",
		Title:    "Storage Account Allows Public Blob Access",
		Severity: model.SeverityHigh,
		Check:    checkAzureStoragePublicBlob,
	})

	RegisterRule(Rule{
		ID:       "AZR-SEC-004",
		Title:    "MSSQL Database Transparent Data Encryption Disabled",
		Severity: model.SeverityHigh,
		Check:    checkAzureMSSQLTDE,
	})
}

func checkAzureNSGSensitiveIngress(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "azurerm_network_security_group" && node.Type != "azurerm_network_security_rule" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	checkRule := func(r map[string]interface{}) *model.SecurityFinding {
		dir, _ := r["direction"].(string)
		access, _ := r["access"].(string)
		if !strings.EqualFold(dir, "Inbound") || !strings.EqualFold(access, "Allow") {
			return nil
		}

		sourcePrefix, _ := r["source_address_prefix"].(string)
		isInternet := sourcePrefix == "*" || strings.EqualFold(sourcePrefix, "Internet") || sourcePrefix == "0.0.0.0/0"

		if !isInternet {
			if prefixes, ok := r["source_address_prefixes"].([]interface{}); ok {
				for _, p := range prefixes {
					pStr, _ := p.(string)
					if pStr == "*" || strings.EqualFold(pStr, "Internet") || pStr == "0.0.0.0/0" {
						isInternet = true
						break
					}
				}
			}
		}

		if !isInternet {
			return nil
		}

		destPort, _ := r["destination_port_range"].(string)
		for p, name := range sensitivePorts {
			if destPort == "*" || destPort == fmt.Sprintf("%d", p) {
				return &model.SecurityFinding{
					RuleID:      "AZR-SEC-001",
					Title:       fmt.Sprintf("Azure NSG Allows Public Inbound %s (Port %d)", name, p),
					Severity:    model.SeverityCritical,
					Resource:    node.Address,
					Type:        node.Type,
					Description: fmt.Sprintf("Inbound security rule permits internet traffic to sensitive port %d (%s).", p, name),
					Remediation: "Restrict source_address_prefix to specific corporate Virtual Networks or IP blocks.",
					Attribute:   "destination_port_range",
				}
			}
		}

		return nil
	}

	if node.Type == "azurerm_network_security_rule" {
		return checkRule(after)
	}

	if secRules, ok := after["security_rule"].([]interface{}); ok {
		for _, sr := range secRules {
			if srMap, ok := sr.(map[string]interface{}); ok {
				if f := checkRule(srMap); f != nil {
					return f
				}
			}
		}
	}

	return nil
}

func checkAzureStorageHTTPSOnly(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "azurerm_storage_account" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	if https, ok := after["enable_https_traffic_only"].(bool); ok && !https {
		return &model.SecurityFinding{
			RuleID:      "AZR-SEC-002",
			Title:       "Azure Storage Account Insecure HTTP Traffic Allowed",
			Severity:    model.SeverityHigh,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "enable_https_traffic_only is set to false, allowing unencrypted HTTP transfers.",
			Remediation: "Set enable_https_traffic_only = true to enforce TLS transfers.",
			Attribute:   "enable_https_traffic_only",
		}
	}
	return nil
}

func checkAzureStoragePublicBlob(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "azurerm_storage_account" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	if pub, ok := after["allow_nested_items_to_be_public"].(bool); ok && pub {
		return &model.SecurityFinding{
			RuleID:      "AZR-SEC-003",
			Title:       "Azure Storage Account Public Blob Access Enabled",
			Severity:    model.SeverityHigh,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "allow_nested_items_to_be_public or allow_blob_public_access is enabled.",
			Remediation: "Set allow_nested_items_to_be_public = false.",
			Attribute:   "allow_nested_items_to_be_public",
		}
	}

	if pub, ok := after["allow_blob_public_access"].(bool); ok && pub {
		return &model.SecurityFinding{
			RuleID:      "AZR-SEC-003",
			Title:       "Azure Storage Account Public Blob Access Enabled",
			Severity:    model.SeverityHigh,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "allow_blob_public_access is set to true.",
			Remediation: "Set allow_blob_public_access = false.",
			Attribute:   "allow_blob_public_access",
		}
	}

	return nil
}

func checkAzureMSSQLTDE(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "azurerm_mssql_database" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	if tde, ok := after["transparent_data_encryption_enabled"].(bool); ok && !tde {
		return &model.SecurityFinding{
			RuleID:      "AZR-SEC-004",
			Title:       "Azure MSSQL Database Transparent Data Encryption Disabled",
			Severity:    model.SeverityHigh,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "Database encryption at rest via TDE is turned off.",
			Remediation: "Set transparent_data_encryption_enabled = true.",
			Attribute:   "transparent_data_encryption_enabled",
		}
	}

	return nil
}
