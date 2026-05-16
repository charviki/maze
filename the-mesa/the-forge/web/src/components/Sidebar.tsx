import { useState, useEffect, useCallback } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Home, Database, Plus, Trash2, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { archives, type Archive } from '@/api';
import { useToast } from '@maze/fabrication';
import { useModal } from '@/hooks/useModal';
import ModalPortal from '@/components/ModalContent';
import { useSidebar } from '@/hooks/useSidebar';
import { useIdentity } from '@/hooks/useIdentity';
import HexLogo from './HexLogo';

function Tooltip({
  children,
  text,
  show,
}: {
  children: React.ReactNode;
  text: string;
  show: boolean;
}) {
  if (!show) return <>{children}</>;
  return (
    <div className="relative group/tooltip">
      {children}
      <div className="absolute left-full top-1/2 -translate-y-1/2 ml-2 px-2 py-1 text-sm font-mono bg-secondary border border-border rounded text-foreground whitespace-nowrap opacity-0 group-hover/tooltip:opacity-100 transition-opacity pointer-events-none z-50">
        {text}
      </div>
    </div>
  );
}

export default function Sidebar() {
  const location = useLocation();
  const { showToast } = useToast();
  const { collapsed, toggleCollapsed } = useSidebar();
  const { username } = useIdentity();
  const [archiveList, setArchiveList] = useState<Archive[]>([]);
  const { config: modalConfig, confirm, prompt, close } = useModal();

  const isActive = (path: string) => {
    if (path === '/') return location.pathname === '/';
    return location.pathname.startsWith(path);
  };

  const loadArchives = useCallback(async () => {
    try {
      const res = await archives.list();
      setArchiveList(res.items || []);
    } catch {
      setArchiveList([]);
    }
  }, []);

  const handleCreateArchive = () => {
    void prompt('Archive name:', '', 'NEW ARCHIVE', 'Enter archive name')
      .then(async (name) => {
        if (!name?.trim()) return;
        await archives.create({ name: name.trim() });
        void loadArchives();
      })
      .catch(() => {
        showToast('error', 'Failed to create archive');
      });
  };

  const handleDeleteArchive = (archive: Archive) => {
    void confirm(
      `Delete archive "${archive.name}" and all its documents? This cannot be undone.`,
      'DELETE',
      'danger',
    )
      .then(async (ok) => {
        if (!ok) return;
        if (!archive.id) return;
        await archives.remove(archive.id);
        showToast('success', `Archive "${archive.name}" deleted`);
        void loadArchives();
      })
      .catch(() => {
        showToast('error', 'Failed to delete archive');
      });
  };

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    void loadArchives();
  }, [loadArchives]);

  const sidebarWidth = collapsed ? 'w-14' : 'w-60';

  return (
    <>
      <aside
        className={`fixed left-0 top-0 h-full ${sidebarWidth} bg-card border-r border-border flex flex-col transition-all duration-200 z-10`}
      >
        <div className="p-3 border-b border-border text-center">
          <Link to="/" className="block" title={collapsed ? 'THE FORGE' : undefined}>
            <HexLogo collapsed={collapsed} />
            {!collapsed && (
              <>
                <h1 className="text-primary font-mono font-bold text-base tracking-wider">
                  THE FORGE
                </h1>
                <p className="text-muted-foreground text-sm font-mono mt-0.5">DELOS // SECTOR 19</p>
              </>
            )}
          </Link>
        </div>

        <nav className="flex-1 py-3 overflow-y-auto overflow-x-hidden">
          <Tooltip text="DASHBOARD" show={collapsed}>
            <Link
              to="/"
              className={`flex items-center ${collapsed ? 'justify-center px-0' : 'gap-2 px-3'} py-2 text-sm font-mono transition-all duration-150 ${
                isActive('/') && !location.pathname.startsWith('/docs')
                  ? 'text-primary bg-primary/10 border-l-2 border-primary'
                  : 'text-muted-foreground hover:text-foreground hover:bg-border/30 border-l-2 border-transparent'
              }`}
            >
              <Home size={15} className="shrink-0" />
              {!collapsed && 'DASHBOARD'}
            </Link>
          </Tooltip>

          <div className="mt-2">
            {!collapsed && (
              <div className="flex items-center justify-between px-3 py-1">
                <p className="text-sm font-mono text-muted-foreground tracking-widest">ARCHIVES</p>
                <button
                  onClick={handleCreateArchive}
                  className="text-muted-foreground hover:text-primary transition"
                  title="Create archive"
                >
                  <Plus size={11} />
                </button>
              </div>
            )}
            {collapsed && (
              <div className="flex justify-center py-1">
                <button
                  onClick={handleCreateArchive}
                  className="text-muted-foreground hover:text-primary transition"
                  title="Create archive"
                >
                  <Plus size={11} />
                </button>
              </div>
            )}
            {archiveList.map((archive) => {
              const archiveActive = location.pathname.startsWith(`/docs/${archive.id}`);

              if (collapsed) {
                return (
                  <Link
                    key={archive.id}
                    to={`/docs/${archive.id}`}
                    title={archive.name}
                    className={`flex items-center justify-center py-2 text-sm font-mono transition-colors ${
                      archiveActive ? 'text-primary' : 'text-muted-foreground hover:text-primary'
                    }`}
                  >
                    <Database size={15} />
                  </Link>
                );
              }

              return (
                <div key={archive.id} className="flex items-center group">
                  <Link
                    to={`/docs/${archive.id}`}
                    className={`flex-1 flex items-center gap-1.5 px-3 py-1.5 text-sm font-mono transition-colors ${
                      archiveActive
                        ? 'text-primary bg-primary/5'
                        : 'text-muted-foreground hover:text-primary'
                    }`}
                  >
                    <Database size={11} />
                    <span className="truncate">{archive.name}</span>
                  </Link>
                  <Link
                    to={`/docs/${archive.id}/new`}
                    className="px-1 text-muted-foreground opacity-0 group-hover:opacity-100 hover:text-primary transition"
                    title="New document"
                  >
                    <Plus size={11} />
                  </Link>
                  <button
                    onClick={() => handleDeleteArchive(archive)}
                    className="px-1.5 text-muted-foreground opacity-0 group-hover:opacity-100 hover:text-destructive transition"
                    title="Delete archive"
                  >
                    <Trash2 size={9} />
                  </button>
                </div>
              );
            })}
            {!collapsed && archiveList.length === 0 && (
              <p className="px-3 py-2 text-sm font-mono text-muted-foreground">No archives</p>
            )}
          </div>
        </nav>

        <div className="px-2 py-1.5 border-t border-border">
          <button
            onClick={toggleCollapsed}
            className="w-full flex items-center justify-center text-muted-foreground hover:text-primary transition"
            title={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {collapsed ? <ChevronsRight size={15} /> : <ChevronsLeft size={15} />}
          </button>
        </div>

        <div className={`border-t border-border ${collapsed ? 'p-1.5' : 'p-3'}`}>
          <div className={`flex items-center ${collapsed ? 'justify-center' : 'gap-2'}`}>
            <div className="w-5 h-5 rounded bg-primary flex items-center justify-center text-primary-foreground font-bold text-sm shrink-0">
              {username?.[0]?.toUpperCase() || '?'}
            </div>
            {!collapsed && (
              <p className="text-foreground text-sm font-mono truncate flex-1 min-w-0">
                {username || 'USER'}
              </p>
            )}
          </div>
        </div>
      </aside>
      <ModalPortal config={modalConfig} onClose={close} />
    </>
  );
}
