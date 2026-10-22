package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatalf("unexpected error loading default config: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("expected version 1, got %d", cfg.Version)
	}
	if len(cfg.Rules.Ignore) != 0 {
		t.Errorf("expected 0 ignored rules by default")
	}
}

func TestLoadConfig_ValidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, ".tofuforge.yaml")

	content := `
version: 1
rules:
  ignore:
    - "AWS-SEC-003"
strictness:
  fail_on_critical: true
  max_blast_radius: 40
  max_cost_delta: 250.0
tags_enforcement:
  required:
    - "Environment"
    - "Owner"
`

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("failed to load custom config: %v", err)
	}

	if len(cfg.Rules.Ignore) != 1 || cfg.Rules.Ignore[0] != "AWS-SEC-003" {
		t.Errorf("expected AWS-SEC-003 in ignore list, got %v", cfg.Rules.Ignore)
	}
	if !cfg.Strictness.FailOnCritical {
		t.Errorf("expected FailOnCritical to be true")
	}
	if cfg.Strictness.MaxBlastRadius != 40 {
		t.Errorf("expected MaxBlastRadius to be 40, got %d", cfg.Strictness.MaxBlastRadius)
	}
	if len(cfg.TagsEnforcement.Required) != 2 || cfg.TagsEnforcement.Required[0] != "Environment" {
		t.Errorf("expected Environment and Owner in required tags, got %v", cfg.TagsEnforcement.Required)
	}
}
