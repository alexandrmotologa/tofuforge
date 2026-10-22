package parser

import (
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// ClassifyCategory determines the functional domain of a resource type.
func ClassifyCategory(resourceType string) model.ResourceCategory {
	rt := strings.ToLower(resourceType)

	// Database
	if strings.Contains(rt, "db_") || strings.Contains(rt, "rds_") ||
		strings.Contains(rt, "dynamodb") || strings.Contains(rt, "sql") ||
		strings.Contains(rt, "elasticache") || strings.Contains(rt, "redis") ||
		strings.Contains(rt, "spanner") || strings.Contains(rt, "bigtable") ||
		strings.Contains(rt, "cosmosdb") || strings.Contains(rt, "postgresql") ||
		strings.Contains(rt, "mongodb") || strings.Contains(rt, "opensearch") ||
		strings.Contains(rt, "elasticsearch") || strings.Contains(rt, "redshift") {
		return model.CategoryDatabase
	}

	// Storage
	if strings.Contains(rt, "s3_") || strings.Contains(rt, "storage_") ||
		strings.Contains(rt, "ebs_") || strings.Contains(rt, "disk") ||
		strings.Contains(rt, "efs_") || strings.Contains(rt, "filestore") ||
		strings.Contains(rt, "persistent_volume") {
		return model.CategoryStorage
	}

	// Network
	if strings.Contains(rt, "vpc") || strings.Contains(rt, "subnet") ||
		strings.Contains(rt, "network") || strings.Contains(rt, "router") ||
		strings.Contains(rt, "gateway") || strings.Contains(rt, "route_table") ||
		strings.Contains(rt, "peering") || strings.Contains(rt, "load_balancer") ||
		strings.Contains(rt, "lb") || strings.Contains(rt, "alb") || strings.Contains(rt, "nlb") ||
		strings.Contains(rt, "firewall") || strings.Contains(rt, "security_group") ||
		strings.Contains(rt, "nsg") {
		return model.CategoryNetwork
	}

	// Compute
	if strings.Contains(rt, "instance") || strings.Contains(rt, "virtual_machine") ||
		strings.Contains(rt, "container") || strings.Contains(rt, "ecs_") ||
		strings.Contains(rt, "eks_") || strings.Contains(rt, "gke_") ||
		strings.Contains(rt, "aks_") || strings.Contains(rt, "lambda_") ||
		strings.Contains(rt, "cloud_run") || strings.Contains(rt, "function") ||
		strings.Contains(rt, "pod") || strings.Contains(rt, "deployment") {
		return model.CategoryCompute
	}

	// Security & IAM
	if strings.Contains(rt, "iam_") || strings.Contains(rt, "kms_") ||
		strings.Contains(rt, "secret") || strings.Contains(rt, "vault") ||
		strings.Contains(rt, "certificate") || strings.Contains(rt, "key_vault") ||
		strings.Contains(rt, "policy") || strings.Contains(rt, "role") {
		return model.CategorySecurity
	}

	// DNS
	if strings.Contains(rt, "dns") || strings.Contains(rt, "route53") {
		return model.CategoryDNS
	}

	return model.CategoryOther
}

// IsStatefulResource determines whether a resource holds critical persistent data
// or foundational network state where recreate or destroy would cause catastrophic outages.
func IsStatefulResource(resourceType string) bool {
	rt := strings.ToLower(resourceType)

	// All databases are stateful
	cat := ClassifyCategory(resourceType)
	if cat == model.CategoryDatabase {
		return true
	}

	// Persistent volumes and object buckets
	if strings.Contains(rt, "s3_bucket") ||
		strings.Contains(rt, "storage_bucket") ||
		strings.Contains(rt, "storage_account") ||
		strings.Contains(rt, "persistent_volume") ||
		strings.Contains(rt, "ebs_volume") ||
		strings.Contains(rt, "compute_disk") ||
		strings.Contains(rt, "managed_disk") {
		return true
	}

	// Foundational networking & encryption keys
	if strings.Contains(rt, "aws_vpc") ||
		strings.Contains(rt, "google_compute_network") ||
		strings.Contains(rt, "azurerm_virtual_network") ||
		strings.Contains(rt, "kms_key") ||
		strings.Contains(rt, "kms_crypto_key") ||
		strings.Contains(rt, "key_vault_key") {
		return true
	}

	return false
}
