import { useState, useCallback, createContext, useContext, type ReactNode } from 'react';
import { Panel } from './Panel';

type ToastType = 'success' | 'error' | 'warning';

interface ToastItem {
  id: number;
  type: ToastType;
  message: string;
  exiting?: boolean;
}

interface ToastContextValue {
  showToast: (type: ToastType, message: string) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast must be used within ToastProvider');
  return ctx;
}

const variantMap: Record<ToastType, 'default' | 'destructive' | 'warning' | 'success'> = {
  success: 'success',
  error: 'destructive',
  warning: 'warning',
};

const iconMap: Record<ToastType, string> = {
  success: '[ OK ]',
  error: '[ ERR ]',
  warning: '[ WARN ]',
};

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);

  const showToast = useCallback((type: ToastType, message: string) => {
    // 用时间戳 + 随机数生成唯一 ID，避免极端并发下的碰撞
    const id = Date.now() + Math.random();
    setToasts((prev) => [...prev, { id, type, message }]);
    setTimeout(() => {
      setToasts((prev) => prev.map((t) => (t.id === id ? { ...t, exiting: true } : t)));
    }, 3700);
    setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 4000);
  }, []);

  const removeToast = (id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  return (
    <ToastContext.Provider value={{ showToast }}>
      {children}
      {/* Toast 容器：固定在右下角，不影响布局流 */}
      {/* aria-live="polite" 使屏幕阅读器在空闲时自动播报新 Toast，避免打断用户当前操作 */}
      <div
        role="status"
        aria-live="polite"
        className="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 pointer-events-none"
      >
        {toasts.map((toast) => (
          <div
            key={toast.id}
            className={`pointer-events-auto ${toast.exiting ? 'animate-out fade-out-0 slide-out-to-right-5 duration-300' : 'animate-in slide-in-from-right-5 fade-in-0 duration-300'}`}
            onClick={() => {
              removeToast(toast.id);
            }}
          >
            <Panel variant={variantMap[toast.type]} cornerSize={8} className="min-w-[280px]">
              <div className="flex items-center gap-2 text-xs font-mono">
                <span className="shrink-0">{iconMap[toast.type]}</span>
                <span className="truncate">{toast.message}</span>
              </div>
            </Panel>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}
