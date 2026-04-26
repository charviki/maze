import { useEffect, useRef, useCallback } from 'react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import { AttachAddon } from '@xterm/addon-attach';
import '@xterm/xterm/css/xterm.css';

interface XtermTerminalProps {
  wsUrl: string;
  backgroundComponent?: React.ReactNode;
  theme?: Record<string, string>;
  allowTransparency?: boolean;
}

function toThemeHslValue(value: string, alpha?: number) {
  const normalized = value.trim();
  if (!normalized) return undefined;
  return alpha === undefined ? `hsl(${normalized})` : `hsl(${normalized} / ${alpha})`;
}

function getFallbackTerminalTheme() {
  return {
    background: '#0d1117',
    foreground: '#38bdf8',
    cursor: '#38bdf8',
    cursorAccent: '#0d1117',
    selectionBackground: 'rgba(56, 189, 248, 0.3)',
  };
}

function getThemeSourceElement(element?: HTMLElement | null): HTMLElement {
  if (!element) return document.documentElement;
  const themedAncestor = element.closest('.dark, .light') as HTMLElement | null;
  return themedAncestor ?? element.parentElement ?? element;
}

function readTerminalThemeFromDOM(element?: HTMLElement | null) {
  if (typeof window === 'undefined') {
    return getFallbackTerminalTheme();
  }

  const source = getThemeSourceElement(element);
  const style = getComputedStyle(source);

  return {
    background: toThemeHslValue(style.getPropertyValue('--terminal-background')) ?? '#0d1117',
    foreground: toThemeHslValue(style.getPropertyValue('--terminal-foreground')) ?? '#38bdf8',
    cursor: toThemeHslValue(style.getPropertyValue('--terminal-cursor')) ?? '#38bdf8',
    cursorAccent: toThemeHslValue(style.getPropertyValue('--terminal-cursor-accent')) ?? '#0d1117',
    selectionBackground: toThemeHslValue(style.getPropertyValue('--terminal-selection'), 0.3) ?? 'rgba(56, 189, 248, 0.3)',
  };
}

function buildTerminalTheme(overrides?: Record<string, string>, element?: HTMLElement | null) {
  return { ...readTerminalThemeFromDOM(element), ...overrides };
}

function isTransparentBackground(bg?: string) {
  const v = bg?.trim().toLowerCase();
  return v === 'transparent' || v === 'rgba(0, 0, 0, 0)' || v === 'rgba(0,0,0,0)';
}

export function XtermTerminal({ wsUrl, backgroundComponent, theme, allowTransparency = false }: XtermTerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const resizeFrameRef = useRef<number | null>(null);

  const bgRef = useRef(backgroundComponent);
  bgRef.current = backgroundComponent;
  const themeRef = useRef(theme);
  themeRef.current = theme;

  const sendResize = useCallback((cols: number, rows: number) => {
    const ws = wsRef.current;
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: 'resize', cols, rows }));
    }
  }, []);

  const fitAndSync = useCallback(() => {
    const term = xtermRef.current;
    const fitAddon = fitAddonRef.current;
    if (!term || !fitAddon) return;

    fitAddon.fit();

    // 终端尺寸需要同步到后端 PTY/tmux，否则只放大前端容器时，
    // shell 仍会按旧的 80x24 网格排版，视觉上就像只占左上角一小块。
    if (term.cols > 0 && term.rows > 0) {
      sendResize(term.cols, term.rows);
    }
  }, [sendResize]);

  const scheduleFitAndSync = useCallback(() => {
    if (typeof window === 'undefined') return;
    if (resizeFrameRef.current !== null) {
      window.cancelAnimationFrame(resizeFrameRef.current);
    }
    // 使用 rAF 等待父容器布局稳定，避免在折叠侧栏、面板切换时拿到旧尺寸。
    resizeFrameRef.current = window.requestAnimationFrame(() => {
      resizeFrameRef.current = null;
      fitAndSync();
    });
  }, [fitAndSync]);

  // Effect 1: Terminal 实例创建 + 主题管理
  // 仅在视觉配置变化时重建 Terminal，不影响 WebSocket 连接
  useEffect(() => {
    if (!terminalRef.current) return;

    const resolvedTheme = buildTerminalTheme(theme, terminalRef.current);
    const transparent = !!backgroundComponent || allowTransparency || isTransparentBackground(resolvedTheme.background);

    if (backgroundComponent) {
      resolvedTheme.background = 'transparent';
    }

    const term = new Terminal({
      theme: resolvedTheme,
      fontFamily: "'JetBrains Mono', 'Fira Code', 'Cascadia Code', monospace",
      fontSize: 13,
      cursorBlink: true,
      disableStdin: false,
      allowTransparency: transparent,
      scrollback: 5000,
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalRef.current);

    if (backgroundComponent) {
      const viewport = terminalRef.current.querySelector('.xterm-viewport') as HTMLElement | null;
      if (viewport) {
        viewport.style.backgroundColor = 'transparent';
      }
      const scrollable = terminalRef.current.querySelector('.xterm-scrollable-element') as HTMLElement | null;
      if (scrollable) {
        scrollable.style.backgroundColor = 'transparent';
      }
    }

    xtermRef.current = term;
    fitAddonRef.current = fitAddon;
    scheduleFitAndSync();

    const handleThemeChange = () => {
      if (!xtermRef.current || !terminalRef.current) return;
      const newTheme = buildTerminalTheme(themeRef.current, terminalRef.current);
      const currentTheme = xtermRef.current.options.theme;
      const changed = !currentTheme ||
        currentTheme.foreground !== newTheme.foreground ||
        currentTheme.background !== newTheme.background ||
        currentTheme.cursor !== newTheme.cursor;
      if (changed) {
        xtermRef.current.options.theme = newTheme;
      }
      if (bgRef.current) {
        const viewport = terminalRef.current.querySelector('.xterm-viewport') as HTMLElement | null;
        if (viewport && viewport.style.backgroundColor !== 'transparent') {
          viewport.style.backgroundColor = 'transparent';
        }
        const scrollable = terminalRef.current.querySelector('.xterm-scrollable-element') as HTMLElement | null;
        if (scrollable && scrollable.style.backgroundColor !== 'transparent') {
          scrollable.style.backgroundColor = 'transparent';
        }
      }
    };

    const themeObserver = new MutationObserver(handleThemeChange);
    const nodesToObserve: HTMLElement[] = [document.documentElement, document.body];
    const themeSource = getThemeSourceElement(terminalRef.current);
    if (!nodesToObserve.includes(themeSource)) {
      nodesToObserve.push(themeSource);
    }
    for (const node of nodesToObserve) {
      themeObserver.observe(node, { attributes: true, attributeFilter: ['class', 'style'] });
    }

    const darkSchemeQuery = window.matchMedia('(prefers-color-scheme: dark)');
    darkSchemeQuery.addEventListener('change', handleThemeChange);

    return () => {
      darkSchemeQuery.removeEventListener('change', handleThemeChange);
      themeObserver.disconnect();
      if (resizeFrameRef.current !== null) {
        window.cancelAnimationFrame(resizeFrameRef.current);
        resizeFrameRef.current = null;
      }
      term.dispose();
      xtermRef.current = null;
      fitAddonRef.current = null;
    };
  }, [backgroundComponent, theme, allowTransparency, scheduleFitAndSync]);

  // Effect 2: WebSocket 连接
  // 仅在 wsUrl 变化时重建连接，通过 ref 读取 Terminal 实例
  useEffect(() => {
    const term = xtermRef.current;
    const fitAddon = fitAddonRef.current;
    if (!term || !fitAddon) return;

    const ws = new WebSocket(wsUrl);
    ws.binaryType = 'arraybuffer';
    wsRef.current = ws;

    ws.addEventListener('error', () => {});

    ws.addEventListener('open', () => {
      if (ws.readyState !== WebSocket.OPEN) return;
      const attachAddon = new AttachAddon(ws);
      term.loadAddon(attachAddon);

      // WebSocket 连接建立后立即显式同步一次尺寸，避免 attach 后 tmux 仍停留在默认 80x24。
      scheduleFitAndSync();
    });

    // 终端尺寸更多受父容器布局影响，而不只是 window.resize，
    // 因此同时监听容器本身，确保侧栏折叠或面板重排后也能同步到后端。
    const handleResize = () => {
      scheduleFitAndSync();
    };
    window.addEventListener('resize', handleResize);

    const resizeObserver = new ResizeObserver(() => {
      scheduleFitAndSync();
    });
    if (terminalRef.current) {
      resizeObserver.observe(terminalRef.current);
    }

    const resizeDisposable = term.onResize(({ cols, rows }) => {
      sendResize(cols, rows);
    });

    return () => {
      window.removeEventListener('resize', handleResize);
      resizeObserver.disconnect();
      resizeDisposable.dispose();
      ws.close();
      wsRef.current = null;
    };
  }, [wsUrl, sendResize, scheduleFitAndSync]);

  return (
    <div className="relative h-full w-full overflow-hidden rounded-md bg-transparent">
      {backgroundComponent}
      <div ref={terminalRef} className="absolute inset-0 z-10 h-full w-full" />
    </div>
  );
}
