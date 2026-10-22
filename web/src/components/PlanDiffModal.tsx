import React, { useState, useMemo } from 'react';
import { 
  X, 
  GitCompare, 
  CheckCircle2, 
  ShieldCheck, 
  DollarSign, 
  ArrowRight, 
  Radio,
  Flame,
  Layers,
  Upload
} from 'lucide-react';
import type { AnalyzedPlan } from '../types';
import { SAMPLE_PLANS } from '../samples/samplePlans';
import { comparePlansClientSide } from '../utils/diffComparator';

interface PlanDiffModalProps {
  currentPlan: AnalyzedPlan;
  onApplyPlan?: (plan: AnalyzedPlan) => void;
  onClose: () => void;
}

export const PlanDiffModal: React.FC<PlanDiffModalProps> = ({
  currentPlan,
  onApplyPlan,
  onClose,
}) => {
  // Plan A (Base) and Plan B (Target)
  const [basePlan, setBasePlan] = useState<AnalyzedPlan>(SAMPLE_PLANS[0].plan);
  const [targetPlan, setTargetPlan] = useState<AnalyzedPlan>(SAMPLE_PLANS[1]?.plan || currentPlan);

  const [baseLabel, setBaseLabel] = useState<string>('Plan Revision 1 (Base)');
  const [targetLabel, setTargetLabel] = useState<string>('Plan Revision 2 (Target)');

  // Compare diff
  const diffResult = useMemo(() => {
    return comparePlansClientSide(basePlan, targetPlan);
  }, [basePlan, targetPlan]);

  const handleFileUpload = (
    e: React.ChangeEvent<HTMLInputElement>,
    side: 'base' | 'target'
  ) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = evt => {
      try {
        const parsed = JSON.parse(evt.target?.result as string);
        if (parsed.stats && parsed.nodes) {
          if (side === 'base') {
            setBasePlan(parsed);
            setBaseLabel(file.name);
          } else {
            setTargetPlan(parsed);
            setTargetLabel(file.name);
          }
        } else {
          alert('Uploaded JSON is not an analyzed plan. Ingest it first or upload an analyzed file.');
        }
      } catch (err: any) {
        alert('Invalid JSON file: ' + err.message);
      }
    };
    reader.readAsText(file);
  };

  const handleSelectSample = (sampleId: string, side: 'base' | 'target') => {
    const found = SAMPLE_PLANS.find(s => s.id === sampleId);
    if (found) {
      if (side === 'base') {
        setBasePlan(found.plan);
        setBaseLabel(found.title);
      } else {
        setTargetPlan(found.plan);
        setTargetLabel(found.title);
      }
    }
  };

  const isBlastRadiusImproved = diffResult.blast_radius_delta < 0;
  const isBlastRadiusWorse = diffResult.blast_radius_delta > 0;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md overflow-y-auto">
      <div className="relative w-full max-w-4xl bg-slate-900 border border-slate-800 rounded-2xl shadow-2xl flex flex-col max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="p-4 px-6 border-b border-slate-800 flex items-center justify-between bg-slate-900/90">
          <div className="flex items-center gap-3">
            <div className="p-2 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400">
              <GitCompare className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-base font-bold text-white flex items-center gap-2">
                Plan-to-Plan Diff & Safety Comparison
              </h2>
              <p className="text-xs text-slate-400">
                Compare two Terraform/OpenTofu plan revisions to verify resolved destructions and security regressions
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-1.5 rounded-lg bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Plan Selectors Bar */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 p-4 px-6 bg-slate-950/50 border-b border-slate-800 text-xs">
          {/* Base Plan Selection */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <span className="font-semibold uppercase tracking-wider text-slate-400 text-[10px]">
                Base Plan (Before)
              </span>
              <label className="text-[10px] text-blue-400 hover:text-blue-300 cursor-pointer flex items-center gap-1">
                <Upload className="w-3 h-3" /> Upload JSON
                <input
                  type="file"
                  accept=".json"
                  className="hidden"
                  onChange={e => handleFileUpload(e, 'base')}
                />
              </label>
            </div>
            <select
              value={SAMPLE_PLANS.find(s => s.plan === basePlan)?.id || ''}
              onChange={e => handleSelectSample(e.target.value, 'base')}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2 text-slate-200 font-mono text-xs focus:outline-none"
            >
              {SAMPLE_PLANS.map(s => (
                <option key={s.id} value={s.id}>
                  {s.title}
                </option>
              ))}
            </select>
            <div className="text-[11px] text-slate-400 truncate">Active: {baseLabel}</div>
          </div>

          {/* Target Plan Selection */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <span className="font-semibold uppercase tracking-wider text-slate-400 text-[10px]">
                Target Plan (After / Remediated)
              </span>
              <label className="text-[10px] text-blue-400 hover:text-blue-300 cursor-pointer flex items-center gap-1">
                <Upload className="w-3 h-3" /> Upload JSON
                <input
                  type="file"
                  accept=".json"
                  className="hidden"
                  onChange={e => handleFileUpload(e, 'target')}
                />
              </label>
            </div>
            <select
              value={SAMPLE_PLANS.find(s => s.plan === targetPlan)?.id || ''}
              onChange={e => handleSelectSample(e.target.value, 'target')}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg p-2 text-slate-200 font-mono text-xs focus:outline-none"
            >
              {SAMPLE_PLANS.map(s => (
                <option key={s.id} value={s.id}>
                  {s.title}
                </option>
              ))}
            </select>
            <div className="text-[11px] text-slate-400 truncate">Active: {targetLabel}</div>
          </div>
        </div>

        {/* Content Body */}
        <div className="flex-1 overflow-y-auto p-6 space-y-6">
          {/* Key Metric Diff Cards */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {/* Blast Radius Delta */}
            <div className={`p-3 rounded-xl border ${
              isBlastRadiusImproved
                ? 'bg-emerald-950/20 border-emerald-500/40'
                : isBlastRadiusWorse
                ? 'bg-rose-950/20 border-rose-500/40'
                : 'bg-slate-950/50 border-slate-800'
            }`}>
              <div className="flex items-center gap-1.5 text-xs text-slate-400 mb-1">
                <Radio className="w-3.5 h-3.5" /> Blast Radius Delta
              </div>
              <div className={`text-xl font-bold font-mono ${
                isBlastRadiusImproved
                  ? 'text-emerald-400'
                  : isBlastRadiusWorse
                  ? 'text-rose-400'
                  : 'text-slate-200'
              }`}>
                {diffResult.blast_radius_delta > 0 ? `+` : ``}{diffResult.blast_radius_delta}
              </div>
              <div className="text-[10px] text-slate-400 mt-1">
                {isBlastRadiusImproved ? 'Risk reduced' : isBlastRadiusWorse ? 'Risk increased' : 'No change'}
              </div>
            </div>

            {/* Resolved Destructions */}
            <div className="p-3 rounded-xl border bg-emerald-950/20 border-emerald-500/40">
              <div className="flex items-center gap-1.5 text-xs text-slate-400 mb-1">
                <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" /> Resolved Destructions
              </div>
              <div className="text-xl font-bold font-mono text-emerald-400">
                {diffResult.resolved_destruction.length}
              </div>
              <div className="text-[10px] text-slate-400 mt-1">
                {diffResult.resolved_destruction.length > 0 ? 'Replacements prevented' : 'None resolved'}
              </div>
            </div>

            {/* Security Fixes */}
            <div className="p-3 rounded-xl border bg-blue-950/20 border-blue-500/40">
              <div className="flex items-center gap-1.5 text-xs text-slate-400 mb-1">
                <ShieldCheck className="w-3.5 h-3.5 text-blue-400" /> Security Fixes
              </div>
              <div className="text-xl font-bold font-mono text-blue-400">
                {diffResult.resolved_security.length}
              </div>
              <div className="text-[10px] text-slate-400 mt-1">
                Flaws closed in target
              </div>
            </div>

            {/* Cost Delta */}
            <div className="p-3 rounded-xl border bg-slate-950/50 border-slate-800">
              <div className="flex items-center gap-1.5 text-xs text-slate-400 mb-1">
                <DollarSign className="w-3.5 h-3.5 text-slate-300" /> Monthly Delta Shift
              </div>
              <div className="text-xl font-bold font-mono text-slate-200">
                {diffResult.cost_delta > 0 ? '+' : ''}${diffResult.cost_delta.toFixed(2)}/mo
              </div>
              <div className="text-[10px] text-slate-400 mt-1">
                Cost difference between plans
              </div>
            </div>
          </div>

          {/* Resolved Destructions Section */}
          {diffResult.resolved_destruction.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-emerald-400 flex items-center gap-1.5">
                <CheckCircle2 className="w-4 h-4" /> Destructive Risks Successfully Remediated ({diffResult.resolved_destruction.length})
              </h3>
              <div className="space-y-2">
                {diffResult.resolved_destruction.map((alert, i) => (
                  <div key={i} className="p-3 rounded-xl bg-emerald-950/20 border border-emerald-500/30 text-xs">
                    <div className="flex items-center justify-between gap-2 mb-1">
                      <span className="font-mono font-semibold text-emerald-200 break-all">
                        {alert.resource_address}
                      </span>
                      <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-300 font-semibold text-[10px] uppercase">
                        Safeguarded
                      </span>
                    </div>
                    <div className="text-slate-300 text-[11px] leading-relaxed">
                      Originally flagged for {alert.action === 'replace' ? 'destruction & recreation' : 'deletion'}. In the updated plan, this destructive cascade has been resolved.
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* New Destructions Introduced */}
          {diffResult.new_destruction.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-rose-400 flex items-center gap-1.5">
                <Flame className="w-4 h-4" /> New Destructive Actions Introduced ({diffResult.new_destruction.length})
              </h3>
              <div className="space-y-2">
                {diffResult.new_destruction.map((alert, i) => (
                  <div key={i} className="p-3 rounded-xl bg-rose-950/20 border border-rose-500/30 text-xs">
                    <div className="font-mono font-semibold text-rose-200 break-all mb-1">
                      {alert.resource_address}
                    </div>
                    <div className="text-slate-300 text-[11px]">
                      {alert.message}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Resolved Security Fixes */}
          {diffResult.resolved_security.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-blue-400 flex items-center gap-1.5">
                <ShieldCheck className="w-4 h-4" /> Security Violations Resolved ({diffResult.resolved_security.length})
              </h3>
              <div className="space-y-2">
                {diffResult.resolved_security.map((sec, i) => (
                  <div key={i} className="p-3 rounded-xl bg-blue-950/20 border border-blue-500/30 text-xs">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-mono font-bold text-blue-300">{sec.rule_id}</span>
                      <span className="text-slate-300 font-medium">{sec.title}</span>
                    </div>
                    <div className="text-[11px] font-mono text-slate-400 break-all">
                      {sec.resource}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Action Transitions */}
          {diffResult.action_changes.length > 0 && (
            <div className="space-y-2">
              <h3 className="text-xs font-semibold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                <Layers className="w-4 h-4" /> Resource Action Transitions ({diffResult.action_changes.length})
              </h3>
              <div className="rounded-xl border border-slate-800 bg-slate-950/60 overflow-hidden">
                <table className="w-full text-left text-xs font-mono">
                  <thead className="bg-slate-900 border-b border-slate-800 text-[10px] text-slate-400 uppercase">
                    <tr>
                      <th className="p-2.5 px-3">Resource</th>
                      <th className="p-2.5 px-3">Base Action</th>
                      <th className="p-2.5 px-3">Target Action</th>
                      <th className="p-2.5 px-3">Outcome</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-800/80">
                    {diffResult.action_changes.map((ac, idx) => {
                      const wasDestructive = ac.base_action === 'replace' || ac.base_action === 'delete';
                      const isNowSafe = ac.target_action === 'update' || ac.target_action === 'no-op' || ac.target_action === 'read';
                      const isImproved = wasDestructive && isNowSafe;

                      return (
                        <tr key={idx} className="hover:bg-slate-800/40 transition">
                          <td className="p-2.5 px-3 text-slate-200 font-medium truncate max-w-xs" title={ac.address}>
                            {ac.address}
                          </td>
                          <td className="p-2.5 px-3">
                            <span className={`px-1.5 py-0.5 rounded text-[10px] uppercase font-semibold ${
                              ac.base_action === 'replace' ? 'bg-rose-500/20 text-rose-300' :
                              ac.base_action === 'delete' ? 'bg-red-500/20 text-red-300' :
                              'bg-slate-800 text-slate-400'
                            }`}>
                              {ac.base_action}
                            </span>
                          </td>
                          <td className="p-2.5 px-3">
                            <span className={`px-1.5 py-0.5 rounded text-[10px] uppercase font-semibold ${
                              ac.target_action === 'replace' ? 'bg-rose-500/20 text-rose-300' :
                              ac.target_action === 'delete' ? 'bg-red-500/20 text-red-300' :
                              ac.target_action === 'update' ? 'bg-amber-500/20 text-amber-300' :
                              'bg-emerald-500/20 text-emerald-300'
                            }`}>
                              {ac.target_action}
                            </span>
                          </td>
                          <td className="p-2.5 px-3 text-[11px] font-sans">
                            {isImproved ? (
                              <span className="text-emerald-400 font-semibold flex items-center gap-1">
                                <CheckCircle2 className="w-3.5 h-3.5" /> Remediated
                              </span>
                            ) : (
                              <span className="text-slate-400">Shifted</span>
                            )}
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="p-4 px-6 border-t border-slate-800 flex items-center justify-between bg-slate-900/90 text-xs">
          <div className="text-slate-400">
            Comparing <b>{diffResult.base_stats.total_resources}</b> vs <b>{diffResult.target_stats.total_resources}</b> resources
          </div>

          <div className="flex items-center gap-2">
            {onApplyPlan && (
              <button
                onClick={() => {
                  onApplyPlan(targetPlan);
                  onClose();
                }}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-semibold shadow-sm transition"
              >
                <span>Load Target Plan into Studio</span>
                <ArrowRight className="w-3.5 h-3.5" />
              </button>
            )}
            <button
              onClick={onClose}
              className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
            >
              Close
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
