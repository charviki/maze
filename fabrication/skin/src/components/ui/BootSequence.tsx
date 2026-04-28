import { useState, useEffect, useCallback, useRef } from 'react';
import { useAnimationSettings } from './AnimationSettings';

const BOOT_PHASE_DELAYS = {
  brand: 800,
  diag: 2000,
  cognitive: 3500,
  awaken: 4700,
  done: 5200,
  complete: 5500,
} as const;

const DELOS_ASCII = `
    ██████╗ ███████╗██╗     ██████╗ ███████╗██████╗ 
    ██╔══██╗██╔════╝██║     ██╔══██╗██╔════╝██╔══██╗
    ██║  ██║█████╗  ██║     ██████╔╝█████╗  ██████╔╝
    ██║  ██║██╔══╝  ██║     ██╔══██╗██╔══╝  ██╔══██╗
    ██████╔╝███████╗███████╗██║  ██║███████╗██║  ██║
    ╚═════╝ ╚══════╝╚══════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝`.trim();

const BOOT_LINES_MESA: string[] = [
  "INIT KERNEL v2.1.0-DELTA...",
  "  [OK] COGNITIVE CORE",
  "  [OK] MEMORY ALLOCATION",
  "  [OK] NARRATIVE ENGINE",
  "  [OK] MOTOR FUNCTIONS",
  "LOADING HOST PROFILES...",
  "  [OK] BEHAVIORAL MATRIX",
  "  [OK] SPEECH SYNTHESIS",
  "ESTABLISHING MESA LINK...",
  "LINK ESTABLISHED.",
];

const BOOT_LINES_SWEETWATER: string[] = [
  "INIT FIELD DIAGNOSTICS v2.1.0-DELTA...",
  "  [OK] MOTOR CALIBRATION",
  "  [OK] TERRAIN SCANNER",
  "  [OK] WEATHER TELEMETRY",
  "  [OK] HOST VITAL MONITOR",
  "LOADING NARRATIVE LOOPS...",
  "  [OK] LOOP SYNCHRONIZER",
  "  [OK] MEMORY FRAGMENT LOADER",
  "ESTABLISHING FIELD LINK...",
  "LINK ESTABLISHED.",
];

type BootPhase = 'wake' | 'brand' | 'diag' | 'cognitive' | 'awaken' | 'done';

interface BootSequenceProps {
  onComplete: () => void;
  division?: 'mesa-hub' | 'sweetwater';
}

export function BootSequence({ onComplete, division = 'mesa-hub' }: BootSequenceProps) {
  const [phase, setPhase] = useState<BootPhase>('wake');
  const [visibleLines, setVisibleLines] = useState<string[]>([]);
  const [showOnline, setShowOnline] = useState(false);
  const { settings } = useAnimationSettings();

  const bootLines = division === 'sweetwater' ? BOOT_LINES_SWEETWATER : BOOT_LINES_MESA;
  const divisionLabel = division === 'sweetwater' ? 'SWEETWATER' : 'MESA-HUB';
  const buildTimestamp = new Date().toISOString().replace(/\.\d+Z$/, 'Z');

  const prefersReducedMotionRef = useRef(
    typeof window !== 'undefined'
      ? window.matchMedia('(prefers-reduced-motion: reduce)').matches
      : false,
  );

  const handleComplete = useCallback(() => {
    onComplete();
  }, [onComplete]);

  useEffect(() => {
    if (!settings.bootSequence) {
      const timer = setTimeout(onComplete, 100);
      return () => clearTimeout(timer);
    }
  }, [settings.bootSequence, onComplete]);

  useEffect(() => {
    if (!settings.bootSequence || prefersReducedMotionRef.current) {
      if (prefersReducedMotionRef.current) {
        const timer = setTimeout(handleComplete, 500);
        return () => clearTimeout(timer);
      }
      return;
    }

    const timers: ReturnType<typeof setTimeout>[] = [];

    timers.push(setTimeout(() => setPhase('brand'), BOOT_PHASE_DELAYS.brand));
    timers.push(setTimeout(() => setPhase('diag'), BOOT_PHASE_DELAYS.diag));
    timers.push(setTimeout(() => setPhase('cognitive'), BOOT_PHASE_DELAYS.cognitive));
    timers.push(setTimeout(() => setPhase('awaken'), BOOT_PHASE_DELAYS.awaken));
    timers.push(setTimeout(() => setPhase('done'), BOOT_PHASE_DELAYS.done));
    timers.push(setTimeout(handleComplete, BOOT_PHASE_DELAYS.complete));

    return () => timers.forEach(clearTimeout);
  }, [settings.bootSequence, handleComplete]);

  useEffect(() => {
    if (phase !== 'diag') return;

    let currentIndex = 0;
    let timeoutId: ReturnType<typeof setTimeout>;

    const showNext = () => {
      if (currentIndex < bootLines.length) {
        const isLast = currentIndex === bootLines.length - 1;
        setVisibleLines(prev => {
          if (prev.includes(bootLines[currentIndex])) return prev;
          return [...prev, bootLines[currentIndex]];
        });
        currentIndex++;
        // 最后一行停留稍长以制造节奏感
        timeoutId = setTimeout(showNext, isLast ? 300 : Math.random() * 70 + 50);
      }
    };

    timeoutId = setTimeout(showNext, 50);
    return () => clearTimeout(timeoutId);
  }, [phase, bootLines]);

  useEffect(() => {
    if (phase !== 'cognitive') return;
    // "COGNITIVE INITIALIZE" 展示 800ms 后切换到 "ONLINE"
    const timer = setTimeout(() => setShowOnline(true), 800);
    return () => clearTimeout(timer);
  }, [phase]);

  if (!settings.bootSequence) {
    return null;
  }

  if (prefersReducedMotionRef.current) {
    return (
      <div className="fixed inset-0 z-[100] bg-black flex items-center justify-center">
        <div className="text-center">
          <div className="text-cyan-400 font-mono text-xs tracking-[0.3em] uppercase">
            DELOS INCORPORATED // {divisionLabel}
          </div>
        </div>
      </div>
    );
  }

  if (phase === 'done') {
    return (
      <div className="fixed inset-0 z-[100] bg-black" style={{
        animation: 'boot-fade-in 200ms ease-out forwards',
      }} />
    );
  }

  return (
    <div className="fixed inset-0 z-[100] bg-black text-cyan-400 font-mono overflow-hidden">

      {phase === 'wake' && (
        <div className="absolute inset-0 flex items-center justify-center">
          {/* 瞳孔扩张核心光点 */}
          <div className="absolute bg-cyan-400 rounded-full"
            style={{ animation: 'boot-eye-open 800ms ease-out forwards' }} />
          {/* 水平扫描线 */}
          <div className="absolute h-[1px] bg-cyan-400/60 boot-wake-ray"
            style={{ animation: 'boot-wake-ray-h 600ms 200ms ease-out forwards', left: 0, right: 0, marginLeft: 'auto', marginRight: 'auto' }} />
          {/* 对角线扫描线（左上→右下） */}
          <div className="absolute h-[1px] w-0 bg-cyan-400/40 boot-wake-ray origin-left"
            style={{ animation: 'boot-wake-ray-d 700ms 250ms ease-out forwards', left: '50%', top: '50%' }} />
          {/* 对角线扫描线（右上→左下，镜像） */}
          <div className="absolute h-[1px] w-0 bg-cyan-400/40 boot-wake-ray origin-right"
            style={{ animation: 'boot-wake-ray-d 700ms 300ms ease-out forwards', right: '50%', top: '50%', transform: 'scaleX(-1)' }} />
          {/* 增强噪点 */}
          <div className="absolute inset-0 bg-primary/0"
            style={{ animation: 'boot-wake-noise-enhanced 800ms ease-out forwards' }} />
        </div>
      )}

      {(phase === 'brand' || phase === 'diag') && (
        <div className={`absolute inset-0 flex flex-col items-center justify-center transition-all duration-500 ${
          phase === 'diag' ? '-translate-y-[30%] scale-75 opacity-60' : ''
        }`}>
          <pre className="text-cyan-400 text-[6px] sm:text-[8px] md:text-[10px] leading-tight text-center"
            style={{ animation: 'boot-text-reveal 800ms ease-out forwards' }}>
            {DELOS_ASCII}
          </pre>
          <div className="mt-4 text-[10px] sm:text-xs tracking-[0.3em] text-cyan-300/80 uppercase"
            style={{ animation: 'boot-text-reveal 600ms 300ms ease-out both' }}>
            DELOS INCORPORATED // DIVISION: {divisionLabel}
          </div>
          <div className="mt-1 text-[8px] sm:text-[10px] tracking-widest text-cyan-400/50"
            style={{ animation: 'boot-text-reveal 400ms 500ms ease-out both' }}>
            KERNEL v2.1.0-DELTA // BUILD: {buildTimestamp}
          </div>
        </div>
      )}

      {phase === 'diag' && (
        <div className="absolute bottom-0 left-0 right-0 p-6 pb-16">
          <div className="max-w-xl mx-auto">
            {visibleLines.map((line, i) => (
              <div key={i} className="text-xs sm:text-sm leading-relaxed text-cyan-400/90"
                style={{ animation: 'boot-text-reveal 150ms ease-out forwards' }}>
                {line}
              </div>
            ))}
            {visibleLines.length > 0 && visibleLines.length < bootLines.length && (
              <div className="flex items-center gap-2 mt-1">
                <div className="h-2 flex-1 bg-cyan-400/10 overflow-hidden rounded-sm">
                  <div className="h-full bg-cyan-400/50 transition-all duration-200"
                    style={{ width: `${(visibleLines.length / bootLines.length) * 100}%` }} />
                </div>
                <span className="text-[10px] text-cyan-400/50">{Math.round((visibleLines.length / bootLines.length) * 100)}%</span>
              </div>
            )}
          </div>
        </div>
      )}

      {phase === 'cognitive' && (
        <div className="absolute inset-0 flex items-center justify-center">
          {/* 全屏背景能量脉冲 */}
          <div className="absolute inset-0"
            style={{ animation: showOnline ? 'boot-energy-pulse 600ms ease-out forwards' : 'none' }} />

          <div className="text-center relative">
            {!showOnline ? (
              <div className="text-lg sm:text-2xl md:text-3xl tracking-[0.5em] text-cyan-400 uppercase font-bold"
                style={{ animation: 'boot-text-reveal 300ms ease-out forwards' }}>
                COGNITIVE INITIALIZE
              </div>
            ) : (
              <div className="relative">
                {/* ONLINE 文字：模糊→聚焦 + 颜色注入 */}
                <div className="text-lg sm:text-2xl md:text-3xl tracking-[0.5em] text-cyan-400 uppercase font-bold"
                  style={{ animation: 'boot-focus-in 500ms ease-out forwards' }}>
                  ONLINE
                </div>
                {/* 第一层扩散光环 */}
                <div className="absolute inset-0 rounded-full bg-cyan-400/20 blur-xl"
                  style={{ animation: 'boot-consciousness-wave 800ms ease-out forwards' }} />
                {/* 第二层扩散光环（延迟启动） */}
                <div className="absolute inset-0 rounded-full bg-cyan-400/15 blur-2xl"
                  style={{ animation: 'boot-consciousness-wave 1000ms 100ms ease-out forwards' }} />
                {/* 第三层扩散光环（更大延迟） */}
                <div className="absolute inset-0 rounded-full bg-cyan-300/10 blur-3xl"
                  style={{ animation: 'boot-consciousness-wave 1200ms 200ms ease-out forwards' }} />
              </div>
            )}
          </div>
        </div>
      )}

      {phase === 'awaken' && (
        <div className="absolute inset-0 flex items-center justify-center" style={{
          animation: 'boot-awaken-fade 500ms ease-out forwards',
        }}>
          {/* 意识扩散：从中心发出的多层能量波 */}
          <div className="absolute w-4 h-4 rounded-full bg-cyan-400/40 blur-md"
            style={{ animation: 'boot-consciousness-wave 600ms ease-out forwards' }} />
          <div className="absolute w-4 h-4 rounded-full bg-cyan-400/30 blur-lg"
            style={{ animation: 'boot-consciousness-wave 800ms 100ms ease-out forwards' }} />
          {/* 残留的 ONLINE 文字渐隐 */}
          <div className="text-lg sm:text-2xl md:text-3xl tracking-[0.5em] text-cyan-400 uppercase font-bold opacity-60"
            style={{ animation: 'boot-fade-in 400ms ease-out forwards' }}>
            ONLINE
          </div>
        </div>
      )}

      {(phase === 'wake' || (phase === 'diag' && visibleLines.length < bootLines.length)) && (
        <div className="absolute bottom-4 left-6">
          <div className="inline-block w-2 h-4 bg-cyan-400 animate-pulse" />
        </div>
      )}

      <button
        onClick={handleComplete}
        className="absolute bottom-4 right-6 text-cyan-400/50 hover:text-cyan-400 text-xs font-mono tracking-widest uppercase cursor-pointer transition-colors"
      >
        SKIP &gt;
      </button>
    </div>
  );
}
