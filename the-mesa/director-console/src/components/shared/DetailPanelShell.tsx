import type { ReactNode } from 'react';
import type { LucideIcon } from 'lucide-react';
import { Panel, Button, DecryptText } from '@maze/fabrication';

interface DetailPanelShellProps {
  icon: LucideIcon;
  emptyText: string;
  createTitle: string;
  editTitle: string;
  item: { name?: string } | null;
  isCreating: boolean;
  submitting: boolean;
  canSubmit: boolean;
  canEdit?: boolean;
  readOnlyView?: ReactNode;
  onCancel: () => void;
  onSubmit: () => void;
  children: ReactNode;
}

export function DetailPanelShell({
  icon: Icon,
  emptyText,
  createTitle,
  editTitle,
  item,
  isCreating,
  submitting,
  canSubmit,
  canEdit = true,
  readOnlyView,
  onCancel,
  onSubmit,
  children,
}: DetailPanelShellProps) {
  if (!item && !isCreating) {
    return (
      <div className="flex-1 flex items-center justify-center text-primary/40 uppercase tracking-widest text-xs">
        <div className="text-center space-y-4">
          <Icon className="w-16 h-16 mx-auto opacity-20" />
          <DecryptText text={emptyText} />
        </div>
      </div>
    );
  }

  if (canEdit === false && item && !isCreating) {
    return (
      <div className="flex-1 min-w-0 flex flex-col bg-background relative z-10 overflow-hidden">
        <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
          <div className="pb-4 border-b border-primary/20">
            <div className="flex items-center gap-2 text-primary">
              <Icon className="w-4 h-4" />
              <h2 className="text-xs font-bold uppercase tracking-widest">
                <DecryptText text={editTitle} />
              </h2>
            </div>
          </div>
          <div className="flex-1 overflow-y-auto pt-4 space-y-4 px-1">{readOnlyView}</div>
        </Panel>
      </div>
    );
  }

  return (
    <div className="flex-1 min-w-0 flex flex-col bg-background relative z-10 overflow-hidden">
      <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
        <div className="pb-4 border-b border-primary/20">
          <div className="flex items-center gap-2 text-primary">
            <Icon className="w-4 h-4" />
            <h2 className="text-xs font-bold uppercase tracking-widest">
              <DecryptText text={isCreating ? createTitle : editTitle} />
            </h2>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto pt-4 space-y-4 px-1">{children}</div>

        <div className="pt-4 border-t border-primary/20 flex justify-end gap-2">
          <Button
            variant="ghost"
            onClick={onCancel}
            className="font-mono uppercase tracking-widest text-xs rounded-none"
          >
            CANCEL
          </Button>
          <Button
            onClick={onSubmit}
            disabled={!canSubmit || submitting}
            className="font-mono uppercase tracking-widest text-xs rounded-none bg-primary hover:bg-primary/90 text-primary-foreground"
          >
            {submitting ? 'SAVING...' : isCreating ? 'FABRICATE' : 'COMMIT'}
          </Button>
        </div>
      </Panel>
    </div>
  );
}
