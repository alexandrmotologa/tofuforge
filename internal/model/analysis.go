package model

// Severity defines risk severity for security findings and destruction alerts.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityHigh     Severity = "HIGH"
	SeverityMedium   Severity = "MEDIUM"
	SeverityLow      Severity = "LOW"
	SeverityInfo     Severity = "INFO"
)

// CostEstimate represents estimated monthly infrastructure expenses.
type CostEstimate struct {
	MonthlyBefore float64 `json:"monthly_before"`
	MonthlyAfter  float64 `json:"monthly_after"`
	MonthlyDelta  float64 `json:"monthly_delta"`
	Currency      string  `json:"currency"`
	Description   string  `json:"description,omitempty"`
}

// ResourceNode represents an evaluated resource in the dependency DAG.
type ResourceNode struct {
	ID                string            `json:"id"`
	Address           string            `json:"address"`
	Name              string            `json:"name"`
	Type              string            `json:"type"`
	Provider          string            `json:"provider"`
	Module            string            `json:"module"`
	Mode              string            `json:"mode"`
	Action            Action            `json:"action"`
	Category          ResourceCategory  `json:"category"`
	IsStateful        bool              `json:"is_stateful"`
	BlastRadiusScore  int               `json:"blast_radius_score"`
	DirectImpact      bool              `json:"direct_impact"`
	DownstreamCount   int               `json:"downstream_count"`
	DownstreamIDs     []string          `json:"downstream_ids,omitempty"`
	UpstreamIDs       []string          `json:"upstream_ids,omitempty"`
	Dependencies      []string          `json:"dependencies,omitempty"`
	AttributeDiffs    []AttributeDiff   `json:"attribute_diffs,omitempty"`
	SecurityFindings  []SecurityFinding `json:"security_findings,omitempty"`
	Cost              *CostEstimate     `json:"cost,omitempty"`
	RawBefore         map[string]interface{} `json:"raw_before,omitempty"`
	RawAfter          map[string]interface{} `json:"raw_after,omitempty"`
}

// DependencyEdge represents a directed dependency link (Source -> Target).
// Target depends on Source (Source must be created/present before Target).
type DependencyEdge struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Target   string `json:"target"`
	Implicit bool   `json:"implicit"`
	Reason   string `json:"reason,omitempty"`
}

// DestructionAlert represents a critical warning for resources scheduled for deletion or recreation.
type DestructionAlert struct {
	ResourceAddress string           `json:"resource_address"`
	ResourceType    string           `json:"resource_type"`
	Category        ResourceCategory `json:"category"`
	Action          Action           `json:"action"`
	Severity        Severity         `json:"severity"`
	IsStateful      bool             `json:"is_stateful"`
	BlastRadius     int              `json:"blast_radius"`
	DownstreamCount int              `json:"downstream_count"`
	DownstreamPaths []string         `json:"downstream_paths,omitempty"`
	Message         string           `json:"message"`
	ReplacementKeys []string         `json:"replacement_keys,omitempty"`
}

// SecurityFinding describes an identified security violation or posture flaw.
type SecurityFinding struct {
	RuleID      string   `json:"rule_id"`
	Title       string   `json:"title"`
	Severity    Severity `json:"severity"`
	Resource    string   `json:"resource"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Remediation string   `json:"remediation"`
	Attribute   string   `json:"attribute,omitempty"`
}

// SummaryStats aggregates plan metrics.
type SummaryStats struct {
	TotalResources       int     `json:"total_resources"`
	CreateCount          int     `json:"create_count"`
	UpdateCount          int     `json:"update_count"`
	DeleteCount          int     `json:"delete_count"`
	ReplaceCount         int     `json:"replace_count"`
	NoOpCount            int     `json:"no_op_count"`
	ReadCount            int     `json:"read_count"`
	TotalBlastRadius     int     `json:"total_blast_radius"`
	CriticalAlertsCount  int     `json:"critical_alerts_count"`
	SecurityIssuesCount  int     `json:"security_issues_count"`
	TotalMonthlyBefore   float64 `json:"total_monthly_before"`
	TotalMonthlyAfter    float64 `json:"total_monthly_after"`
	TotalMonthlyDelta    float64 `json:"total_monthly_delta"`
}

// AnalyzedPlan contains the full computed analysis of a plan.
type AnalyzedPlan struct {
	TerraformVersion   string             `json:"terraform_version"`
	FormatVersion      string             `json:"format_version"`
	Stats              SummaryStats       `json:"stats"`
	Nodes              []ResourceNode     `json:"nodes"`
	Edges              []DependencyEdge   `json:"edges"`
	DestructionAlerts  []DestructionAlert `json:"destruction_alerts"`
	SecurityFindings   []SecurityFinding  `json:"security_findings"`
	TopologicalOrder   []string           `json:"topological_order"`
}
