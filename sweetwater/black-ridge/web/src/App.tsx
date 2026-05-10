import { useState } from 'react';
import {
  AgentPanel,
  BootSequence,
  TerrainBackground,
  Button,
  AnimationSettingsPanel,
  AppShell,
  AppNavbar,
  CalibrationMarks,
} from '@maze/fabrication';
import { Activity, Settings } from 'lucide-react';
import { api } from './api/client';
import './index.css';

function AppContent() {
  const [showAnimSettings, setShowAnimSettings] = useState(false);

  return (
    <>
      <div className="h-screen w-screen bg-background text-foreground dark relative overflow-hidden grid grid-rows-[56px_1fr] grid-cols-1">
        <CalibrationMarks />

        <AppNavbar
          title="SWEETWATER // FIELD DIAGNOSTIC UNIT"
          icon={<Activity className="w-5 h-5 text-primary" />}
          rightContent={
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
          }
        />

        <div className="flex w-full h-full relative z-10 bg-background overflow-hidden">
          <AgentPanel
            apiClient={api}
            nodeName="SWEETWATER"
            terminalBackground={<TerrainBackground />}
          />
        </div>
      </div>
      <AnimationSettingsPanel open={showAnimSettings} onOpenChange={setShowAnimSettings} />
    </>
  );
}

function App() {
  const [isBooting, setIsBooting] = useState(true);

  if (isBooting) {
    return (
      <AppShell>
        <BootSequence
          onComplete={() => {
            setIsBooting(false);
          }}
          division="sweetwater"
        />
      </AppShell>
    );
  }

  return (
    <AppShell>
      <AppContent />
    </AppShell>
  );
}

export default App;
