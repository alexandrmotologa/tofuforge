import React from 'react';
import { Flame, ArrowRight, CornerDownRight, Focus, AlertOctagon } from 'lucide-react';
import type { DestructionAlert, ResourceNode } from '../types';

interface DestructionAlertsProps {
  alerts: DestructionAlert[];
  nodes: ResourceNode[];
  onSelectNode: (node: ResourceNode) => void;
}

export const DestructionAlerts: React.FC<DestructionAlertsProps> = ({
  alerts,
  nodes,
  onSelectNode,
}) => {
  const nodeMap = React.useMemo(() => {
    const map = new Map<string, ResourceNode>();
    nodes.forEach(n => map.set(n.address, n));
    return map;
  }, [nodes]);

  if (alerts.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center p-8 text-center text-slate-400">
        <div className="p-3 rounded-full bg-emerald-500/10 text-emerald-400 mb-3 border border-emerald-500/20">
          <AlertOctagon className="w-6 h-6" />
        </div>
        <div className="text-sm font-semibold text-slate-200">No Destructive Changes</div>
        <div className="text-xs text-slate-400 mt-1 max-w-xs">
          Zero resources are scheduled for deletion or recreation in this plan.
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-3 p-4">
      <div className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-2">
        Destructive & Replacement Actions ({alerts.length})
      </div>

      {alerts.map((alert, idx) => {
        const targetNode = nodeMap.get(alert.resource_address);

        return (
          <div
            key={idx}
            className={`rounded-xl border p-3.5 transition-all ${
              alert.severity === 'CRITICAL'
                ? 'bg-rose-950/20 border-rose-500/40 shadow-sm shadow-rose-950'
                : 'bg-amber-950/20 border-amber-500/40'
            }`}
          >
            {/* Header */}
            <div className="flex items-start justify-between gap-2 mb-2">
              <div className="flex items-center gap-1.5 flex-wrap">
                <span
                  className={`text-[10px] font-bold uppercase tracking-wider px-2 py-0.5 rounded-full ${
                    alert.severity === 'CRITICAL'
                      ? 'bg-rose-500 text-white'
                      : 'bg-amber-500 text-slate-950'
                  }`}
                >
                  {alert.severity}
                </span>

                <span className="flex items-center gap-1 text-[11px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
                  {alert.action === 'replace' ? (
                    <>
                      <Flame className="w-3 h-3 text-rose-400" /> Destroy & Recreate
                    </>
                  ) : (
                    'Delete'
                  )}
                </span>

                {alert.is_stateful && (
                  <span className="text-[10px] font-semibold uppercase tracking-wider px-1.5 py-0.5 rounded bg-amber-500/20 text-amber-300 border border-amber-500/30">
                    Stateful {alert.category}
                  </span>
                )}
              </div>

              {targetNode && (
                <button
                  onClick={() => onSelectNode(targetNode)}
                  className="flex items-center gap-1 text-[11px] font-medium text-slate-300 hover:text-white px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 border border-slate-700 transition"
                  title="Focus node in graph"
                >
                  <Focus className="w-3 h-3" /> Focus
                </button>
              )}
            </div>

            {/* Address */}
            <div className="font-mono text-xs font-semibold text-slate-100 break-all mb-1">
              {alert.resource_address}
            </div>

            {/* Explanatory Message */}
            <p className="text-xs text-slate-300 leading-relaxed mb-2.5">
              {alert.message}
            </p>

            {/* Replacement Triggers */}
            {alert.replacement_keys && alert.replacement_keys.length > 0 && (
              <div className="mb-2 p-2 rounded bg-slate-900/80 border border-slate-800 text-xs">
                <div className="text-[11px] font-medium text-slate-400 mb-1">Replacement Triggers:</div>
                <div className="flex flex-wrap gap-1">
                  {alert.replacement_keys.map(k => (
                    <span
                      key={k}
                      className="px-1.5 py-0.5 rounded font-mono text-[10px] bg-rose-500/10 text-rose-300 border border-rose-500/30"
                    >
                      {k}
                    </span>
                  ))}
                </div>
              </div>
            )}

            {/* Downstream Cascade List */}
            {alert.downstream_paths && alert.downstream_paths.length > 0 && (
              <div className="mt-2 text-xs">
                <div className="flex items-center gap-1 text-slate-400 text-[11px] font-medium mb-1">
                  <CornerDownRight className="w-3 h-3 text-amber-400" />
                  Cascades Downstream to {alert.downstream_paths.length} resource{alert.downstream_paths.length > 1 ? 's' : ''}:
                </div>
                <div className="space-y-1 pl-3 border-l-2 border-slate-800">
                  {alert.downstream_paths.map(dp => {
                    const childNode = nodeMap.get(dp);
                    return (
                      <div
                        key={dp}
                        onClick={() => childNode && onSelectNode(childNode)}
                        className="flex items-center justify-between text-[11px] font-mono text-slate-300 hover:text-amber-300 cursor-pointer p-1 rounded hover:bg-slate-800/60 transition"
                      >
                        <span className="truncate">{dp}</span>
                        <ArrowRight className="w-3 h-3 opacity-50 shrink-0 ml-1" />
                      </div>
                    );
                  })}
                </div>
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
};
