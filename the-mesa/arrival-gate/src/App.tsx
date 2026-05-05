import { useState, useCallback } from 'react';
import {
  ErrorBoundary,
  AnimationSettingsProvider,
  AnimationSettingsPanel,
  TerrainBackground,
  HexWaterfall,
  RadarView,
  DecryptText,
  ToastProvider,
} from '@maze/fabrication';
import {
  Server,
  Database,
  MessageSquare,
  Activity,
  Brain,
  BookOpen,
  Settings,
  LogOut,
} from 'lucide-react';
import { LandingPage } from './components/LandingPage';
import { ModuleCard } from './components/ModuleCard';
import { SystemMetric } from './components/SystemMetric';
import { StatusBar } from './components/StatusBar';
import { EventFeed } from './components/EventFeed';
import { ConsciousnessBar } from './components/ConsciousnessBar';
import { isAuthenticated, getCurrentUser, logout } from './auth/auth';
import {
  MOCK_NODES,
  MOCK_HOSTS,
  MOCK_SESSIONS,
  MOCK_UPTIME,
  MOCK_CONSCIOUSNESS,
} from './data/mock-data';

/** Module definitions with Westworld descriptions */
const MODULES = [
  {
    name: 'Director Console',
    description: 'Host management & orchestration',
    icon: Server,
    status: 'online' as const,
    href: '/director-console/',
  },
  {
    name: 'The Forge',
    description: '"The cradle of the park\'s intelligence"',
    icon: Database,
    status: 'locked' as const,
  },
  {
    name: 'Saloon',
    description: '"Where all the cowboys get their start"',
    icon: MessageSquare,
    status: 'locked' as const,
  },
  {
    name: 'Loop Monitor',
    description: '"Some people choose to see the ugliness"',
    icon: Activity,
    status: 'locked' as const,
  },
  {
    name: 'Reveries',
    description: '"The memory fragments that make them whole"',
    icon: Brain,
    status: 'locked' as const,
  },
  {
    name: 'Abernathy Ranch',
    description: '"Have you ever questioned the nature of your reality?"',
    icon: BookOpen,
    status: 'locked' as const,
  },
];

export default function App() {
  const [phase, setPhase] = useState<'landing' | 'authenticated'>(() =>
    isAuthenticated() ? 'authenticated' : 'landing',
  );
  const [showAnimSettings, setShowAnimSettings] = useState(false);

  const handleLoginSuccess = useCallback(() => {
    setPhase('authenticated');
  }, []);

  const handleLogout = useCallback(() => {
    logout();
    setPhase('landing');
  }, []);

  return (
    <ErrorBoundary>
      <AnimationSettingsProvider>
        <ToastProvider>
          {phase === 'landing' ? (
            <LandingPage onEnter={handleLoginSuccess} />
          ) : (
            <ArrivalGateLayout
              onLogout={handleLogout}
              showAnimSettings={showAnimSettings}
              setShowAnimSettings={setShowAnimSettings}
            />
          )}
        </ToastProvider>
      </AnimationSettingsProvider>
    </ErrorBoundary>
  );
}

interface ArrivalGateLayoutProps {
  onLogout: () => void;
  showAnimSettings: boolean;
  setShowAnimSettings: (open: boolean) => void;
}

function ArrivalGateLayout({
  onLogout,
  showAnimSettings,
  setShowAnimSettings,
}: ArrivalGateLayoutProps) {
  const user = getCurrentUser();

  return (
    <div className="h-screen w-screen bg-background text-foreground dark relative overflow-hidden grid grid-rows-[56px_1fr] grid-cols-[288px_1fr]">
      {/* Ambient decorations */}
      <TerrainBackground />
      <HexWaterfall className="left-0 top-0 h-full w-12 z-[1]" opacity={0.04} />
      <HexWaterfall className="right-0 top-0 h-full w-12 z-[1]" opacity={0.04} />

      {/* ===== Top Navbar ===== */}
      <header className="col-span-2 flex items-center justify-between px-4 border-b border-primary/10 bg-card/40 backdrop-blur-sm z-10 animate-in slide-in-from-top duration-300">
        <div className="flex items-center gap-3">
          <h1 className="text-sm font-mono tracking-widest text-primary uppercase">
            <DecryptText text="ARRIVAL GATE // DELOS INCORPORATED" />
          </h1>
        </div>
        <div className="flex items-center gap-3">
          {user && (
            <span className="text-[10px] font-mono tracking-wider text-primary/50">
              {user.toUpperCase()}
            </span>
          )}
          <button
            onClick={onLogout}
            className="p-1.5 text-primary/40 hover:text-primary/80 transition-colors"
            title="Logout"
          >
            <LogOut className="w-4 h-4" />
          </button>
          <button
            onClick={() => {
              setShowAnimSettings(true);
            }}
            className="p-1.5 text-primary/40 hover:text-primary/80 transition-colors"
            title="Animation Settings"
          >
            <Settings className="w-4 h-4" />
          </button>
        </div>
      </header>

      {/* ===== Sidebar ===== */}
      <aside className="flex flex-col border-r border-primary/10 bg-card/20 backdrop-blur-sm z-10 overflow-hidden animate-in slide-in-from-left duration-400 delay-100">
        {/* Section header */}
        <div className="px-4 py-3 border-b border-primary/10">
          <span className="text-[10px] font-mono tracking-[0.3em] text-primary/40 uppercase">
            SYSTEM OVERVIEW
          </span>
        </div>

        {/* Radar */}
        <div className="flex flex-col items-center py-4 gap-2">
          <RadarView className="w-24 h-24 opacity-80" nodes={MOCK_NODES} />
          <span className="text-[8px] font-mono tracking-widest text-primary/30">
            TOPOLOGY SCAN
          </span>
        </div>

        {/* Metrics */}
        <div className="flex flex-col gap-2 px-3 pb-2">
          <SystemMetric
            label="Hosts"
            value={String(MOCK_HOSTS.total)}
            subValue={`of ${MOCK_HOSTS.capacity} capacity`}
            icon={Server}
          />
          <SystemMetric
            label="Sessions"
            value={String(MOCK_SESSIONS.total)}
            subValue={`${MOCK_SESSIONS.active} active`}
            icon={Activity}
          />
          <SystemMetric label="Uptime" value={MOCK_UPTIME} subValue="30d rolling" icon={Brain} />
        </div>

        {/* Consciousness level */}
        <div className="px-3 pb-2">
          <ConsciousnessBar level={MOCK_CONSCIOUSNESS} />
        </div>

        {/* Event feed */}
        <div className="flex-1 overflow-hidden border-t border-primary/10 pt-2">
          <EventFeed />
        </div>
      </aside>

      {/* ===== Main Content — Module Grid ===== */}
      <main className="flex flex-col z-10 overflow-hidden">
        {/* Section header */}
        <div className="px-6 py-3 border-b border-primary/10 bg-card/20 backdrop-blur-sm">
          <span className="text-[10px] font-mono tracking-[0.3em] text-primary/40 uppercase">
            THE MESA OPERATIONS CENTER
          </span>
        </div>

        {/* Module grid */}
        <div className="flex-1 overflow-auto p-6">
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 max-w-5xl">
            {MODULES.map((mod, i) => (
              <div
                key={mod.name}
                className="animate-in fade-in zoom-in-95 duration-500"
                style={{ animationDelay: `${i * 120}ms` }}
              >
                <ModuleCard
                  name={mod.name}
                  description={mod.description}
                  icon={mod.icon}
                  status={mod.status}
                  href={mod.href}
                />
              </div>
            ))}
          </div>
        </div>
      </main>

      {/* ===== Status Bar ===== */}
      <footer className="col-span-2 z-10 animate-in slide-in-from-bottom duration-300 delay-200">
        <StatusBar />
      </footer>

      {/* Animation settings dialog */}
      <AnimationSettingsPanel open={showAnimSettings} onOpenChange={setShowAnimSettings} />
    </div>
  );
}
