import { createContext, useContext } from 'react';

interface SidebarContextType {
  collapsed: boolean;
  toggleCollapsed: () => void;
}

export const SidebarContext = createContext<SidebarContextType>({
  collapsed: false,
  toggleCollapsed: () => {},
});

export const useSidebar = () => useContext(SidebarContext);
