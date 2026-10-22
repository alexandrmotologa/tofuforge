import React, { useState, useMemo } from 'react';
import { ShieldCheck, Focus } from 'lucide-react';
import type { SecurityFinding, ResourceNode, Severity } from '../types';

interface SecurityAuditListProps {
  findings: SecurityFinding[];
  nodes: ResourceNode[];
  onSelectNode: (node: ResourceNode) => void;
}

export const SecurityAuditList: React.FC<SecurityAuditListProps> = ({
  findings,
  nodes,
  onSelectNode,
}) => {
  const [selectedSeverity, setSelectedSeverity] = useState<string>('all');

  const nodeMap = useMemo(() => {
    const map = new Map<string, ResourceNode>();
    nodes.forEach(n => map.set(n.address, n));
    return map;
  }, [nodes]);

  const filteredFindings = useMemo(() => {
    if (selectedSeverity === 'all') return findings;
    return findings.filter(f => f.severity === selectedSeverity);
  }, [findings, selectedSeverity]);

  if (findings.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center text-slate-400">
        <div className="p-3 rounded-full bg-emerald-500/10 text-emerald-400 mb-3 border border-emerald-500/20">
          <ShieldCheck className="w-6 h-6" />
        </div>
        <div className="text-sm font-semibold text-slate-200">No Security Flaws Detected</div>
        <div className="text-xs text-slate-400 mt-1 max-w-xs">
          Pre-flight linting found zero violations against cloud security baseline policies.
        </div>
      </div>
    );
  }

  const getSeverityBadge = (sev: Severity) => {
    switch (sev) {
      case 'CRITICAL':
        return 'bg-red-500 text-white';
      case 'HIGH':
        return 'bg-orange-500 text-white';
      case 'MEDIUM':
        return 'bg-amber-500 text-slate-950 font-semibold';
      case 'LOW':
        return 'bg-blue-500 text-white';
      default:
        return 'bg-slate-700 text-slate-300';
    }
  };

  return (
    <div className="space-y-3 p-4">
      {/* Header and Severity Filter Pills */}
      <div className="flex items-center justify-between flex-wrap gap-2 mb-2">
        <div className="text-xs font-semibold uppercase tracking-wider text-slate-400">
          Security Findings ({findings.length})
        </div>
        <div className="flex items-center gap-1 text-[11px]">
          {['all', 'CRITICAL', 'HIGH', 'MEDIUM'].map(sev => (
            <button
              key={sev}
              onClick={() => setSelectedSeverity(sev)}
              className={`px-2 py-0.5 rounded-full capitalize transition ${
                selectedSeverity === sev
                  ? 'bg-slate-700 text-white font-semibold'
                  : 'bg-slate-800/80 text-slate-400 hover:text-slate-200'
              }`}
            >
              {sev.toLowerCase()}
            </button>
          ))}
        </div>
      </div>

      {filteredFindings.map((finding, idx) => {
        const targetNode = nodeMap.get(finding.resource);

        return (
          <div
            key={idx}
            className="rounded-xl border border-slate-800 bg-slate-900/70 p-3.5 hover:border-slate-700 transition"
          >
            {/* Top row */}
            <div className="flex items-start justify-between gap-2 mb-1.5">
              <div className="flex items-center gap-1.5 flex-wrap">
                <span className={`text-[10px] uppercase font-bold tracking-wider px-2 py-0.5 rounded-full ${getSeverityBadge(finding.severity)}`}>
                  {finding.severity}
                </span>
                <span className="font-mono text-xs font-semibold text-slate-300">
                  {finding.rule_id}
                </span>
              </div>

              {targetNode && (
                <button
                  onClick={() => onSelectNode(targetNode)}
                  className="flex items-center gap-1 text-[11px] font-medium text-slate-300 hover:text-white px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 border border-slate-700 transition shrink-0"
                  title="Focus node in graph"
                >
                  <Focus className="w-3 h-3" /> Focus
                </button>
              )}
            </div>

            {/* Title */}
            <div className="text-sm font-semibold text-slate-100 mb-1">
              {finding.title}
            </div>

            {/* Target Resource */}
            <div className="font-mono text-xs text-slate-400 break-all mb-2">
              {finding.resource}
            </div>

            {/* Description */}
            <p className="text-xs text-slate-300 leading-relaxed mb-2.5">
              {finding.description}
            </p>

            {/* Remediation Tip */}
            <div className="p-2.5 rounded-lg bg-emerald-950/20 border border-emerald-500/20 text-xs">
              <div className="font-semibold text-emerald-400 flex items-center gap-1 mb-0.5 text-[11px]">
                Remediation:
              </div>
              <div className="text-emerald-200/90 text-[11px] leading-relaxed">
                {finding.remediation}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};
