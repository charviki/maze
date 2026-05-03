import { memo, useEffect, useState, type ReactNode } from 'react';
import type { ConfigFileSnapshot, PipelineStep, SessionConfigView } from '../../types';
import type { IAgentApiClient } from '../../api';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '../ui/dialog';
import { SessionPipeline } from '../ui/SessionPipeline';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { CreateSessionWithTemplateDialog } from './CreateSessionWithTemplateDialog';
import { TemplateManager } from './TemplateManager';
import { NodeConfigPanel } from './NodeConfigPanel';
import type { SessionDisplay } from './SessionList';
import { useToast } from '../ui/Toast';

export interface SessionDialogsProps {
  killTarget: SessionDisplay | null;
  killing: boolean;
  restoreTarget: SessionDisplay | null;
  viewPipelineSession: SessionDisplay | null;
  viewPipelineSteps: PipelineStep[];
  configOpen: boolean;
  selectedSessionId: string | null;
  nodeName: string;
  restoring: boolean;
  showCreate: boolean;
  onShowCreateChange: (v: boolean) => void;
  showTemplateManager: boolean;
  onShowTemplateManagerChange: (v: boolean) => void;
  showNodeConfig: boolean;
  onShowNodeConfigClose: () => void;
  apiClient: IAgentApiClient;
  renderCreateDialog?: (props: {
    open: boolean;
    onOpenChange: (open: boolean) => void;
    onSuccess: (sessionName: string) => void;
  }) => ReactNode;
  onSessionCreated: (sessionName: string) => void;
  onKillConfirm: () => void;
  onKillCancel: () => void;
  onRestoreConfirm: () => void;
  onRestoreCancel: () => void;
  onPipelineClose: () => void;
  onConfigClose: (open: boolean) => void;
  getSnapshotSummary: (snapshot?: string, lines?: number) => string;
}

export const SessionDialogs = memo(function SessionDialogs({
  killTarget,
  killing,
  restoreTarget,
  viewPipelineSession,
  viewPipelineSteps,
  configOpen,
  selectedSessionId,
  nodeName,
  restoring,
  showCreate,
  onShowCreateChange,
  showTemplateManager,
  onShowTemplateManagerChange,
  showNodeConfig,
  onShowNodeConfigClose,
  apiClient,
  renderCreateDialog,
  onSessionCreated,
  onKillConfirm,
  onKillCancel,
  onRestoreConfirm,
  onRestoreCancel,
  onPipelineClose,
  onConfigClose,
  getSnapshotSummary,
}: SessionDialogsProps) {
  const { showToast } = useToast();
  return (
    <>
      {renderCreateDialog ? (
        renderCreateDialog({
          open: showCreate,
          onOpenChange: onShowCreateChange,
          onSuccess: onSessionCreated,
        })
      ) : (
        <CreateSessionWithTemplateDialog
          open={showCreate}
          onOpenChange={onShowCreateChange}
          apiClient={apiClient}
          nodeName={nodeName}
          onSuccess={onSessionCreated}
          onOpenTemplateManager={() => {
            onShowTemplateManagerChange(true);
          }}
        />
      )}

      {showTemplateManager && (
        <TemplateManager
          open={showTemplateManager}
          onClose={() => {
            onShowTemplateManagerChange(false);
          }}
          apiClient={apiClient}
        />
      )}

      {showNodeConfig && (
        <NodeConfigPanel
          nodeName={nodeName}
          apiClient={apiClient}
          onClose={onShowNodeConfigClose}
        />
      )}

      {killTarget && (
        <Dialog open={!!killTarget} onOpenChange={(open) => !open && !killing && onKillCancel()}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>确认终止 Loop</DialogTitle>
              <DialogDescription>
                确认终止 Loop「{killTarget.name}」并清理所有数据？此操作不可恢复。
              </DialogDescription>
            </DialogHeader>
            <DialogFooter>
              <Button variant="ghost" onClick={onKillCancel} disabled={killing}>
                取消
              </Button>
              {/* 删除请求期间显式展示进行中状态，避免用户误以为点击无效。 */}
              <Button variant="destructive" onClick={onKillConfirm} disabled={killing}>
                {killing ? '终止中...' : '确认终止'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {restoreTarget && (
        <Dialog open={!!restoreTarget} onOpenChange={(open) => !open && onRestoreCancel()}>
          <DialogContent className="max-w-4xl max-h-[85vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>确认恢复 Loop</DialogTitle>
              <DialogDescription>
                确认恢复 Loop「{restoreTarget.name}」？将重新执行创建时的命令管线。
              </DialogDescription>
            </DialogHeader>
            {restoreTarget.terminalSnapshot && (
              <div className="mt-2 w-full overflow-hidden">
                <span className="text-xs text-muted-foreground">终端快照摘要：</span>
                <pre className="mt-1 p-2 bg-muted/50 rounded text-xs font-mono overflow-y-auto overflow-x-auto whitespace-pre-wrap break-all max-h-64 w-full">
                  {getSnapshotSummary(restoreTarget.terminalSnapshot)}
                </pre>
              </div>
            )}
            <DialogFooter>
              <Button variant="ghost" onClick={onRestoreCancel}>
                取消
              </Button>
              <Button onClick={onRestoreConfirm} disabled={restoring}>
                {restoring ? '恢复中...' : '确认恢复'}
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      )}

      {viewPipelineSession && (
        <Dialog open={!!viewPipelineSession} onOpenChange={(open) => !open && onPipelineClose()}>
          <DialogContent className="max-w-[80vw] max-h-[85vh] overflow-y-auto">
            <DialogHeader>
              <DialogTitle>命令管线 — {viewPipelineSession.name}</DialogTitle>
              <DialogDescription>该 Loop 创建时执行的命令管线步骤</DialogDescription>
            </DialogHeader>
            <SessionPipeline steps={viewPipelineSteps} onChange={() => {}} readOnly />
          </DialogContent>
        </Dialog>
      )}

      <SessionConfigEditorDialog
        open={configOpen}
        onOpenChange={onConfigClose}
        selectedSessionId={selectedSessionId}
        nodeName={nodeName}
        apiClient={apiClient}
        showToast={showToast}
      />
    </>
  );
});

interface SessionConfigEditorDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  selectedSessionId: string | null;
  nodeName: string;
  apiClient: IAgentApiClient;
  showToast: (type: 'success' | 'error' | 'warning', message: string) => void;
}

function SessionConfigEditorDialog({
  open,
  onOpenChange,
  selectedSessionId,
  nodeName,
  apiClient,
  showToast,
}: SessionConfigEditorDialogProps) {
  const [config, setConfig] = useState<SessionConfigView | null>(null);
  const [files, setFiles] = useState<ConfigFileSnapshot[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (!open || !selectedSessionId) {
      return;
    }

    let cancelled = false;
    const load = async () => {
      setLoading(true);
      setError('');
      const res = await apiClient.getSessionConfig(selectedSessionId);
      if (cancelled) return;
      setLoading(false);
      if (res.status !== 'ok' || !res.data) {
        const message = res.message || '加载 Loop 配置失败';
        setError(message);
        showToast('error', message);
        return;
      }
      setConfig(res.data);
      setFiles(
        (res.data.files ?? []).map((f) => ({
          ...f,
          exists: f._exists ?? false,
        })),
      );
    };

    void load();
    return () => {
      cancelled = true;
    };
  }, [open, selectedSessionId, apiClient, showToast]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Narrative Loop 配置</DialogTitle>
          <DialogDescription>
            {nodeName} / {selectedSessionId || '-'} — 仅供查看，Session 运行中修改不会自动生效
          </DialogDescription>
        </DialogHeader>

        {loading ? (
          <div className="text-sm text-muted-foreground">正在读取工作目录中的真实配置文件...</div>
        ) : (
          <div className="space-y-4">
            {error && (
              <div className="text-xs text-destructive bg-destructive/10 border border-destructive/20 px-3 py-2 rounded">
                {error}
              </div>
            )}

            {config && (
              <div className="grid grid-cols-2 gap-x-6 gap-y-2 text-sm">
                <div className="min-w-0">
                  <span className="text-muted-foreground">工作目录: </span>
                  <span className="font-mono text-xs break-all">{config.workingDir}</span>
                </div>
                <div className="min-w-0">
                  <span className="text-muted-foreground">模板: </span>
                  <span className="font-mono text-xs break-all">{config.templateId}</span>
                </div>
              </div>
            )}

            {files.map((file) => (
              <div key={file.path} className="border border-border rounded p-3 space-y-2">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex-1">
                    <label className="text-xs text-muted-foreground">固定路径</label>
                    <Input value={file.path} readOnly className="font-mono text-sm" />
                  </div>
                  <div className="mt-5 text-[11px] text-muted-foreground shrink-0">
                    {file.exists ? '文件已存在' : '文件不存在，当前按空内容处理'}
                  </div>
                </div>
                <div>
                  <label className="text-xs text-muted-foreground">真实内容</label>
                  <textarea
                    value={file.content}
                    readOnly
                    className="w-full h-32 bg-muted/50 text-foreground text-xs font-mono p-2 rounded border border-border cursor-default select-none"
                  />
                </div>
              </div>
            ))}
          </div>
        )}

        <DialogFooter>
          <Button
            variant="ghost"
            onClick={() => {
              onOpenChange(false);
            }}
          >
            关闭
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
