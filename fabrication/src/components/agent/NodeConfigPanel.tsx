import { useState, useEffect, useCallback } from 'react';
import { Variable, Plus, Trash2, Save } from 'lucide-react';
import type { IAgentApiClient } from '../../api';
import type { LocalAgentConfig } from '../../types';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '../ui/dialog';

interface NodeConfigPanelProps {
  nodeName: string;
  apiClient: IAgentApiClient;
  onClose: () => void;
}

export function NodeConfigPanel({ nodeName, apiClient, onClose }: NodeConfigPanelProps) {
  const [config, setConfig] = useState<LocalAgentConfig>({ working_dir: '', env: {} });
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    setLoading(true);
    const res = await apiClient.getLocalConfig();
    if (res.status === 'ok' && res.data) {
      setConfig(res.data);
    }
    setLoading(false);
  }, [apiClient]);

  useEffect(() => { load(); }, [load]);

  const handleSave = async () => {
    await apiClient.updateLocalConfig({ env: config.env });
    onClose();
  };

  const updateEnv = (oldKey: string, newKey: string, value: string) => {
    const newEnv = { ...config.env };
    if (oldKey !== newKey) {
      delete newEnv[oldKey];
    }
    newEnv[newKey] = value;
    setConfig({ ...config, env: newEnv });
  };

  const removeEnv = (key: string) => {
    const newEnv = { ...config.env };
    delete newEnv[key];
    setConfig({ ...config, env: newEnv });
  };

  const addEnv = () => {
    setConfig({ ...config, env: { ...config.env, '': '' } });
  };

  if (loading) return null;

  return (
    <Dialog open onOpenChange={(v) => { if (!v) onClose(); }}>
      <DialogContent className="max-w-4xl max-h-[85vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Variable className="w-5 h-5" />
            Host 配置: {nodeName}
          </DialogTitle>
          <DialogDescription>
            管理 Host 默认环境变量；基础工作目录由服务端配置决定，只读展示。
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          <div>
            <label className="text-xs text-muted-foreground mb-1 block">基础工作目录</label>
            <Input 
              value={config.working_dir} 
              readOnly
              className="font-mono text-sm"
            />
          </div>

          <div>
            <div className="flex items-center justify-between mb-2">
              <h3 className="font-medium text-sm text-muted-foreground flex items-center gap-1">
                <Variable className="w-4 h-4" /> 环境变量
              </h3>
              <Button size="sm" variant="outline" onClick={addEnv}><Plus className="w-3 h-3 mr-1" /> 新增</Button>
            </div>
            
            <div className="space-y-2">
              {Object.entries(config.env).map(([key, value]) => (
                <div key={key} className="flex items-center gap-2">
                  <Input 
                    value={key} 
                    onChange={e => updateEnv(key, e.target.value, String(value))} 
                    placeholder="KEY" 
                    className="w-1/3 font-mono text-sm"
                  />
                  <span className="text-muted-foreground">=</span>
                  <Input 
                    value={String(value)} 
                    onChange={e => updateEnv(key, key, e.target.value)} 
                    placeholder="VALUE" 
                    className="flex-1 font-mono text-sm"
                  />
                  <Button variant="ghost" size="sm" onClick={() => removeEnv(key)} className="text-red-500">
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              ))}
              {Object.keys(config.env).length === 0 && (
                <div className="text-sm text-muted-foreground p-4 text-center border border-dashed rounded">暂无环境变量配置</div>
              )}
            </div>
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-4 border-t border-border">
          <Button variant="outline" onClick={onClose}>取消</Button>
          <Button onClick={handleSave}><Save className="w-4 h-4 mr-1" /> 保存</Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
