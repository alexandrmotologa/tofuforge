package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents TofuForge operational and security governance settings.
type Config struct {
	Version         int             `yaml:"version" json:"version"`
	Rules           RulesConfig     `yaml:"rules" json:"rules"`
	Strictness      StrictnessConfig `yaml:"strictness" json:"strictness"`
	TagsEnforcement TagsConfig      `yaml:"tags_enforcement" json:"tags_enforcement"`
}

type RulesConfig struct {
	Ignore []string `yaml:"ignore" json:"ignore"`
}

type StrictnessConfig struct {
	FailOnDestructive bool    `yaml:"fail_on_destructive" json:"fail_on_destructive"`
	FailOnCritical    bool    `yaml:"fail_on_critical" json:"fail_on_critical"`
	FailOnSecurity    bool    `yaml:"fail_on_security" json:"fail_on_security"`
	MaxBlastRadius    int     `yaml:"max_blast_radius" json:"max_blast_radius"`
	MaxCostDelta      float64 `yaml:"max_cost_delta" json:"max_cost_delta"`
}

type TagsConfig struct {
	Required []string `yaml:"required" json:"required"`
}

// DefaultConfig returns an initialized default config with no rule ignores or forced tags.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Rules: RulesConfig{
			Ignore: []string{},
		},
		Strictness: StrictnessConfig{
			FailOnDestructive: false,
			FailOnCritical:    false,
			FailOnSecurity:    false,
			MaxBlastRadius:    0,
			MaxCostDelta:      0,
		},
		TagsEnforcement: TagsConfig{
			Required: []string{},
		},
	}
}

// LoadConfig loads configuration from path or default filenames (.tofuforge.yaml, .tofuforge.yml).
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	candidatePath := path
	if candidatePath == "" {
		if _, err := os.Stat(".tofuforge.yaml"); err == nil {
			candidatePath = ".tofuforge.yaml"
		} else if _, err := os.Stat(".tofuforge.yml"); err == nil {
			candidatePath = ".tofuforge.yml"
		} else {
			return cfg, nil // clean fallback to defaults
		}
	}

	data, err := os.ReadFile(candidatePath)
	if err != nil {
		if path != "" {
			return nil, fmt.Errorf("failed to read config file '%s': %w", path, err)
		}
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid YAML in '%s': %w", candidatePath, err)
	}

	return cfg, nil
}
