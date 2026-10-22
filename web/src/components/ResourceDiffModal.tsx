import React, { useState } from 'react';
import { X, Flame, AlertTriangle, Radio, ShieldAlert } from 'lucide-react';
import type { ResourceNode } from '../types';

interface ResourceDiffModalProps {
  node: ResourceNode | null;
  onClose: () => void;
}

export const ResourceDiffModal: React.FC<ResourceDiffModalProps> = ({ node, onClose }) => {
  const [activeTab, setActiveTab] = useState<'diff' | 'raw'>('diff');

  if (!node) return null;

  return (
    <div className="fixed inset-y-0 right-0 w-full max-w-xl bg-slate-900/95 border-l border-slate-800 shadow-2xl backdrop-blur-xl z-50 flex flex-col transition-all">
      {/* Header */}
      <div className="p-4 border-b border-slate-800 flex items-start justify-between gap-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2 flex-wrap mb-1">
            <span className="text-xs font-mono font-medium text-slate-400">
              {node.type}
            </span>
            <span className="text-[10px] font-semibold uppercase px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
              {node.provider}
            </span>
            <span className="text-[10px] font-semibold uppercase px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
              {node.action}
            </span>
          </div>
          <h2 className="text-base font-bold font-mono text-slate-100 truncate" title={node.address}>
            {node.address}
          </h2>
        </div>

        <button
          onClick={onClose}
          className="p-1.5 rounded-lg bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 transition"
        >
          <X className="w-5 h-5" />
        </button>
      </div>

      {/* Meta Bar */}
      <div className="px-4 py-3 bg-slate-950/60 border-b border-slate-800/80 flex items-center justify-between gap-2 text-xs flex-wrap">
        <div className="flex items-center gap-3">
          {node.is_stateful && (
            <span className="flex items-center gap-1 text-amber-400 font-medium">
              <AlertTriangle className="w-3.5 h-3.5" /> Stateful {node.category}
            </span>
          )}
          {node.blast_radius_score > 0 && (
            <span className="flex items-center gap-1 text-rose-400 font-medium">
              <Radio className="w-3.5 h-3.5" /> Blast Radius: {node.blast_radius_score}
            </span>
          )}
        </div>

        <div className="flex items-center gap-1 p-0.5 rounded bg-slate-800 border border-slate-700">
          <button
            onClick={() => setActiveTab('diff')}
            className={`px-2.5 py-1 rounded text-[11px] font-medium transition ${
              activeTab === 'diff'
                ? 'bg-slate-700 text-white shadow-sm'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Diffs ({node.attribute_diffs?.length || 0})
          </button>
          <button
            onClick={() => setActiveTab('raw')}
            className={`px-2.5 py-1 rounded text-[11px] font-medium transition ${
              activeTab === 'raw'
                ? 'bg-slate-700 text-white shadow-sm'
                : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Raw JSON
          </button>
        </div>
      </div>

      {/* Content Area */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {/* Security Alerts attached to this node */}
        {node.security_findings && node.security_findings.length > 0 && (
          <div className="p-3 rounded-xl bg-red-950/20 border border-red-500/30 text-xs space-y-2">
            <div className="flex items-center gap-1.5 font-bold text-red-400">
              <ShieldAlert className="w-4 h-4" /> Pre-Flight Security Violations
            </div>
            {node.security_findings.map((f, i) => (
              <div key={i} className="pl-3 border-l-2 border-red-500/40 text-slate-300">
                <span className="font-mono text-red-300 font-medium">{f.rule_id}</span>: {f.title}
              </div>
            ))}
          </div>
        )}

        {activeTab === 'diff' ? (
          <div>
            {!node.attribute_diffs || node.attribute_diffs.length === 0 ? (
              <div className="p-6 text-center text-slate-400 text-xs">
                No attribute differences recorded for this resource action.
              </div>
            ) : (
              <div className="space-y-3">
                {node.attribute_diffs.map((diff, index) => (
                  <div
                    key={index}
                    className={`rounded-xl border p-3 text-xs transition ${
                      diff.forces_replacement
                        ? 'bg-rose-950/20 border-rose-500/40'
                        : 'bg-slate-950/40 border-slate-800'
                    }`}
                  >
                    <div className="flex items-center justify-between gap-2 mb-2 pb-1.5 border-b border-slate-800/60">
                      <span className="font-mono font-semibold text-slate-200 break-all">
                        {diff.path}
                      </span>
                      {diff.forces_replacement && (
                        <span className="flex items-center gap-1 px-2 py-0.5 rounded text-[10px] font-bold uppercase bg-rose-500 text-white shrink-0 shadow-sm shadow-rose-950">
                          <Flame className="w-3 h-3" /> Forces Replacement
                        </span>
                      )}
                    </div>

                    <div className="grid grid-cols-2 gap-2 text-[11px] font-mono">
                      <div className="p-2 rounded bg-slate-900/80 border border-slate-800">
                        <div className="text-[10px] uppercase font-sans text-slate-400 mb-1">Before:</div>
                        <div className="text-red-400/90 break-all">
                          {diff.before !== undefined && diff.before !== null
                            ? JSON.stringify(diff.before, null, 2)
                            : '(none)'}
                        </div>
                      </div>

                      <div className="p-2 rounded bg-slate-900/80 border border-slate-800">
                        <div className="text-[10px] uppercase font-sans text-slate-400 mb-1">After:</div>
                        <div className="text-emerald-400/90 break-all">
                          {diff.after !== undefined && diff.after !== null
                            ? JSON.stringify(diff.after, null, 2)
                            : '(none)'}
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        ) : (
          <div className="space-y-3 text-xs font-mono">
            <div>
              <div className="text-slate-400 font-sans text-xs mb-1">Planned 'After' State:</div>
              <pre className="p-3 rounded-xl bg-slate-950 border border-slate-800 text-emerald-400 overflow-x-auto text-[11px] max-h-60">
                {JSON.stringify(node.raw_after || {}, null, 2)}
              </pre>
            </div>
            <div>
              <div className="text-slate-400 font-sans text-xs mb-1">Prior 'Before' State:</div>
              <pre className="p-3 rounded-xl bg-slate-950 border border-slate-800 text-red-400 overflow-x-auto text-[11px] max-h-60">
                {JSON.stringify(node.raw_before || {}, null, 2)}
              </pre>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
