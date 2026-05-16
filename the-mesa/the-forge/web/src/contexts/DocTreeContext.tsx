import { createContext, useContext, useState, useCallback } from 'react';

const DocTreeContext = createContext<{ refreshTree: () => void; refreshKey: number }>({
  refreshTree: () => {},
  refreshKey: 0,
});

export function DocTreeProvider({ children }: { children: React.ReactNode }) {
  const [refreshKey, setRefreshKey] = useState(0);
  const refreshTree = useCallback(() => setRefreshKey((k) => k + 1), []);
  return (
    <DocTreeContext.Provider value={{ refreshTree, refreshKey }}>
      {children}
    </DocTreeContext.Provider>
  );
}

export function useDocTreeRefresh() {
  return useContext(DocTreeContext);
}
