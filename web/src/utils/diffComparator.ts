import type { AnalyzedPlan, PlanComparison, DestructionAlert, SecurityFinding, ActionChange } from '../types';

export function comparePlansClientSide(base: AnalyzedPlan, target: AnalyzedPlan): PlanComparison {
  const baseAlertMap = new Map<string, DestructionAlert>();
  base.destruction_alerts?.forEach(a => baseAlertMap.set(a.resource_address, a));

  const targetAlertMap = new Map<string, DestructionAlert>();
  target.destruction_alerts?.forEach(a => targetAlertMap.set(a.resource_address, a));

  const resolvedDestruction: DestructionAlert[] = [];
  const newDestruction: DestructionAlert[] = [];
  const persistingDestruction: DestructionAlert[] = [];

  baseAlertMap.forEach((alert, addr) => {
    if (!targetAlertMap.has(addr)) {
      resolvedDestruction.push(alert);
    } else {
      persistingDestruction.push(alert);
    }
  });

  targetAlertMap.forEach((alert, addr) => {
    if (!baseAlertMap.has(addr)) {
      newDestruction.push(alert);
    }
  });

  // Security differences
  const baseSecMap = new Map<string, SecurityFinding>();
  base.security_findings?.forEach(s => baseSecMap.set(`${s.rule_id}::${s.resource}`, s));

  const targetSecMap = new Map<string, SecurityFinding>();
  target.security_findings?.forEach(s => targetSecMap.set(`${s.rule_id}::${s.resource}`, s));

  const resolvedSecurity: SecurityFinding[] = [];
  const newSecurity: SecurityFinding[] = [];

  baseSecMap.forEach((finding, key) => {
    if (!targetSecMap.has(key)) {
      resolvedSecurity.push(finding);
    }
  });

  targetSecMap.forEach((finding, key) => {
    if (!baseSecMap.has(key)) {
      newSecurity.push(finding);
    }
  });

  // Resources added / removed / action changed
  const baseNodeMap = new Map<string, any>();
  base.nodes?.forEach(n => baseNodeMap.set(n.address, n));

  const targetNodeMap = new Map<string, any>();
  target.nodes?.forEach(n => targetNodeMap.set(n.address, n));

  const addedResources: string[] = [];
  const removedResources: string[] = [];
  const actionChanges: ActionChange[] = [];

  baseNodeMap.forEach((node, addr) => {
    if (!targetNodeMap.has(addr)) {
      removedResources.push(addr);
    } else {
      const targetNode = targetNodeMap.get(addr);
      if (node.action !== targetNode.action) {
        actionChanges.push({
          address: addr,
          base_action: node.action,
          target_action: targetNode.action,
        });
      }
    }
  });

  targetNodeMap.forEach((_, addr) => {
    if (!baseNodeMap.has(addr)) {
      addedResources.push(addr);
    }
  });

  const blastRadiusDelta = (target.stats.total_blast_radius || 0) - (base.stats.total_blast_radius || 0);
  const costDelta = (target.stats.total_monthly_delta || 0) - (base.stats.total_monthly_delta || 0);

  return {
    base_stats: base.stats,
    target_stats: target.stats,
    blast_radius_delta: blastRadiusDelta,
    cost_delta: costDelta,
    resolved_destruction: resolvedDestruction,
    new_destruction: newDestruction,
    persisting_destruction: persistingDestruction,
    resolved_security: resolvedSecurity,
    new_security: newSecurity,
    added_resources: addedResources,
    removed_resources: removedResources,
    action_changes: actionChanges,
  };
}
