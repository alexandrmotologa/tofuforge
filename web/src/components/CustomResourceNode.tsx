import { memo } from 'react';
import { Handle, Position } from '@xyflow/react';
import { 
  Database, 
  HardDrive, 
  Network, 
  Cpu, 
  ShieldAlert, 
  Globe, 
  Box, 
  AlertTriangle,
  Flame,
  PlusCircle,
  RefreshCw,
  Trash2,
  CheckCircle2,
  DollarSign
} from 'lucide-react';
import type { ResourceNode } from '../types';

interface CustomResourceNodeProps {
  data: {
    node: ResourceNode;
    isHighlighted?: boolean;
    isDownstream?: boolean;
    isSelected?: boolean;
    onSelect?: (node: ResourceNode) => void;
  };
}

export const CustomResourceNode = memo(({ data }: CustomResourceNodeProps) => {
  const { node, isHighlighted, isDownstream, isSelected } = data;

  const getCategoryIcon = () => {
    switch (node.category) {
      case 'database':
        return <Database className="w-4 h-4 text-purple-400" />;
      case 'storage':
        return <HardDrive className="w-4 h-4 text-cyan-400" />;
      case 'network':
        return <Network className="w-4 h-4 text-blue-400" />;
      case 'compute':
        return <Cpu className="w-4 h-4 text-amber-400" />;
      case 'security':
        return <ShieldAlert className="w-4 h-4 text-rose-400" />;
      case 'dns':
        return <Globe className="w-4 h-4 text-emerald-400" />;
      default:
        return <Box className="w-4 h-4 text-slate-400" />;
    }
  };

  const getActionBadge = () => {
    switch (node.action) {
      case 'create':
        return (
          <span className="flex items-center gap-1 text-[10px] font-semibold tracking-wider uppercase px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
            <PlusCircle className="w-3 h-3" /> Create
          </span>
        );
      case 'update':
        return (
          <span className="flex items-center gap-1 text-[10px] font-semibold tracking-wider uppercase px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/30">
            <RefreshCw className="w-3 h-3" /> Update
          </span>
        );
      case 'delete':
        return (
          <span className="flex items-center gap-1 text-[10px] font-semibold tracking-wider uppercase px-2 py-0.5 rounded-full bg-red-500/10 text-red-400 border border-red-500/30">
            <Trash2 className="w-3 h-3" /> Delete
          </span>
        );
      case 'replace':
        return (
          <span className="flex items-center gap-1 text-[10px] font-semibold tracking-wider uppercase px-2 py-0.5 rounded-full bg-rose-500/20 text-rose-300 border border-rose-500/50 shadow-sm shadow-rose-950">
            <Flame className="w-3 h-3 text-rose-400 animate-pulse" /> Replace
          </span>
        );
      case 'no-op':
        return (
          <span className="flex items-center gap-1 text-[10px] font-medium tracking-wider uppercase px-2 py-0.5 rounded-full bg-slate-800 text-slate-400 border border-slate-700">
            <CheckCircle2 className="w-3 h-3" /> No-Op
          </span>
        );
      default:
        return (
          <span className="text-[10px] font-medium tracking-wider uppercase px-2 py-0.5 rounded-full bg-slate-800 text-slate-400">
            {node.action}
          </span>
        );
    }
  };

  let borderClasses = "border-slate-800 hover:border-slate-600 bg-slate-900/90";
  if (isSelected) {
    borderClasses = "border-blue-500 ring-2 ring-blue-500/40 bg-slate-900 shadow-lg shadow-blue-950/50";
  } else if (isDownstream) {
    borderClasses = "border-amber-500/80 ring-2 ring-amber-500/40 bg-slate-900 shadow-lg shadow-amber-950/50";
  } else if (node.action === 'replace') {
    borderClasses = "border-rose-500/70 ring-1 ring-rose-500/30 bg-slate-900/95";
  } else if (node.action === 'delete') {
    borderClasses = "border-red-500/60 bg-slate-900/95";
  } else if (node.action === 'create') {
    borderClasses = "border-emerald-500/40 bg-slate-900/95";
  }

  const opacityClass = (isHighlighted === false) ? "opacity-35 transition-opacity" : "opacity-100 transition-opacity";

  return (
    <div 
      className={`relative min-w-[260px] max-w-[320px] rounded-xl border p-3 shadow-xl backdrop-blur-sm transition-all duration-200 cursor-pointer ${borderClasses} ${opacityClass}`}
      onClick={() => data.onSelect && data.onSelect(node)}
    >
      <Handle type="target" position={Position.Top} className="!w-2.5 !h-2.5 !bg-slate-400 !border-2 !border-slate-900" />

      {/* Header bar */}
      <div className="flex items-center justify-between gap-2 mb-2 pb-2 border-b border-slate-800/80">
        <div className="flex items-center gap-1.5 min-w-0">
          <div className="p-1 rounded bg-slate-800 border border-slate-700/60">
            {getCategoryIcon()}
          </div>
          <span className="text-xs font-mono font-medium text-slate-300 truncate">
            {node.type}
          </span>
        </div>
        {getActionBadge()}
      </div>

      {/* Resource Address & Name */}
      <div className="mb-2.5">
        <div className="text-sm font-semibold text-slate-100 truncate font-mono" title={node.address}>
          {node.name}
        </div>
        <div className="text-[11px] text-slate-400 font-mono truncate" title={node.address}>
          {node.address}
        </div>
      </div>

      {/* Badges / Metrics Footer */}
      <div className="flex flex-wrap items-center gap-1.5 pt-1 text-[11px]">
        {node.is_stateful && (
          <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-300 border border-amber-500/30 font-medium">
            <AlertTriangle className="w-3 h-3 text-amber-400" /> Stateful
          </span>
        )}

        {node.blast_radius_score > 0 && (
          <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-rose-500/10 text-rose-300 border border-rose-500/30 font-medium">
            Blast: {node.blast_radius_score}
          </span>
        )}

        {node.downstream_count > 0 && (
          <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
            Downstream: {node.downstream_count}
          </span>
        )}

        {node.cost && node.cost.monthly_delta !== undefined && Math.abs(node.cost.monthly_delta) > 0.001 && (
          <span className={`inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded font-mono font-medium ${
            node.cost.monthly_delta > 0 
              ? 'bg-amber-500/10 text-amber-300 border border-amber-500/30' 
              : 'bg-emerald-500/10 text-emerald-300 border border-emerald-500/30'
          }`}>
            <DollarSign className="w-3 h-3" />
            {node.cost.monthly_delta > 0 ? '+' : '-'}${Math.abs(node.cost.monthly_delta).toFixed(2)}/mo
          </span>
        )}

        {node.security_findings && node.security_findings.length > 0 && (
          <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded bg-red-500/20 text-red-300 border border-red-500/40 font-medium">
            <ShieldAlert className="w-3 h-3 text-red-400" /> {node.security_findings.length} flaw{node.security_findings.length > 1 ? 's' : ''}
          </span>
        )}
      </div>

      <Handle type="source" position={Position.Bottom} className="!w-2.5 !h-2.5 !bg-slate-400 !border-2 !border-slate-900" />
    </div>
  );
});
CustomResourceNode.displayName = 'CustomResourceNode';
