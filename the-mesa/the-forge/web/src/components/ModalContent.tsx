import { useState } from 'react';
import { createPortal } from 'react-dom';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@maze/fabrication';
import { Button } from '@maze/fabrication';

export interface ModalConfig {
  mode: 'confirm' | 'prompt';
  title: string;
  message: string;
  defaultValue?: string;
  placeholder?: string;
  variant?: 'default' | 'danger';
}

function ModalInner({
  config,
  onClose,
}: {
  config: ModalConfig;
  onClose: (value: boolean | string | null) => void;
}) {
  const [inputValue, setInputValue] = useState(
    config.mode === 'prompt' ? (config.defaultValue ?? '') : '',
  );

  const handleSubmit = () => {
    onClose(config.mode === 'prompt' ? inputValue : true);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') handleSubmit();
  };

  const cancelValue = config.mode === 'confirm' ? false : null;

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose(cancelValue);
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle
            className={config.variant === 'danger' ? 'text-destructive' : 'text-primary'}
          >
            {config.title}
          </DialogTitle>
        </DialogHeader>
        {config.mode === 'confirm' ? (
          <p className="text-foreground font-mono text-sm">{config.message}</p>
        ) : (
          <div>
            <p className="text-foreground font-mono text-sm mb-3">{config.message}</p>
            <input
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder={config.placeholder || ''}
              className="w-full px-3 py-2 bg-background border border-border rounded text-foreground font-mono text-sm focus:outline-none focus:border-primary transition"
              autoFocus
            />
          </div>
        )}
        <DialogFooter>
          <Button variant="ghost" onClick={() => onClose(cancelValue)}>
            CANCEL
          </Button>
          <Button
            variant={config.variant === 'danger' ? 'destructive' : 'default'}
            onClick={handleSubmit}
          >
            {config.mode === 'confirm' ? 'CONFIRM' : 'CREATE'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export default function ModalPortal({
  config,
  onClose,
}: {
  config: ModalConfig | null;
  onClose: (value: boolean | string | null) => void;
}) {
  if (!config) return null;
  return createPortal(<ModalInner config={config} onClose={onClose} />, document.body);
}
