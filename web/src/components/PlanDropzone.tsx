import React, { useState, useRef } from 'react';
import { UploadCloud, FileText, AlertCircle, X } from 'lucide-react';
import type { AnalyzedPlan } from '../types';

interface PlanDropzoneProps {
  onPlanLoaded: (plan: AnalyzedPlan) => void;
  onClose: () => void;
}

export const PlanDropzone: React.FC<PlanDropzoneProps> = ({ onPlanLoaded, onClose }) => {
  const [isDragging, setIsDragging] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [jsonText, setJsonText] = useState('');
  const [activeTab, setActiveTab] = useState<'upload' | 'paste'>('upload');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const processPlanJSON = async (content: string) => {
    setLoading(true);
    setError(null);

    try {
      // 1. Try uploading to backend API /api/upload
      const response = await fetch('/api/upload', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: content,
      });

      if (response.ok) {
        const analyzed: AnalyzedPlan = await response.json();
        onPlanLoaded(analyzed);
        onClose();
        return;
      }
    } catch {
      // Backend not running (e.g. static preview). Fallback to client-side parsing!
    }

    try {
      const raw = JSON.parse(content);
      // If the file is already an analyzed plan JSON (e.g. from tofuforge inspect --json)
      if (raw.nodes && raw.stats) {
        onPlanLoaded(raw as AnalyzedPlan);
        onClose();
        return;
      }

      // Basic validation for raw Terraform/OpenTofu plan JSON
      if (!raw.resource_changes && !raw.format_version && !raw.planned_values) {
        throw new Error('JSON is missing Terraform/OpenTofu plan properties (resource_changes or planned_values)');
      }

      setError('Backend API /api/upload not reachable to analyze raw plan. Please run with "tofuforge serve" or select a pre-loaded sample.');
    } catch (err: any) {
      setError(err.message || 'Failed to parse JSON file');
    } finally {
      setLoading(false);
    }
  };

  const handleFile = (file: File) => {
    if (!file.name.endsWith('.json')) {
      setError('Please provide a valid .json plan file');
      return;
    }

    const reader = new FileReader();
    reader.onload = e => {
      const text = e.target?.result as string;
      if (text) processPlanJSON(text);
    };
    reader.onerror = () => setError('Failed to read file');
    reader.readAsText(file);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragging(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      handleFile(e.dataTransfer.files[0]);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-md">
      <div className="w-full max-w-lg rounded-2xl bg-slate-900 border border-slate-800 shadow-2xl p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <div className="p-2 rounded-lg bg-blue-500/10 text-blue-400 border border-blue-500/20">
              <UploadCloud className="w-5 h-5" />
            </div>
            <div>
              <h3 className="text-base font-bold text-slate-100">Ingest Terraform / OpenTofu Plan</h3>
              <p className="text-xs text-slate-400">Upload JSON plan generated from `tofu show -json`</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tab switch */}
        <div className="flex rounded-lg bg-slate-950 p-1 mb-4 border border-slate-800">
          <button
            onClick={() => setActiveTab('upload')}
            className={`flex-1 py-1.5 text-xs font-semibold rounded-md transition ${
              activeTab === 'upload' ? 'bg-slate-800 text-white' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            File Upload (.json)
          </button>
          <button
            onClick={() => setActiveTab('paste')}
            className={`flex-1 py-1.5 text-xs font-semibold rounded-md transition ${
              activeTab === 'paste' ? 'bg-slate-800 text-white' : 'text-slate-400 hover:text-slate-200'
            }`}
          >
            Paste JSON
          </button>
        </div>

        {activeTab === 'upload' ? (
          <div
            onDragOver={e => {
              e.preventDefault();
              setIsDragging(true);
            }}
            onDragLeave={() => setIsDragging(false)}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            className={`flex flex-col items-center justify-center p-8 rounded-xl border-2 border-dashed transition cursor-pointer ${
              isDragging
                ? 'border-blue-500 bg-blue-500/10'
                : 'border-slate-800 hover:border-slate-700 bg-slate-950/40'
            }`}
          >
            <input
              type="file"
              ref={fileInputRef}
              accept=".json"
              className="hidden"
              onChange={e => e.target.files?.[0] && handleFile(e.target.files[0])}
            />
            <FileText className="w-10 h-10 text-slate-400 mb-3" />
            <div className="text-sm font-semibold text-slate-200 mb-1">
              Drag & Drop your tfplan.json here
            </div>
            <div className="text-xs text-slate-400">or click to browse from your machine</div>
          </div>
        ) : (
          <div className="space-y-3">
            <textarea
              rows={8}
              value={jsonText}
              onChange={e => setJsonText(e.target.value)}
              placeholder="Paste raw Terraform / OpenTofu plan JSON here..."
              className="w-full rounded-xl bg-slate-950 border border-slate-800 p-3 font-mono text-xs text-slate-300 focus:outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
            />
            <button
              onClick={() => processPlanJSON(jsonText)}
              disabled={!jsonText.trim() || loading}
              className="w-full py-2.5 rounded-xl bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-xs font-semibold transition"
            >
              {loading ? 'Analyzing Plan...' : 'Analyze Plan'}
            </button>
          </div>
        )}

        {error && (
          <div className="mt-3 flex items-start gap-2 p-3 rounded-xl bg-red-950/20 border border-red-500/30 text-xs text-red-300">
            <AlertCircle className="w-4 h-4 text-red-400 shrink-0 mt-0.5" />
            <span>{error}</span>
          </div>
        )}
      </div>
    </div>
  );
};
