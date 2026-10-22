export type Action = 'no-op' | 'create' | 'read' | 'update' | 'delete' | 'replace';

export type ResourceCategory = 'database' | 'storage' | 'network' | 'compute' | 'security' | 'dns' | 'other';

export type Severity = 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | 'INFO';

export interface AttributeDiff {
  path: string;
  before?: any;
  after?: any;
  forces_replacement: boolean;
  sensitive?: boolean;
  type: 'added' | 'modified' | 'removed';
}

export interface SecurityFinding {
  rule_id: string;
  title: string;
  severity: Severity;
  resource: string;
  type: string;
  description: string;
  remediation: string;
  attribute?: string;
}

export interface DestructionAlert {
  resource_address: string;
  resource_type: string;
  category: ResourceCategory;
  action: Action;
  severity: Severity;
  is_stateful: boolean;
  blast_radius: number;
  downstream_count: number;
  downstream_paths?: string[];
  message: string;
  replacement_keys?: string[];
}

export interface ResourceNode {
  id: string;
  address: string;
  name: string;
  type: string;
  provider: string;
  module: string;
  mode: string;
  action: Action;
  category: ResourceCategory;
  is_stateful: boolean;
  blast_radius_score: number;
  direct_impact: boolean;
  downstream_count: number;
  downstream_ids?: string[];
  upstream_ids?: string[];
  dependencies?: string[];
  attribute_diffs?: AttributeDiff[];
  security_findings?: SecurityFinding[];
  raw_before?: Record<string, any>;
  raw_after?: Record<string, any>;
}

export interface DependencyEdge {
  id: string;
  source: string;
  target: string;
  implicit: boolean;
  reason?: string;
}

export interface SummaryStats {
  total_resources: number;
  create_count: number;
  update_count: number;
  delete_count: number;
  replace_count: number;
  no_op_count: number;
  read_count: number;
  total_blast_radius: number;
  critical_alerts_count: number;
  security_issues_count: number;
}

export interface AnalyzedPlan {
  terraform_version?: string;
  format_version?: string;
  stats: SummaryStats;
  nodes: ResourceNode[];
  edges: DependencyEdge[];
  destruction_alerts: DestructionAlert[];
  security_findings: SecurityFinding[];
  topological_order?: string[];
}
