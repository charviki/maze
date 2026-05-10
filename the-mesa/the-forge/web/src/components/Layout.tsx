import { useState, type ReactNode } from 'react';
import { useLocation } from 'react-router-dom';
import Sidebar from './Sidebar';
import PageTransition from './PageTransition';
import CommandPalette from './CommandPalette';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';
import { SidebarContext } from '@/hooks/useSidebar';

export default function Layout({ children }: { children: ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);
  const toggleCollapsed = () => setCollapsed((c) => !c);
  const location = useLocation();
  const [showPalette, setShowPalette] = useState(false);
  useKeyboardShortcuts({
    onToggleCommandPalette: () => setShowPalette((prev) => !prev),
    onEscape: () => {
      if (showPalette) setShowPalette(false);
    },
  });

  return (
    <SidebarContext.Provider value={{ collapsed, toggleCollapsed }}>
      <div className="min-h-screen bg-background text-foreground dark flex">
        <Sidebar />
        <main
          className={`flex-1 ${collapsed ? 'ml-14' : 'ml-56'} p-6 overflow-auto transition-all duration-200`}
        >
          <PageTransition key={location.pathname}>{children}</PageTransition>
        </main>
      </div>
      {showPalette && <CommandPalette open={showPalette} onClose={() => setShowPalette(false)} />}
    </SidebarContext.Provider>
  );
}
