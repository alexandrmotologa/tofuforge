import React from 'react';
import { 
  PlusCircle, 
  RefreshCw, 
  Trash2, 
  Flame, 
  Radio, 
  AlertOctagon, 
  ShieldAlert,
  DollarSign
} from 'lucide-react';
import type { SummaryStats } from '../types';

interface StatCardsProps {
  stats: SummaryStats;
  onFilterAction?: (action: string) => void;
  activeFilter?: string;
}

export const StatCards: React.FC<StatCardsProps> = ({ 
  stats, 
  onFilterAction, 
  activeFilter = 'all' 
}) => {
  const hasCost = stats.total_monthly_delta !== undefined;
  const costDelta = stats.total_monthly_delta || 0;

  const cards = [
    {
      id: 'create',
      label: 'Create',
      count: stats.create_count,
      prefix: '+',
      icon: PlusCircle,
      textColor: 'text-emerald-400',
      bgColor: 'bg-emerald-950/20',
      borderColor: 'border-emerald-500/20',
      activeBorder: 'ring-2 ring-emerald-500',
    },
    {
      id: 'update',
      label: 'Update',
      count: stats.update_count,
      prefix: '~',
      icon: RefreshCw,
      textColor: 'text-amber-400',
      bgColor: 'bg-amber-950/20',
      borderColor: 'border-amber-500/20',
      activeBorder: 'ring-2 ring-amber-500',
    },
    {
      id: 'delete',
      label: 'Delete',
      count: stats.delete_count,
      prefix: '-',
      icon: Trash2,
      textColor: 'text-red-400',
      bgColor: 'bg-red-950/20',
      borderColor: 'border-red-500/20',
      activeBorder: 'ring-2 ring-red-500',
    },
    {
      id: 'replace',
      label: 'Replace',
      count: stats.replace_count,
      prefix: '±',
      icon: Flame,
      textColor: 'text-rose-400',
      bgColor: 'bg-rose-950/30',
      borderColor: 'border-rose-500/30',
      activeBorder: 'ring-2 ring-rose-500',
      pulse: stats.replace_count > 0,
    },
    {
      id: 'blast',
      label: 'Blast Radius',
      count: stats.total_blast_radius,
      prefix: '',
      icon: Radio,
      textColor: stats.total_blast_radius > 20 ? 'text-amber-400' : 'text-slate-300',
      bgColor: 'bg-slate-900/60',
      borderColor: 'border-slate-800',
    },
    {
      id: 'alerts',
      label: 'Destruction Alerts',
      count: stats.critical_alerts_count,
      prefix: '',
      icon: AlertOctagon,
      textColor: stats.critical_alerts_count > 0 ? 'text-rose-400' : 'text-slate-400',
      bgColor: stats.critical_alerts_count > 0 ? 'bg-rose-950/20' : 'bg-slate-900/60',
      borderColor: stats.critical_alerts_count > 0 ? 'border-rose-500/30' : 'border-slate-800',
    },
    {
      id: 'security',
      label: 'Security Flaws',
      count: stats.security_issues_count,
      prefix: '',
      icon: ShieldAlert,
      textColor: stats.security_issues_count > 0 ? 'text-red-400' : 'text-slate-400',
      bgColor: stats.security_issues_count > 0 ? 'bg-red-950/20' : 'bg-slate-900/60',
      borderColor: stats.security_issues_count > 0 ? 'border-red-500/30' : 'border-slate-800',
    },
    ...(hasCost ? [{
      id: 'cost',
      label: 'Cost Delta',
      count: `${costDelta > 0 ? '+' : costDelta < 0 ? '-' : ''}$${Math.abs(costDelta).toFixed(2)}/mo`,
      prefix: '',
      icon: DollarSign,
      textColor: costDelta > 0 ? 'text-amber-400' : costDelta < 0 ? 'text-emerald-400' : 'text-slate-300',
      bgColor: costDelta > 0 ? 'bg-amber-950/20' : costDelta < 0 ? 'bg-emerald-950/20' : 'bg-slate-900/60',
      borderColor: costDelta > 0 ? 'border-amber-500/30' : costDelta < 0 ? 'border-emerald-500/30' : 'border-slate-800',
    }] : []),
  ];

  return (
    <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-8 gap-2.5 p-3 bg-slate-900/80 border-b border-slate-800/80 backdrop-blur-md">
      {cards.map(c => {
        const Icon = c.icon;
        const isClickable = onFilterAction && ['create', 'update', 'delete', 'replace'].includes(c.id);
        const isSelected = activeFilter === c.id;

        return (
          <div
            key={c.id}
            onClick={() => isClickable && onFilterAction(isSelected ? 'all' : c.id)}
            className={`flex items-center gap-2.5 px-3 py-2 rounded-lg border ${c.borderColor} ${c.bgColor} transition-all duration-150 ${
              isClickable ? 'cursor-pointer hover:bg-slate-800/60' : ''
            } ${isSelected ? c.activeBorder : ''}`}
          >
            <div className={`p-1.5 rounded-md bg-slate-950/50 ${c.textColor}`}>
              <Icon className={`w-4 h-4 ${c.pulse ? 'animate-pulse' : ''}`} />
            </div>
            <div className="min-w-0">
              <div className="text-[11px] font-medium text-slate-400 uppercase tracking-wider truncate">
                {c.label}
              </div>
              <div className={`text-sm sm:text-base font-bold font-mono truncate ${c.textColor}`}>
                {c.prefix}{c.count}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};
