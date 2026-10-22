package security

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func init() {
	RegisterRule(Rule{
		ID:       "AWS-SEC-001",
		Title:    "Security Group Allows Open Ingress on Sensitive Ports",
		Severity: model.SeverityCritical,
		Check:    checkAWSSecurityGroupSensitiveIngress,
	})

	RegisterRule(Rule{
		ID:       "AWS-SEC-002",
		Title:    "Security Group Ingress Allows All Traffic From Internet",
		Severity: model.SeverityCritical,
		Check:    checkAWSSecurityGroupAllTrafficIngress,
	})

	RegisterRule(Rule{
		ID:       "AWS-SEC-003",
		Title:    "S3 Bucket Missing Server-Side Encryption",
		Severity: model.SeverityHigh,
		Check:    checkAWSS3Encryption,
	})

	RegisterRule(Rule{
		ID:       "AWS-SEC-004",
		Title:    "RDS Database Publicly Accessible",
		Severity: model.SeverityCritical,
		Check:    checkAWSRDSPublicAccess,
	})

	RegisterRule(Rule{
		ID:       "AWS-SEC-005",
		Title:    "RDS Database Storage Encryption Disabled",
		Severity: model.SeverityHigh,
		Check:    checkAWSRDSEncryption,
	})

	RegisterRule(Rule{
		ID:       "AWS-SEC-006",
		Title:    "IAM Policy Contains Wildcard Action",
		Severity: model.SeverityHigh,
		Check:    checkAWSIAMWildcard,
	})

	RegisterRule(Rule{
		ID:       "AWS-SEC-007",
		Title:    "EBS Volume Encryption Disabled",
		Severity: model.SeverityMedium,
		Check:    checkAWSEBSEncryption,
	})
}

var sensitivePorts = map[int]string{
	22:    "SSH",
	3389:  "RDP",
	5432:  "PostgreSQL",
	3306:  "MySQL",
	1433:  "MSSQL",
	27017: "MongoDB",
	6379:  "Redis",
}

func checkAWSSecurityGroupSensitiveIngress(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "aws_security_group" && node.Type != "aws_security_group_rule" && node.Type != "aws_vpc_security_group_ingress_rule" {
		return nil
	}

	after := node.RawAfter
	if after == nil {
		return nil
	}

	// aws_security_group_rule or aws_vpc_security_group_ingress_rule
	if node.Type == "aws_security_group_rule" || node.Type == "aws_vpc_security_group_ingress_rule" {
		ruleType, _ := after["type"].(string)
		if node.Type == "aws_security_group_rule" && ruleType != "ingress" {
			return nil
		}

		cidrList := extractCIDRs(after)
		if hasOpenCIDR(cidrList) {
			fromPort, toPort := extractPorts(after)
			for p, name := range sensitivePorts {
				if portInRange(p, fromPort, toPort) {
					return &model.SecurityFinding{
						RuleID:      "AWS-SEC-001",
						Title:       fmt.Sprintf("Security Group Ingress Allows Public %s (Port %d)", name, p),
						Severity:    model.SeverityCritical,
						Resource:    node.Address,
						Type:        node.Type,
						Description: fmt.Sprintf("Rule exposes port %d (%s) to 0.0.0.0/0 or ::/0 without IP restriction.", p, name),
						Remediation: "Restrict CIDR blocks to known corporate IP ranges or internal VPC CIDRs.",
						Attribute:   "cidr_blocks",
					}
				}
			}
		}
	}

	// aws_security_group with inline ingress blocks
	if node.Type == "aws_security_group" {
		if ingressRules, ok := after["ingress"].([]interface{}); ok {
			for _, r := range ingressRules {
				rMap, ok := r.(map[string]interface{})
				if !ok {
					continue
				}
				cidrList := extractCIDRs(rMap)
				if hasOpenCIDR(cidrList) {
					fromPort, toPort := extractPorts(rMap)
					for p, name := range sensitivePorts {
						if portInRange(p, fromPort, toPort) {
							return &model.SecurityFinding{
								RuleID:      "AWS-SEC-001",
								Title:       fmt.Sprintf("Security Group Ingress Allows Public %s (Port %d)", name, p),
								Severity:    model.SeverityCritical,
								Resource:    node.Address,
								Type:        node.Type,
								Description: fmt.Sprintf("Security group exposes port %d (%s) to the public internet.", p, name),
								Remediation: "Restrict ingress rules to private IP ranges or reference security groups.",
								Attribute:   "ingress",
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func checkAWSSecurityGroupAllTrafficIngress(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "aws_security_group" && node.Type != "aws_security_group_rule" && node.Type != "aws_vpc_security_group_ingress_rule" {
		return nil
	}

	after := node.RawAfter
	if after == nil {
		return nil
	}

	// aws_security_group_rule / ingress
	if node.Type == "aws_security_group_rule" || node.Type == "aws_vpc_security_group_ingress_rule" {
		cidrList := extractCIDRs(after)
		proto, _ := after["protocol"].(string)
		fromPort, toPort := extractPorts(after)
		if hasOpenCIDR(cidrList) && (proto == "-1" || proto == "all" || (fromPort == 0 && toPort == 65535) || (fromPort == 0 && toPort == 0)) {
			return &model.SecurityFinding{
				RuleID:      "AWS-SEC-002",
				Title:       "Security Group Ingress Allows Unrestricted Inbound Traffic",
				Severity:    model.SeverityCritical,
				Resource:    node.Address,
				Type:        node.Type,
				Description: "Ingress rule permits all protocols and ports from the entire internet (0.0.0.0/0).",
				Remediation: "Limit protocol and ports to specific application requirements and restrict source CIDRs.",
				Attribute:   "protocol",
			}
		}
	}

	return nil
}

func checkAWSS3Encryption(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "aws_s3_bucket" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	// Check if server_side_encryption_configuration is declared inline
	if sse, exists := after["server_side_encryption_configuration"]; exists && sse != nil {
		return nil
	}

	// Check modern acl or bucket
	return &model.SecurityFinding{
		RuleID:      "AWS-SEC-003",
		Title:       "S3 Bucket Missing Server-Side Encryption Configuration",
		Severity:    model.SeverityHigh,
		Resource:    node.Address,
		Type:        node.Type,
		Description: "Bucket does not explicitly configure server_side_encryption_configuration or aws_s3_bucket_server_side_encryption_configuration.",
		Remediation: "Enforce AES256 or AWS KMS encryption via aws_s3_bucket_server_side_encryption_configuration.",
		Attribute:   "server_side_encryption_configuration",
	}
}

func checkAWSRDSPublicAccess(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "aws_db_instance" && node.Type != "aws_rds_cluster_instance" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	if pub, ok := after["publicly_accessible"].(bool); ok && pub {
		return &model.SecurityFinding{
			RuleID:      "AWS-SEC-004",
			Title:       "RDS Database Instance Is Publicly Accessible",
			Severity:    model.SeverityCritical,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "RDS instance has publicly_accessible set to true, assigning a public IP address.",
			Remediation: "Set publicly_accessible = false and deploy the instance in a private subnet.",
			Attribute:   "publicly_accessible",
		}
	}
	return nil
}

func checkAWSRDSEncryption(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "aws_db_instance" && node.Type != "aws_rds_cluster" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	if enc, ok := after["storage_encrypted"].(bool); ok && !enc {
		return &model.SecurityFinding{
			RuleID:      "AWS-SEC-005",
			Title:       "RDS Database Storage Encryption Is Disabled",
			Severity:    model.SeverityHigh,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "Database instance storage is unencrypted at rest.",
			Remediation: "Set storage_encrypted = true and specify a KMS key ID.",
			Attribute:   "storage_encrypted",
		}
	}
	return nil
}

func checkAWSIAMWildcard(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "aws_iam_policy" && node.Type != "aws_iam_role_policy" && node.Type != "aws_iam_user_policy" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	policyStr, ok := after["policy"].(string)
	if !ok || policyStr == "" {
		return nil
	}

	var doc struct {
		Statement []struct {
			Effect   string      `json:"Effect"`
			Action   interface{} `json:"Action"`
			Resource interface{} `json:"Resource"`
		} `json:"Statement"`
	}

	if err := json.Unmarshal([]byte(policyStr), &doc); err != nil {
		return nil
	}

	for _, stmt := range doc.Statement {
		if strings.EqualFold(stmt.Effect, "Allow") {
			if containsWildcard(stmt.Action) {
				return &model.SecurityFinding{
					RuleID:      "AWS-SEC-006",
					Title:       "IAM Policy Grants Wildcard Full Administrative Actions (*)",
					Severity:    model.SeverityHigh,
					Resource:    node.Address,
					Type:        node.Type,
					Description: "IAM policy statement allows Action = '*', granting unrestricted cloud permissions.",
					Remediation: "Apply the principle of least privilege by specifying granular action verbs.",
					Attribute:   "policy",
				}
			}
		}
	}
	return nil
}

func checkAWSEBSEncryption(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "aws_ebs_volume" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	if enc, ok := after["encrypted"].(bool); ok && !enc {
		return &model.SecurityFinding{
			RuleID:      "AWS-SEC-007",
			Title:       "EBS Volume Encryption Is Disabled",
			Severity:    model.SeverityMedium,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "EBS volume block storage is unencrypted at rest.",
			Remediation: "Set encrypted = true on the aws_ebs_volume resource.",
			Attribute:   "encrypted",
		}
	}
	return nil
}

// Helpers

func extractCIDRs(m map[string]interface{}) []string {
	var cidrs []string
	if raw, exists := m["cidr_blocks"]; exists {
		if list, ok := raw.([]interface{}); ok {
			for _, item := range list {
				if s, ok := item.(string); ok {
					cidrs = append(cidrs, s)
				}
			}
		}
	}
	if raw, exists := m["cidr_ipv4"]; exists {
		if s, ok := raw.(string); ok {
			cidrs = append(cidrs, s)
		}
	}
	return cidrs
}

func hasOpenCIDR(cidrs []string) bool {
	for _, c := range cidrs {
		if c == "0.0.0.0/0" || c == "::/0" {
			return true
		}
	}
	return false
}

func extractPorts(m map[string]interface{}) (int, int) {
	from := getInt(m["from_port"])
	to := getInt(m["to_port"])
	return from, to
}

func getInt(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

func portInRange(port, from, to int) bool {
	if from == 0 && to == 0 {
		return true // all ports
	}
	return port >= from && port <= to
}

func containsWildcard(val interface{}) bool {
	switch v := val.(type) {
	case string:
		return v == "*"
	case []interface{}:
		for _, item := range v {
			if item == "*" {
				return true
			}
		}
	}
	return false
}
