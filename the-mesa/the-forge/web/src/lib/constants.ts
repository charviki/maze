export const typeColorMap: Record<string, { bg: string; text: string; border: string }> = {
  requirement: { bg: 'bg-primary/15', text: 'text-primary', border: 'border-primary/30' },
  shared: { bg: 'bg-blue-500/15', text: 'text-blue-400', border: 'border-blue-500/30' },
  ops: { bg: 'bg-green-500/15', text: 'text-green-500', border: 'border-green-500/30' },
  narrative: { bg: 'bg-purple-400/15', text: 'text-purple-400', border: 'border-purple-400/30' },
  memory: {
    bg: 'bg-muted-foreground/15',
    text: 'text-muted-foreground',
    border: 'border-muted-foreground/30',
  },
};

export const folderColorMap = {
  bg: 'bg-primary/15',
  text: 'text-primary',
  border: 'border-primary/30',
};

export function getTypeColor(type: string) {
  return (
    typeColorMap[type] || {
      bg: 'bg-background',
      text: 'text-muted-foreground',
      border: 'border-border',
    }
  );
}

export const statusColors: Record<string, { bg: string; text: string }> = {
  pending: { bg: 'bg-amber-500/15', text: 'text-amber-400' },
  in_progress: { bg: 'bg-blue-500/15', text: 'text-blue-400' },
  completed: { bg: 'bg-green-500/15', text: 'text-green-500' },
  cancelled: { bg: 'bg-muted-foreground/15', text: 'text-muted-foreground' },
};

export const priorityColors: Record<string, { bg: string; text: string }> = {
  high: { bg: 'bg-red-500/15', text: 'text-red-400' },
  medium: { bg: 'bg-amber-500/15', text: 'text-amber-400' },
  low: { bg: 'bg-green-500/15', text: 'text-green-500' },
};
