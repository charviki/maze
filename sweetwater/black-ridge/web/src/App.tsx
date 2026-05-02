import { useState } from 'react';
import {
  AgentPanel,
  DecryptText,
  BootSequence,
  TerrainBackground,
  ErrorBoundary,
  Button,
  AnimationSettingsProvider,
  AnimationSettingsPanel,
  ToastProvider,
} from '@maze/fabrication';
import { Activity, Settings } from 'lucide-react';
import { api } from './api/client';
import './index.css';

export default function App() {
  const [isBooting, setIsBooting] = useState(true);
  const [showAnimSettings, setShowAnimSettings] = useState(false);

  if (isBooting) {
    return (
      <ErrorBoundary>
        <AnimationSettingsProvider>
          <BootSequence
            onComplete={() => {
              setIsBooting(false);
            }}
            division="sweetwater"
          />
        </AnimationSettingsProvider>
      </ErrorBoundary>
    );
  }

  return (
    <ErrorBoundary>
      <AnimationSettingsProvider>
        <ToastProvider>
          <div className="h-screen w-screen bg-background text-foreground dark relative overflow-hidden grid grid-rows-[56px_1fr] grid-cols-1">
            <div className="pointer-events-none absolute inset-0">
              <div className="absolute top-4 left-4 w-4 h-4 border-t border-l border-primary/30"></div>
              <div className="absolute top-4 right-4 w-4 h-4 border-t border-r border-primary/30"></div>
              <div className="absolute bottom-4 left-4 w-4 h-4 border-b border-l border-primary/30"></div>
              <div className="absolute bottom-4 right-4 w-4 h-4 border-b border-r border-primary/30"></div>

              <div className="absolute top-1/2 left-4 w-2 h-[1px] bg-primary/30"></div>
              <div className="absolute top-1/2 right-4 w-2 h-[1px] bg-primary/30"></div>
              <div className="absolute top-4 left-1/2 w-[1px] h-2 bg-primary/30"></div>
              <div className="absolute bottom-4 left-1/2 w-[1px] h-2 bg-primary/30"></div>
            </div>

            <div className="border-b border-border/50 flex items-center justify-between px-6 bg-background relative overflow-hidden z-10">
              <div className="absolute top-0 left-0 w-full h-[1px] bg-primary/20"></div>
              <div className="flex items-center gap-3 font-bold text-lg tracking-wider text-primary">
                <Activity className="w-5 h-5 text-primary" />
                <DecryptText text="SWEETWATER // FIELD DIAGNOSTIC UNIT" className="uppercase" />
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="text-primary hover:text-primary hover:bg-primary/20 rounded-none"
                onClick={() => {
                  setShowAnimSettings(true);
                }}
                title="Visual Effects"
              >
                <Settings className="w-4 h-4" />
              </Button>
            </div>

            <div className="flex w-full h-full relative z-10 bg-background overflow-hidden">
              <AgentPanel
                apiClient={api}
                nodeName="SWEETWATER"
                terminalBackground={<TerrainBackground />}
              />
            </div>
          </div>
        </ToastProvider>
        <AnimationSettingsPanel open={showAnimSettings} onOpenChange={setShowAnimSettings} />
      </AnimationSettingsProvider>
    </ErrorBoundary>
  );
}
