import type { ReactNode } from 'react';
import type { LucideIcon } from 'lucide-react';
import { Plus, Trash2 } from 'lucide-react';
import { Button, Panel, DecryptText } from '@maze/fabrication';
import { clipPathHalf } from '@maze/fabrication';
import type { FabricationEntity } from './types';

interface FabricationListColumnProps<T extends FabricationEntity> {
  icon: LucideIcon;
  label: string;
  emptyText: string;
  items: T[];
  isCreating: boolean;
  editingItem: T | null;
  selectedName?: string | null;
  renderItemSubtitle: (item: T) => ReactNode;
  renderItemExtra?: (item: T) => ReactNode;
  onCreate: () => void;
  onEdit: (item: T) => void;
  onDeleteClick: (name: string) => void;
}

export function FabricationListColumn<T extends FabricationEntity>({
  icon: Icon,
  label,
  emptyText,
  items,
  isCreating,
  editingItem,
  selectedName,
  renderItemSubtitle,
  renderItemExtra,
  onCreate,
  onEdit,
  onDeleteClick,
}: FabricationListColumnProps<T>) {
  return (
    <div className="border-r border-border/50 flex flex-col bg-background/50 relative z-10 overflow-hidden min-w-[320px]">
      <div className="absolute right-0 top-0 w-[1px] h-full bg-gradient-to-b from-primary/20 to-transparent" />
      <Panel className="flex flex-col h-full relative m-2" cornerSize={16}>
        <div className="pb-4 border-b border-primary/20 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-primary">
              <Icon className="w-4 h-4" />
              <h2 className="text-xs font-bold uppercase tracking-widest">
                <DecryptText text={label} />
              </h2>
            </div>
            <Button
              variant="ghost"
              size="icon"
              className="text-primary/60 hover:text-primary hover:bg-primary/20 rounded-none"
              onClick={onCreate}
            >
              <Plus className="w-4 h-4" />
            </Button>
          </div>
        </div>

        <div className="flex-1 overflow-y-auto pt-4 space-y-2">
          {items.length === 0 && !isCreating ? (
            <div className="text-center text-primary/30 text-[10px] uppercase tracking-widest py-12 font-mono">
              [ {emptyText} ]
            </div>
          ) : (
            items.map((item) => {
              const isSelected =
                selectedName != null ? selectedName === item.name : editingItem?.name === item.name;
              return (
                <div
                  key={item.name}
                  onClick={() => onEdit(item)}
                  className={`group flex items-center justify-between p-3 border-l-2
                    transition-all cursor-pointer backdrop-blur-sm
                    ${
                      isSelected
                        ? 'bg-primary/20 border-primary shadow-[0_0_15px_rgba(0,255,255,0.2)]'
                        : 'bg-black/40 border-primary/30 hover:border-primary/60 hover:bg-primary/10'
                    }`}
                  style={{ clipPath: clipPathHalf(8) }}
                >
                  <div className="flex-1 min-w-0 pl-1">
                    <div className="flex flex-col overflow-hidden">
                      <span className="text-sm font-mono font-bold tracking-wide uppercase truncate text-primary">
                        {isSelected ? <DecryptText text={item.name!} /> : item.name}
                      </span>
                      <span className="text-[10px] text-primary/50 font-mono uppercase tracking-widest truncate mt-0.5">
                        {renderItemSubtitle(item)}
                      </span>
                      {renderItemExtra?.(item)}
                    </div>
                  </div>
                  <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity pr-1">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 rounded-none text-destructive/60 hover:text-destructive hover:bg-destructive/20"
                      onClick={(e) => {
                        e.stopPropagation();
                        onDeleteClick(item.name!);
                      }}
                    >
                      <Trash2 className="w-3 h-3" />
                    </Button>
                  </div>
                </div>
              );
            })
          )}
        </div>
      </Panel>
    </div>
  );
}
