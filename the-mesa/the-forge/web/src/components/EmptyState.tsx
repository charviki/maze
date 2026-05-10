import { FileText } from 'lucide-react';

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  action?: {
    label: string;
    onClick: () => void;
  };
}

export default function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-16 text-center">
      <div className="w-12 h-12 rounded-full bg-primary/10 flex items-center justify-center mb-4">
        {icon || <FileText size={20} className="text-primary" />}
      </div>
      <h3 className="text-foreground font-mono text-sm mb-1">{title}</h3>
      {description && (
        <p className="text-muted-foreground font-mono text-xs max-w-sm">{description}</p>
      )}
      {action && (
        <button
          onClick={action.onClick}
          className="mt-4 px-4 py-2 bg-primary text-primary-foreground font-mono text-sm rounded hover:bg-primary/90 transition"
        >
          {action.label}
        </button>
      )}
    </div>
  );
}
