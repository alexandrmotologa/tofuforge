import React, { useState, useEffect } from 'react';
import { 
  Flame, 
  ShieldAlert, 
  UploadCloud, 
  FileText, 
  Search, 
  AlertTriangle, 
  SlidersHorizontal,
  ChevronRight,
  ChevronLeft,
  X,
  Sparkles,
  GitCompare,
  Download
} from 'lucide-react';
import type { AnalyzedPlan, ResourceNode } from './types';
import { SAMPLE_PLANS } from './samples/samplePlans';
import { StatCards } from './components/StatCards';
import { PlanGraphView } from './components/PlanGraphView';
import { DestructionAlerts } from './components/DestructionAlerts';
import { SecurityAuditList } from './components/SecurityAuditList';
import { ResourceDiffModal } from './components/ResourceDiffModal';
import { PlanDropzone } from './components/PlanDropzone';
import { PRMarkdownModal } from './components/PRMarkdownModal';
import { PlanDiffModal } from './components/PlanDiffModal';

export const App: React.FC = () => {
  // Current active plan
  const [currentPlan, setCurrentPlan] = useState<AnalyzedPlan>(() => {
    if (typeof window !== 'undefined' && window.__TOFUFORGE_PRELOADED_PLAN__) {
      return window.__TOFUFORGE_PRELOADED_PLAN__;
    }
    return SAMPLE_PLANS[0].plan;
  });

  const [activeSampleId, setActiveSampleId] = useState<string>(() => {
    if (typeof window !== 'undefined' && window.__TOFUFORGE_PRELOADED_PLAN__) {
      return '';
    }
    return SAMPLE_PLANS[0].id;
  });

  // UI state
  const [selectedNode, setSelectedNode] = useState<ResourceNode | null>(null);
  const [actionFilter, setActionFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [activeSideTab, setActiveSideTab] = useState<'destruction' | 'security'>('destruction');
  const [isSidebarOpen, setIsSidebarOpen] = useState<boolean>(true);
  const [isDropzoneOpen, setIsDropzoneOpen] = useState<boolean>(false);
  const [isMarkdownModalOpen, setIsMarkdownModalOpen] = useState<boolean>(false);
  const [isDiffModalOpen, setIsDiffModalOpen] = useState<boolean>(false);
  const [isServerConnected, setIsServerConnected] = useState<boolean>(false);

  // Try checking backend /api/plan on load if not preloaded
  useEffect(() => {
    if (typeof window !== 'undefined' && window.__TOFUFORGE_PRELOADED_PLAN__) {
      return;
    }

    fetch('/api/plan')
      .then(res => {
        if (res.ok) return res.json();
        throw new Error('Not connected');
      })
      .then((data: AnalyzedPlan) => {
        if (data.nodes && data.nodes.length > 0) {
          setCurrentPlan(data);
          setIsServerConnected(true);
        }
      })
      .catch(() => {
        setIsServerConnected(false);
      });
  }, []);

  const handleSelectSample = (sampleId: string) => {
    const found = SAMPLE_PLANS.find(s => s.id === sampleId);
    if (found) {
      setCurrentPlan(found.plan);
      setActiveSampleId(sampleId);
      setSelectedNode(null);

      if (isServerConnected) {
        fetch('/api/samples/select', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ sample_id: sampleId }),
        }).catch(() => {});
      }
    }
  };

  const handlePlanUploaded = (newPlan: AnalyzedPlan) => {
    setCurrentPlan(newPlan);
    setActiveSampleId('');
    setSelectedNode(null);
  };

  const handleSelectNode = (node: ResourceNode) => {
    setSelectedNode(node);
  };

  const handleExportHTML = () => {
    if (isServerConnected) {
      window.open('/api/report/html', '_blank');
      return;
    }

    // Client-side fallback: generate standalone HTML with embedded state
    const docClone = document.documentElement.outerHTML;
    const injectionScript = `<script>window.__TOFUFORGE_PRELOADED_PLAN__ = ${JSON.stringify(currentPlan)};</script>`;
    const fullHTML = `<!DOCTYPE html>\n${docClone.replace('</head>', `${injectionScript}</head>`)}`;

    const blob = new Blob([fullHTML], { type: 'text/html;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'tofuforge-report.html';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const getStatusBadge = () => {
    if (currentPlan.stats.critical_alerts_count > 0) {
      return {
        text: 'CRITICAL DESTRUCTION RISK',
        className: 'bg-rose-500/20 text-rose-300 border-rose-500/40',
        icon: <Flame className="w-3.5 h-3.5 text-rose-400 animate-pulse" />
      };
    }
    if (currentPlan.stats.replace_count > 0 || currentPlan.stats.delete_count > 0) {
      return {
        text: 'DESTRUCTIVE CHANGES',
        className: 'bg-amber-500/20 text-amber-300 border-amber-500/40',
        icon: <AlertTriangle className="w-3.5 h-3.5 text-amber-400" />
      };
    }
    if (currentPlan.stats.security_issues_count > 0) {
      return {
        text: 'SECURITY FINDINGS',
        className: 'bg-red-500/20 text-red-300 border-red-500/40',
        icon: <ShieldAlert className="w-3.5 h-3.5 text-red-400" />
      };
    }
    return {
      text: 'SAFE TO APPLY',
      className: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40',
      icon: <span className="w-2 h-2 rounded-full bg-emerald-400" />
    };
  };

  const status = getStatusBadge();

  return (
    <div className="flex flex-col h-screen w-screen overflow-hidden bg-slate-950 text-slate-100 select-none">
      {/* Top Navigation Bar */}
      <header className="h-14 px-4 bg-slate-900 border-b border-slate-800 flex items-center justify-between gap-4 shrink-0 z-20">
        {/* Brand */}
        <div className="flex items-center gap-3">
          <div className="flex items-center justify-center w-8 h-8 rounded-lg bg-gradient-to-tr from-amber-600 via-rose-600 to-indigo-600 shadow-md shadow-rose-950/40 text-white font-bold">
            <Flame className="w-5 h-5 text-amber-200" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="font-extrabold tracking-tight text-base text-white">
                Tofu<span className="text-amber-400">Forge</span>
              </span>
              <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700">
                v1.1
              </span>
            </div>
            <div className="text-[10px] text-slate-400 hidden sm:block">
              Terraform & OpenTofu Plan Visualizer & Blast-Radius Guard
            </div>
          </div>
        </div>

        {/* Status & Plan Details */}
        <div className="flex items-center gap-2.5">
          <div className={`hidden md:flex items-center gap-1.5 px-2.5 py-1 rounded-full border text-xs font-semibold ${status.className}`}>
            {status.icon}
            <span>{status.text}</span>
          </div>

          {/* Sample Plans Selector */}
          <div className="flex items-center gap-1 bg-slate-950 p-1 rounded-lg border border-slate-800 text-xs">
            <Sparkles className="w-3.5 h-3.5 text-amber-400 ml-1.5" />
            <select
              value={activeSampleId}
              onChange={e => handleSelectSample(e.target.value)}
              className="bg-transparent text-slate-300 font-medium text-xs focus:outline-none cursor-pointer pr-2"
            >
              {activeSampleId === '' && <option value="" className="bg-slate-900">Custom / Ingested Plan</option>}
              {SAMPLE_PLANS.map(s => (
                <option key={s.id} value={s.id} className="bg-slate-900 text-slate-200">
                  {s.title}
                </option>
              ))}
            </select>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-2">
          {/* Plan-to-Plan Diff Button */}
          <button
            onClick={() => setIsDiffModalOpen(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold border border-slate-700 transition"
            title="Compare two plan revisions"
          >
            <GitCompare className="w-4 h-4 text-purple-400" />
            <span className="hidden lg:inline">Compare Diff</span>
          </button>

          {/* Ingest Plan Button */}
          <button
            onClick={() => setIsDropzoneOpen(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold border border-slate-700 transition"
          >
            <UploadCloud className="w-4 h-4 text-blue-400" />
            <span className="hidden sm:inline">Ingest Plan</span>
          </button>

          {/* PR Comment Markdown */}
          <button
            onClick={() => setIsMarkdownModalOpen(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold shadow-sm transition"
          >
            <FileText className="w-4 h-4" />
            <span className="hidden sm:inline">PR Comment</span>
          </button>

          {/* Export Offline HTML Report */}
          <button
            onClick={handleExportHTML}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-200 text-xs font-semibold border border-slate-700 transition"
            title="Download standalone offline HTML report"
          >
            <Download className="w-4 h-4 text-amber-400" />
            <span className="hidden xl:inline">Export HTML</span>
          </button>

          <a
            href="https://github.com/alexandrmotologa/tofuforge"
            target="_blank"
            rel="noreferrer"
            className="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white border border-slate-700 transition"
            title="GitHub Repository"
          >
            <svg className="w-4 h-4 fill-current" viewBox="0 0 24 24">
              <path fillRule="evenodd" clipRule="evenodd" d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.53 1.032 1.53 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
            </svg>
          </a>
        </div>
      </header>

      {/* Stats Summary Bar */}
      <StatCards
        stats={currentPlan.stats}
        activeFilter={actionFilter}
        onFilterAction={setActionFilter}
      />

      {/* Main Workspace Area */}
      <div className="flex-1 flex overflow-hidden relative">
        {/* Left / Center Graph Visualizer Canvas */}
        <div className="flex-1 flex flex-col relative h-full">
          {/* Canvas Sub-toolbar */}
          <div className="h-11 px-4 bg-slate-900/90 border-b border-slate-800/80 flex items-center justify-between gap-3 shrink-0 z-10 backdrop-blur-sm">
            {/* Search Input */}
            <div className="flex items-center gap-2 max-w-sm w-full bg-slate-950 px-2.5 py-1 rounded-md border border-slate-800 text-xs">
              <Search className="w-3.5 h-3.5 text-slate-400 shrink-0" />
              <input
                type="text"
                placeholder="Filter by resource address or type..."
                value={searchQuery}
                onChange={e => setSearchQuery(e.target.value)}
                className="w-full bg-transparent text-slate-200 placeholder-slate-500 focus:outline-none font-mono text-[11px]"
              />
              {searchQuery && (
                <button onClick={() => setSearchQuery('')} className="text-slate-500 hover:text-slate-300">
                  <X className="w-3 h-3" />
                </button>
              )}
            </div>

            {/* Quick Action Filter Pills */}
            <div className="flex items-center gap-1 text-[11px]">
              <span className="text-slate-400 mr-1 hidden md:inline text-[10px] uppercase font-semibold">Action:</span>
              {[
                { id: 'all', label: 'All' },
                { id: 'replace', label: 'Replace', color: 'text-rose-400' },
                { id: 'delete', label: 'Delete', color: 'text-red-400' },
                { id: 'create', label: 'Create', color: 'text-emerald-400' },
                { id: 'update', label: 'Update', color: 'text-amber-400' },
              ].map(f => (
                <button
                  key={f.id}
                  onClick={() => setActionFilter(f.id)}
                  className={`px-2.5 py-0.5 rounded-full capitalize transition font-medium ${
                    actionFilter === f.id
                      ? 'bg-slate-700 text-white shadow-sm'
                      : `bg-slate-800/80 text-slate-400 hover:text-slate-200 ${f.color || ''}`
                  }`}
                >
                  {f.label}
                </button>
              ))}
            </div>

            {/* Selection info / Reset */}
            {selectedNode && (
              <button
                onClick={() => setSelectedNode(null)}
                className="flex items-center gap-1 text-[11px] text-slate-400 hover:text-white px-2 py-0.5 rounded bg-slate-800 border border-slate-700 transition"
              >
                <span>Clear Selection</span>
                <X className="w-3 h-3" />
              </button>
            )}
          </div>

          {/* Graph Component */}
          <div className="flex-1 w-full h-full relative">
            <PlanGraphView
              nodes={currentPlan.nodes}
              edges={currentPlan.edges}
              selectedNode={selectedNode}
              onSelectNode={handleSelectNode}
              actionFilter={actionFilter}
              searchQuery={searchQuery}
            />
          </div>
        </div>

        {/* Right Collapsible Panel */}
        <aside
          className={`border-l border-slate-800 bg-slate-900/95 flex flex-col shrink-0 transition-all duration-200 z-10 ${
            isSidebarOpen ? 'w-96' : 'w-10'
          }`}
        >
          {/* Sidebar Toggle & Tabs Header */}
          <div className="h-11 border-b border-slate-800 flex items-center justify-between px-2 bg-slate-900">
            {isSidebarOpen ? (
              <div className="flex items-center gap-1 flex-1">
                <button
                  onClick={() => setActiveSideTab('destruction')}
                  className={`flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-semibold transition ${
                    activeSideTab === 'destruction'
                      ? 'bg-slate-800 text-white shadow-sm'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <AlertTriangle className="w-3.5 h-3.5 text-amber-400" />
                  <span>Destruction ({currentPlan.destruction_alerts?.length || 0})</span>
                </button>

                <button
                  onClick={() => setActiveSideTab('security')}
                  className={`flex items-center gap-1.5 px-3 py-1 rounded-md text-xs font-semibold transition ${
                    activeSideTab === 'security'
                      ? 'bg-slate-800 text-white shadow-sm'
                      : 'text-slate-400 hover:text-slate-200'
                  }`}
                >
                  <ShieldAlert className="w-3.5 h-3.5 text-red-400" />
                  <span>Security ({currentPlan.security_findings?.length || 0})</span>
                </button>
              </div>
            ) : (
              <div className="w-full flex justify-center text-slate-400">
                <SlidersHorizontal className="w-4 h-4" />
              </div>
            )}

            <button
              onClick={() => setIsSidebarOpen(!isSidebarOpen)}
              className="p-1 rounded text-slate-400 hover:text-white hover:bg-slate-800 transition"
              title={isSidebarOpen ? 'Collapse panel' : 'Expand panel'}
            >
              {isSidebarOpen ? <ChevronRight className="w-4 h-4" /> : <ChevronLeft className="w-4 h-4" />}
            </button>
          </div>

          {/* Sidebar Body */}
          {isSidebarOpen && (
            <div className="flex-1 overflow-y-auto">
              {activeSideTab === 'destruction' ? (
                <DestructionAlerts
                  alerts={currentPlan.destruction_alerts || []}
                  nodes={currentPlan.nodes}
                  onSelectNode={handleSelectNode}
                />
              ) : (
                <SecurityAuditList
                  findings={currentPlan.security_findings || []}
                  nodes={currentPlan.nodes}
                  onSelectNode={handleSelectNode}
                />
              )}
            </div>
          )}
        </aside>
      </div>

      {/* Selected Node Diff Modal */}
      {selectedNode && (
        <ResourceDiffModal
          node={selectedNode}
          onClose={() => setSelectedNode(null)}
        />
      )}

      {/* Ingest Plan Modal */}
      {isDropzoneOpen && (
        <PlanDropzone
          onPlanLoaded={handlePlanUploaded}
          onClose={() => setIsDropzoneOpen(false)}
        />
      )}

      {/* PR Comment Markdown Modal */}
      {isMarkdownModalOpen && (
        <PRMarkdownModal
          plan={currentPlan}
          onClose={() => setIsMarkdownModalOpen(false)}
        />
      )}

      {/* Plan-to-Plan Diff Modal */}
      {isDiffModalOpen && (
        <PlanDiffModal
          currentPlan={currentPlan}
          onApplyPlan={handlePlanUploaded}
          onClose={() => setIsDiffModalOpen(false)}
        />
      )}
    </div>
  );
};

export default App;
