import { memo, type ReactNode } from 'react';
import { Plus, Trash2, Search, TerminalSquare, RotateCcw, Eye, Settings } from 'lucide-react';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Panel } from '../ui/Panel';
import { DecryptText } from '../ui/DecryptText';
import { Skeleton } from '../ui/Skeleton';
import { HostVitalSign } from '../ui/HostVitalSign';
import { ReverieEffect } from '../ui/ReverieEffect';

export interface SessionDisplay {
  id: string;
  name: string;
  status: 'running' | 'saved' | 'offline';
  createdAt: string;
  windowCount: number;
  savedAt?: string;
  terminalSnapshot?: string;
}

export interface SessionListProps {
  sessions: SessionDisplay[];
  search: string;
  onSearchChange: (v: string) => void;
  selectedSessionId: string | null;
  onSelectSession: (id: string) => void;
  nodeName: string;
  onCreateClick: () => void;
  onNodeConfigClick: () => void;
  onKill: (session: SessionDisplay) => void;
  onRestore: (session: SessionDisplay) => void;
  onViewPipeline: (e: React.MouseEvent, session: SessionDisplay) => void;
  listHeaderActions?: ReactNode;
  isLoading?: boolean;
}

export const SessionList = memo(function SessionList({
  sessions,
  search,
  onSearchChange,
  selectedSessionId,
  onSelectSession,
  nodeName,
  onCreateClick,
  onNodeConfigClick,
  onKill,
  onRestore,
  onViewPipeline,
  listHeaderActions,
  isLoading = false,
}: SessionListProps) {
  const filteredSessions = sessions.filter(
    (s) =>
      s.name.toLowerCase().includes(search.toLowerCase()) ||
      s.id.toLowerCase().includes(search.toLowerCase()),
  );

  return (
    <div
      data-slot="session-list"
      className="border-r border-border/50 flex flex-col bg-background/50 relative z-10 overflow-hidden min-w-[288px]"
    >
      <div className="absolute right-0 top-0 w-[1px] h-full bg-gradient-to-b from-primary/20 to-transparent"></div>
      <Panel className="flex flex-col h-full relative m-2" cornerSize={16} showCrosshairs={true}>
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(0,255,255,0.03)_0,transparent_100%)] pointer-events-none"></div>
        <div className="pb-4 border-b border-primary/20 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2 text-primary">
              <TerminalSquare className="w-4 h-4" />
              <h2 className="text-xs font-bold uppercase tracking-widest">
                <DecryptText text={`${nodeName} // Narrative Loops`} />
              </h2>
            </div>
            <div className="flex items-center gap-1">
              {listHeaderActions}
              <Button
                variant="ghost"
                size="icon"
                className="text-primary hover:text-primary hover:bg-primary/20 rounded-none border border-transparent hover:border-primary/30 transition-all"
                onClick={onNodeConfigClick}
                title="Node Config"
              >
                <Settings className="w-4 h-4" />
              </Button>
              <Button
                variant="ghost"
                size="icon"
                className="text-primary hover:text-primary hover:bg-primary/20 rounded-none border border-transparent hover:border-primary/30 transition-all"
                onClick={onCreateClick}
                title="Add Loop"
              >
                <Plus className="w-4 h-4" />
              </Button>
            </div>
          </div>

          <div className="relative">
            <Search className="w-4 h-4 absolute left-2.5 top-2.5 text-primary/50" />
            <Input
              placeholder="SEARCH LOOPS..."
              value={search}
              onChange={(e) => {
                onSearchChange(e.target.value);
              }}
              className="pl-8 h-9 text-xs font-mono uppercase tracking-widest bg-black/60 backdrop-blur-md border-primary/30 text-primary placeholder:text-primary/30 focus-visible:ring-primary/50 rounded-none"
              style={{
                clipPath:
                  'polygon(0 0, calc(100% - 6px) 0, 100% 6px, 100% 100%, 6px 100%, 0 calc(100% - 6px))',
              }}
            />
          </div>
        </div>

        <div className="flex-1 overflow-y-auto pt-4 space-y-2 relative z-10 pr-1 custom-scrollbar">
          {isLoading &&
            filteredSessions.length === 0 &&
            Array.from({ length: 3 }).map((_, i) => (
              <div
                key={i}
                className="flex items-center gap-3 p-3 pl-4 bg-black/40 border-l-2 border-primary/30 backdrop-blur-sm"
                style={{
                  clipPath:
                    'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))',
                }}
              >
                <Skeleton className="w-1.5 h-3 bg-primary/10" />
                <div className="flex flex-col gap-1.5">
                  <Skeleton className="h-3 w-28" />
                  <Skeleton className="h-2 w-20 bg-primary/5" />
                </div>
              </div>
            ))}
          {filteredSessions.map((session) => {
            const isSelected = selectedSessionId === session.id;
            const isSaved = session.status === 'saved';
            const isRunning = session.status === 'running';
            return (
              <ReverieEffect key={session.id} isActive={isRunning}>
                <div
                  onClick={() => !isSaved && onSelectSession(session.id)}
                  className={`group flex items-center justify-between p-3 border-l-2 transition-all relative overflow-hidden backdrop-blur-sm ${
                    isSaved ? 'cursor-default' : 'cursor-pointer'
                  } ${
                    isSelected
                      ? 'bg-primary/20 border-primary shadow-[0_0_15px_rgba(0,255,255,0.2)]'
                      : 'bg-black/40 border-primary/30 hover:border-primary/60 hover:bg-primary/10'
                  }`}
                  style={{
                    clipPath:
                      'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))',
                  }}
                >
                  {isSelected && (
                    <div
                      className="absolute right-0 top-0 w-[1px] h-full bg-primary/50"
                      style={{ animation: 'organic-pulse 2s ease-in-out infinite' }}
                    ></div>
                  )}
                  {/* running 状态的背景呼吸光效 */}
                  {isRunning && (
                    <div
                      className="absolute inset-0 pointer-events-none"
                      style={{
                        background:
                          'radial-gradient(ellipse at 12px 50%, rgba(0,255,255,0.06) 0%, transparent 60%)',
                        animation: 'organic-pulse 3s ease-in-out infinite',
                      }}
                    />
                  )}

                  <div className="flex items-center gap-3 overflow-hidden pl-1">
                    <HostVitalSign status={session.status} />
                    <div className="flex flex-col overflow-hidden">
                      <div className="flex items-center gap-2">
                        <span
                          className={`text-sm font-mono font-bold tracking-wide uppercase truncate ${isSelected ? 'text-primary' : 'text-primary/80'}`}
                        >
                          {isSelected ? <DecryptText text={session.name} /> : session.name}
                        </span>
                        {isSaved && (
                          <span
                            className="text-[9px] font-mono px-1.5 py-0.5 border border-yellow-500/50 bg-yellow-500/10 text-yellow-500 uppercase tracking-widest whitespace-nowrap shrink-0"
                            style={{
                              clipPath:
                                'polygon(0 0, calc(100% - 4px) 0, 100% 4px, 100% 100%, 4px 100%, 0 calc(100% - 4px))',
                            }}
                          >
                            [ SAVED ]
                          </span>
                        )}
                      </div>
                      <span className="text-[10px] text-primary/50 font-mono uppercase tracking-widest truncate mt-0.5">
                        {isSaved && session.savedAt
                          ? `ARCHIVED: ${new Date(session.savedAt).toLocaleString()}`
                          : `[ LOOP_ID: ${session.id.slice(0, 8)} ]`}
                      </span>
                    </div>
                  </div>
                  <div className="flex items-center gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
                    {isSaved && (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-primary/50 hover:text-blue-400 hover:bg-blue-400/20 rounded-none border border-transparent hover:border-blue-400/30"
                        onClick={(e) => {
                          e.stopPropagation();
                          onRestore(session);
                        }}
                        title="RESTORE"
                      >
                        <RotateCcw className="w-3.5 h-3.5" />
                      </Button>
                    )}
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-primary/50 hover:text-primary hover:bg-primary/20 rounded-none border border-transparent hover:border-primary/30"
                      onClick={(e) => {
                        onViewPipeline(e, session);
                      }}
                      title="VIEW PIPELINE"
                    >
                      <Eye className="w-3.5 h-3.5" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-primary/50 hover:text-destructive hover:bg-destructive/20 rounded-none border border-transparent hover:border-destructive/30"
                      onClick={(e) => {
                        e.stopPropagation();
                        onKill(session);
                      }}
                      title="TERMINATE"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                    </Button>
                  </div>
                </div>
              </ReverieEffect>
            );
          })}
          {!isLoading && filteredSessions.length === 0 && (
            <div
              className="text-center p-8 text-xs font-mono text-primary/40 uppercase tracking-widest animate-pulse border border-primary/10 bg-black/40 backdrop-blur-sm"
              style={{
                clipPath:
                  'polygon(0 0, calc(100% - 8px) 0, 100% 8px, 100% 100%, 8px 100%, 0 calc(100% - 8px))',
              }}
            >
              [ NO LOOPS ACTIVE ]
            </div>
          )}
        </div>
      </Panel>
    </div>
  );
});
