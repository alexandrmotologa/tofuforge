package model

// Action represents the classified change action for a resource.
type Action string

const (
	ActionNoOp    Action = "no-op"
	ActionCreate  Action = "create"
	ActionRead    Action = "read"
	ActionUpdate  Action = "update"
	ActionDelete  Action = "delete"
	ActionReplace Action = "replace"
)

// ResourceCategory categorizes cloud infrastructure by operational role.
type ResourceCategory string

const (
	CategoryDatabase ResourceCategory = "database"
	CategoryStorage  ResourceCategory = "storage"
	CategoryNetwork  ResourceCategory = "network"
	CategoryCompute  ResourceCategory = "compute"
	CategorySecurity ResourceCategory = "security"
	CategoryDNS      ResourceCategory = "dns"
	CategoryOther    ResourceCategory = "other"
)

// RawPlan represents the top-level structure of a Terraform or OpenTofu JSON plan.
type RawPlan struct {
	FormatVersion    string                 `json:"format_version"`
	TerraformVersion string                 `json:"terraform_version,omitempty"`
	Variables        map[string]interface{} `json:"variables,omitempty"`
	PlannedValues    *PlannedValues         `json:"planned_values,omitempty"`
	ResourceChanges  []RawResourceChange    `json:"resource_changes,omitempty"`
	OutputChanges    map[string]RawOutput   `json:"output_changes,omitempty"`
	PriorState       interface{}            `json:"prior_state,omitempty"`
	Configuration    *RawConfiguration      `json:"configuration,omitempty"`
}

// PlannedValues contains the planned state tree.
type PlannedValues struct {
	RootModule *PlannedModule `json:"root_module,omitempty"`
}

// PlannedModule represents resources in a module hierarchy.
type PlannedModule struct {
	Address      string                 `json:"address,omitempty"`
	Resources    []RawResourceState     `json:"resources,omitempty"`
	ChildModules []PlannedModule        `json:"child_modules,omitempty"`
}

// RawResourceState represents an individual planned resource state.
type RawResourceState struct {
	Address      string                 `json:"address"`
	Mode         string                 `json:"mode"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Index        interface{}            `json:"index,omitempty"`
	ProviderName string                 `json:"provider_name"`
	Values       map[string]interface{} `json:"values,omitempty"`
}

// RawResourceChange represents a single resource change from terraform-json.
type RawResourceChange struct {
	Address         string                 `json:"address"`
	PreviousAddress string                 `json:"previous_address,omitempty"`
	ModuleAddress   string                 `json:"module_address,omitempty"`
	Mode            string                 `json:"mode"`
	Type            string                 `json:"type"`
	Name            string                 `json:"name"`
	Index           interface{}            `json:"index,omitempty"`
	ProviderName    string                 `json:"provider_name"`
	DeposedKey      string                 `json:"deposed_key,omitempty"`
	Change          RawChange              `json:"change"`
	ActionReason    string                 `json:"action_reason,omitempty"`
}

// RawChange defines the before/after state and actions.
type RawChange struct {
	Actions         []string               `json:"actions"`
	Before          map[string]interface{} `json:"before,omitempty"`
	After           map[string]interface{} `json:"after,omitempty"`
	AfterUnknown    map[string]interface{} `json:"after_unknown,omitempty"`
	BeforeSensitive interface{}            `json:"before_sensitive,omitempty"`
	AfterSensitive  interface{}            `json:"after_sensitive,omitempty"`
	ReplacePaths    [][]interface{}        `json:"replace_paths,omitempty"`
}

// RawOutput represents output changes in the plan.
type RawOutput struct {
	Actions []string    `json:"actions"`
	Before  interface{} `json:"before,omitempty"`
	After   interface{} `json:"after,omitempty"`
}

// RawConfiguration defines the configuration references tree.
type RawConfiguration struct {
	ProviderConfig map[string]interface{} `json:"provider_config,omitempty"`
	RootModule     *ConfigModule          `json:"root_module,omitempty"`
}

// ConfigModule represents a module configuration with dependencies.
type ConfigModule struct {
	Resources    []ConfigResource            `json:"resources,omitempty"`
	ModuleCalls  map[string]ConfigModuleCall `json:"module_calls,omitempty"`
	Variables    map[string]interface{}      `json:"variables,omitempty"`
	Outputs      map[string]interface{}      `json:"outputs,omitempty"`
}

// ConfigResource represents a resource definition in configuration.
type ConfigResource struct {
	Address           string                 `json:"address"`
	Mode              string                 `json:"mode"`
	Type              string                 `json:"type"`
	Name              string                 `json:"name"`
	ProviderConfigKey string                 `json:"provider_config_key,omitempty"`
	Expressions       map[string]interface{} `json:"expressions,omitempty"`
	DependsOn         []string               `json:"depends_on,omitempty"`
}

// ConfigModuleCall represents a child module call.
type ConfigModuleCall struct {
	Source      string                 `json:"source"`
	Expressions map[string]interface{} `json:"expressions,omitempty"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	Module      *ConfigModule          `json:"module,omitempty"`
}

// AttributeDiff captures the difference between before and after values for a specific attribute.
type AttributeDiff struct {
	Path               string      `json:"path"`
	Before             interface{} `json:"before,omitempty"`
	After              interface{} `json:"after,omitempty"`
	ForcesReplacement  bool        `json:"forces_replacement"`
	Sensitive          bool        `json:"sensitive"`
	Type               string      `json:"type"` // "added", "modified", "removed"
}
