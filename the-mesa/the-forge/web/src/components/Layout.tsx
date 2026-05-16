import { useState, type ReactNode, useMemo } from 'react';
import { useLocation } from 'react-router-dom';
import Sidebar from './Sidebar';
import DocTree from './DocTree';
import PageTransition from './PageTransition';
import CommandPalette from './CommandPalette';
import { useKeyboardShortcuts } from '@/hooks/useKeyboardShortcuts';
import { SidebarContext } from '@/hooks/useSidebar';
import { DocTreeProvider, useDocTreeRefresh } from '@/contexts/DocTreeContext';

function LayoutInner({ children }: { children: ReactNode }) {
  const [collapsed, setCollapsed] = useState(false);
  const toggleCollapsed = () => setCollapsed((c) => !c);
  const location = useLocation();
  const [showPalette, setShowPalette] = useState(false);
  const [treeWidth, setTreeWidth] = useState(280);
  const { refreshKey } = useDocTreeRefresh();

  useKeyboardShortcuts({
    onToggleCommandPalette: () => setShowPalette((prev) => !prev),
    onEscape: () => {
      if (showPalette) setShowPalette(false);
    },
  });

  // Extract archiveId from URL path pattern /docs/:archiveId/...
  const archiveId = useMemo(() => {
    const match = location.pathname.match(/^\/docs\/([^/]+)/);
    return match ? match[1] : undefined;
  }, [location.pathname]);

  const sidebarWidth = collapsed ? 56 : 240;
  const isDocsRoute = archiveId !== undefined;

  return (
    <SidebarContext.Provider value={{ collapsed, toggleCollapsed }}>
      <div className="min-h-screen bg-background text-foreground dark flex">
        <Sidebar />
        {isDocsRoute && (
          <div style={{ marginLeft: sidebarWidth }}>
            <DocTree
              key={refreshKey}
              archiveId={archiveId}
              width={treeWidth}
              onResize={(delta) => setTreeWidth((w) => Math.max(200, Math.min(500, w + delta)))}
            />
          </div>
        )}
        <main
          className="flex-1 p-6 overflow-auto transition-all duration-200"
          style={{
            marginLeft: isDocsRoute ? 0 : `${sidebarWidth}px`,
          }}
        >
          <PageTransition key={location.pathname}>{children}</PageTransition>
        </main>
      </div>
      {showPalette && <CommandPalette open={showPalette} onClose={() => setShowPalette(false)} />}
    </SidebarContext.Provider>
  );
}

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <DocTreeProvider>
      <LayoutInner>{children}</LayoutInner>
    </DocTreeProvider>
  );
}
