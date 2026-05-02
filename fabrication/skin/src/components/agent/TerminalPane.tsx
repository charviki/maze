import { memo, type ReactNode } from 'react';
import { TerminalSquare, Terminal as TerminalIcon, Info, Save, AlertCircle } from 'lucide-react';
import { Button } from '../ui/button';
import { Panel } from '../ui/Panel';
import { DecryptText } from '../ui/DecryptText';
import { XtermTerminal } from '../ui/XtermTerminal';
import type { IAgentApiClient } from '../../api';

export interface TerminalPaneProps {
  selectedSessionId: string | null;
  nodeName: string;
  apiClient: IAgentApiClient;
  terminalBackground?: ReactNode;
  lastSaveTime: string | null;
  saving: boolean;
  saveCooldown: boolean;
  actionError: string | null;
  onSave: () => void;
  onOpenConfig: (open: boolean) => void;
  loadingConfig: boolean;
  headerActions?: ReactNode;
}

export const TerminalPane = memo(function TerminalPane({
  selectedSessionId,
  nodeName,
  apiClient,
  terminalBackground,
  lastSaveTime,
  saving,
  saveCooldown,
  actionError,
  onSave,
  onOpenConfig,
  loadingConfig,
  headerActions,
}: TerminalPaneProps) {
  return (
    <div className="flex flex-col min-w-0 bg-background relative p-4 z-10 overflow-hidden flex-1">
      <Panel className="flex-1 flex flex-col h-full overflow-hidden" cornerSize={24}>
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(0,255,255,0.03)_0,transparent_100%)] pointer-events-none"></div>
        {selectedSessionId ? (
          <div
            key={selectedSessionId}
            className="flex flex-col h-full overflow-hidden relative terminal-pane-animate"
          >
            <div className="h-14 border-b border-primary/20 flex items-center justify-between px-6 shrink-0 bg-black/40 backdrop-blur-md z-10 relative">
              <div className="absolute top-0 left-0 w-full h-[1px] bg-gradient-to-r from-transparent via-primary/50 to-transparent"></div>
              <div className="flex items-center gap-3">
                <TerminalIcon className="w-4 h-4 text-primary" />
                <span className="font-mono text-sm tracking-widest uppercase text-primary font-bold">
                  <DecryptText text={`${nodeName} // ${selectedSessionId}`} />
                </span>
              </div>
              <div className="flex items-center gap-2">
                {headerActions}
                {actionError && (
                  <span className="text-[10px] text-destructive flex items-center gap-1 font-mono uppercase tracking-widest bg-destructive/10 px-2 py-1 border border-destructive/30">
                    <AlertCircle className="w-3 h-3" />
                    {actionError}
                  </span>
                )}
                {lastSaveTime && !actionError && (
                  <span className="text-[10px] text-primary/50 font-mono uppercase tracking-widest px-2">
                    SYNCED: {new Date(lastSaveTime).toLocaleString()}
                  </span>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-primary hover:text-primary hover:bg-primary/20 rounded-none border border-transparent hover:border-primary/30 uppercase tracking-widest text-[10px] font-mono transition-all"
                  onClick={onSave}
                  disabled={saving || saveCooldown}
                  title="保存所有 Loops"
                >
                  <Save className="w-3.5 h-3.5 mr-1" />
                  <DecryptText
                    text={saving ? 'SYNCING...' : saveCooldown ? 'SYNCED ✓' : 'SYNC STATE'}
                  />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  className="text-primary hover:text-primary hover:bg-primary/20 rounded-none border border-transparent hover:border-primary/30 uppercase tracking-widest text-[10px] font-mono transition-all"
                  onClick={() => {
                    onOpenConfig(true);
                  }}
                  disabled={loadingConfig}
                >
                  <Info className="w-3.5 h-3.5 mr-1" />
                  <DecryptText text={loadingConfig ? 'LOADING...' : 'VIEW CONFIG'} />
                </Button>
              </div>
            </div>

            <div className="flex-1 p-4 overflow-hidden relative z-0">
              <div className="absolute inset-0 bg-[linear-gradient(rgba(0,255,255,0.02)_1px,transparent_1px),linear-gradient(90deg,rgba(0,255,255,0.02)_1px,transparent_1px)] bg-[size:40px_40px] pointer-events-none"></div>
              <Panel className="w-full h-full" cornerSize={16} transparent={!!terminalBackground}>
                <div
                  className={`w-full h-full overflow-hidden relative z-10 ${terminalBackground ? 'bg-transparent' : 'bg-[hsl(var(--terminal-background))]'} text-[hsl(var(--terminal-foreground))]`}
                >
                  <XtermTerminal
                    wsUrl={apiClient.buildWsUrl(selectedSessionId)}
                    backgroundComponent={terminalBackground}
                  />
                </div>
              </Panel>
            </div>
          </div>
        ) : (
          <div className="flex-1 flex items-center justify-center text-primary/40 uppercase tracking-widest text-xs">
            <div className="text-center space-y-4">
              <TerminalSquare className="w-16 h-16 mx-auto opacity-20" />
              <DecryptText text="AWAITING LOOP SELECTION" />
            </div>
          </div>
        )}
      </Panel>
    </div>
  );
});
