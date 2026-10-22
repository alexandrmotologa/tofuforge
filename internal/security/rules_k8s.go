package security

import (
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

func init() {
	RegisterRule(Rule{
		ID:       "K8S-SEC-001",
		Title:    "Kubernetes Container Configured in Privileged Mode",
		Severity: model.SeverityCritical,
		Check:    checkK8sPrivilegedContainer,
	})

	RegisterRule(Rule{
		ID:       "K8S-SEC-002",
		Title:    "Kubernetes Container Permitted to Run as Root",
		Severity: model.SeverityHigh,
		Check:    checkK8sRunAsRoot,
	})

	RegisterRule(Rule{
		ID:       "K8S-SEC-003",
		Title:    "Kubernetes Container Missing Resource Limits",
		Severity: model.SeverityMedium,
		Check:    checkK8sResourceLimits,
	})
}

func checkK8sPrivilegedContainer(node *model.ResourceNode) *model.SecurityFinding {
	if !strings.HasPrefix(node.Type, "kubernetes_") {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	// Scan nested map for security_context.privileged = true
	if isMapFlagTrue(after, "privileged") {
		return &model.SecurityFinding{
			RuleID:      "K8S-SEC-001",
			Title:       "Kubernetes Container Runs With Full Privileged Access",
			Severity:    model.SeverityCritical,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "Container has privileged = true, giving host-level access and bypassing container isolation.",
			Remediation: "Remove privileged flag and declare minimal required Linux capabilities instead.",
			Attribute:   "security_context.privileged",
		}
	}

	return nil
}

func checkK8sRunAsRoot(node *model.ResourceNode) *model.SecurityFinding {
	if !strings.HasPrefix(node.Type, "kubernetes_") {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	if isMapFlagFalse(after, "run_as_non_root") {
		return &model.SecurityFinding{
			RuleID:      "K8S-SEC-002",
			Title:       "Kubernetes Container Does Not Enforce Non-Root Execution",
			Severity:    model.SeverityHigh,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "run_as_non_root is set to false, allowing containers to execute processes as root.",
			Remediation: "Set run_as_non_root = true and specify run_as_user with a non-zero UID.",
			Attribute:   "security_context.run_as_non_root",
		}
	}

	return nil
}

func checkK8sResourceLimits(node *model.ResourceNode) *model.SecurityFinding {
	if node.Type != "kubernetes_deployment" && node.Type != "kubernetes_pod" && node.Type != "kubernetes_stateful_set" {
		return nil
	}
	after := node.RawAfter
	if after == nil {
		return nil
	}

	// Check if resources.limits is present anywhere
	if !hasKeyDeep(after, "limits") {
		return &model.SecurityFinding{
			RuleID:      "K8S-SEC-003",
			Title:       "Kubernetes Workload Missing CPU or Memory Resource Limits",
			Severity:    model.SeverityMedium,
			Resource:    node.Address,
			Type:        node.Type,
			Description: "Containers do not define compute resource limits, risking node memory exhaustion or noisy neighbors.",
			Remediation: "Configure resources.limits.cpu and resources.limits.memory.",
			Attribute:   "spec.template.spec.container.resources.limits",
		}
	}

	return nil
}

func isMapFlagTrue(m map[string]interface{}, key string) bool {
	for k, v := range m {
		if k == key {
			if b, ok := v.(bool); ok && b {
				return true
			}
		}
		if sub, ok := v.(map[string]interface{}); ok {
			if isMapFlagTrue(sub, key) {
				return true
			}
		}
		if list, ok := v.([]interface{}); ok {
			for _, item := range list {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if isMapFlagTrue(itemMap, key) {
						return true
					}
				}
			}
		}
	}
	return false
}

func isMapFlagFalse(m map[string]interface{}, key string) bool {
	for k, v := range m {
		if k == key {
			if b, ok := v.(bool); ok && !b {
				return true
			}
		}
		if sub, ok := v.(map[string]interface{}); ok {
			if isMapFlagFalse(sub, key) {
				return true
			}
		}
		if list, ok := v.([]interface{}); ok {
			for _, item := range list {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if isMapFlagFalse(itemMap, key) {
						return true
					}
				}
			}
		}
	}
	return false
}

func hasKeyDeep(m map[string]interface{}, targetKey string) bool {
	for k, v := range m {
		if k == targetKey {
			return true
		}
		if sub, ok := v.(map[string]interface{}); ok {
			if hasKeyDeep(sub, targetKey) {
				return true
			}
		}
		if list, ok := v.([]interface{}); ok {
			for _, item := range list {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if hasKeyDeep(itemMap, targetKey) {
						return true
					}
				}
			}
		}
	}
	return false
}
