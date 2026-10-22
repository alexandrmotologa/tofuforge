package parser

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/alexandrmotologa/tofuforge/internal/model"
)

// ParsePlan decodes a Terraform or OpenTofu plan JSON from a reader.
func ParsePlan(r io.Reader) (*model.RawPlan, error) {
	decoder := json.NewDecoder(r)
	var raw model.RawPlan
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to parse plan json: %w", err)
	}

	if raw.FormatVersion == "" && len(raw.ResourceChanges) == 0 && raw.PlannedValues == nil {
		return nil, fmt.Errorf("invalid plan format: missing format_version and resource changes")
	}

	return &raw, nil
}

// NormalizeProvider cleans up long provider strings (e.g., "registry.terraform.io/hashicorp/aws" -> "aws").
func NormalizeProvider(providerName string) string {
	if providerName == "" {
		return "unknown"
	}
	parts := strings.Split(providerName, "/")
	last := parts[len(parts)-1]
	// strip prefix if provider is hashicorp-aws or similar
	return strings.TrimPrefix(last, "terraform-provider-")
}

// ExtractModuleAddress returns the parent module path for a resource address.
func ExtractModuleAddress(address string) string {
	parts := strings.Split(address, ".")
	if len(parts) > 2 && parts[0] == "module" {
		// e.g. module.vpc.aws_subnet.private -> module.vpc
		for i := len(parts) - 1; i >= 0; i-- {
			if parts[i] == "module" {
				return strings.Join(parts[:i+2], ".")
			}
		}
	}
	return "root"
}
