// Westworld 主题配色：暖色沙石 + 冷色金属
// status: active/done/pending/failed — 体现任务/待办状态

export const statusColors: Record<string, { bg: string; text: string; border: string }> = {
  active: { bg: 'bg-amber-500/15', text: 'text-amber-400', border: 'border-amber-500/30' },
  pending: { bg: 'bg-slate-500/15', text: 'text-slate-400', border: 'border-slate-500/30' },
  done: { bg: 'bg-emerald-500/15', text: 'text-emerald-400', border: 'border-emerald-500/30' },
  failed: { bg: 'bg-red-500/15', text: 'text-red-400', border: 'border-red-500/30' },
};

export const priorityColors: Record<string, { bg: string; text: string }> = {
  critical: { bg: 'bg-red-500/15', text: 'text-red-400' },
  high: { bg: 'bg-amber-500/15', text: 'text-amber-400' },
  medium: { bg: 'bg-blue-500/15', text: 'text-blue-400' },
  normal: { bg: 'bg-slate-500/15', text: 'text-slate-400' },
  low: { bg: 'bg-slate-500/10', text: 'text-slate-500' },
};

export function getStatusColor(status: string | undefined): {
  bg: string;
  text: string;
  border: string;
} {
  if (!status) {
    return {
      bg: 'bg-background',
      text: 'text-muted-foreground',
      border: 'border-border',
    };
  }
  return (
    statusColors[status] ?? {
      bg: 'bg-background',
      text: 'text-muted-foreground',
      border: 'border-border',
    }
  );
}
