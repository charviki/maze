import { useEffect, useRef } from 'react';
import { cn } from '../../utils';
import { useAnimationSettings } from './AnimationSettings';

interface TerrainBackgroundProps {
  className?: string;
}

// 确定性伪随机数生成器，保证每次挂载生成一致的地形
function mulberry32(seed: number) {
  return () => {
    let t = (seed += 0x6d2b79f5);
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

interface Vertex {
  baseX: number;
  baseY: number;
  phase: number;
  amp: number;
}

interface Beacon {
  x: number;
  y: number;
  phase: number;
}

interface PulseWave {
  startTime: number;
}

interface Particle {
  x: number;
  y: number;
  vx: number;
  vy: number;
  birth: number;
  alpha: number;
}

const SEED = 0xdeadbeef;
const GRID_COLS = 8;
const GRID_ROWS = 7;
const BEACON_COUNT = 4;
const FRAME_INTERVAL = 1000 / 20;
const PULSE_INTERVAL = 4000;
const PULSE_DURATION = 6000;
const PARTICLE_COUNT = 25;
const PARTICLE_LIFETIME = 3000;

function getPrimaryColor(): string {
  const raw = getComputedStyle(document.body).getPropertyValue('--terminal-foreground').trim();
  if (!raw) return 'hsla(180, 100%, 50%)';
  const parts = raw.split(/\s+/);
  if (parts.length >= 3) {
    const h = parts[0];
    const s = parts[1];
    const l = parts[2].replace('%', '');
    const boostedL = Math.max(parseFloat(l), 50);
    return `hsla(${h}, ${s}, ${boostedL}%)`;
  }
  return 'hsla(180, 100%, 50%)';
}

function generateVertices(w: number, h: number, rng: () => number): Vertex[] {
  const vertices: Vertex[] = [];
  for (let row = 0; row < GRID_ROWS; row++) {
    for (let col = 0; col < GRID_COLS; col++) {
      vertices.push({
        baseX: (col / (GRID_COLS - 1)) * w + (rng() - 0.5) * (w / GRID_COLS) * 0.6,
        baseY: (row / (GRID_ROWS - 1)) * h + (rng() - 0.5) * (h / GRID_ROWS) * 0.6,
        phase: rng() * Math.PI * 2,
        amp: 1 + rng() * 2,
      });
    }
  }
  return vertices;
}

function generateBeacons(w: number, h: number, rng: () => number): Beacon[] {
  const beacons: Beacon[] = [];
  for (let i = 0; i < BEACON_COUNT; i++) {
    beacons.push({
      x: w * 0.15 + rng() * w * 0.7,
      y: h * 0.15 + rng() * h * 0.7,
      phase: rng() * Math.PI * 2,
    });
  }
  return beacons;
}

export function TerrainBackground({ className }: TerrainBackgroundProps) {
  const { settings } = useAnimationSettings();
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (!settings.canvasBackground) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animFrameId: number;
    // 跟踪 rAF 循环是否正在运行，避免重复启动
    let isRunning = false;
    let vertices: Vertex[] = [];
    let beacons: Beacon[] = [];
    let lastFrameTime = 0;
    let lastPulseTime = 0;
    const pulseWaves: PulseWave[] = [];
    const particles: Particle[] = [];
    let cachedRect = { width: 0, height: 0 };

    // 预计算 HSL 基础颜色模板，避免每帧重复解析字符串
    // cachedPrimary 格式为 "hsla(H, S%, L%)"，供 colorTpl 快速拼接透明度
    let cachedPrimary = '';
    const updateColorCache = () => {
      cachedPrimary = getPrimaryColor();
    };

    // 根据预计算的基础色快速生成带透明度的 hsla 字符串
    const colorTpl = (alpha: number) => cachedPrimary.replace(')', `, ${alpha})`);

    const resize = () => {
      const dpr = window.devicePixelRatio || 1;
      const rect =
        canvas.parentElement?.getBoundingClientRect() || document.body.getBoundingClientRect();

      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      canvas.style.width = `${rect.width}px`;
      canvas.style.height = `${rect.height}px`;
      cachedRect = { width: rect.width, height: rect.height };

      // 窗口大小变化时重新生成地形网格和信标
      const rng = mulberry32(SEED);
      vertices = generateVertices(rect.width, rect.height, rng);
      beacons = generateBeacons(rect.width, rect.height, rng);
      updateColorCache();
    };

    updateColorCache();

    window.addEventListener('resize', resize);
    resize();

    const getVertexPos = (v: Vertex, t: number) => ({
      x: v.baseX + Math.sin(t * 0.0005 + v.phase) * v.amp,
      y: v.baseY + Math.cos(t * 0.0004 + v.phase * 1.3) * v.amp,
    });

    // 监听 <html> 的 class/style 变化以检测主题切换
    const themeObserver = new MutationObserver(() => {
      updateColorCache();
    });
    themeObserver.observe(document.documentElement, {
      attributes: true,
      attributeFilter: ['class', 'style'],
    });

    const startAnimation = () => {
      if (isRunning) return;
      isRunning = true;
      animFrameId = requestAnimationFrame(draw);
    };

    const stopAnimation = () => {
      isRunning = false;
      cancelAnimationFrame(animFrameId);
    };

    const draw = (timestamp: number) => {
      if (!isRunning) return;
      animFrameId = requestAnimationFrame(draw);

      if (timestamp - lastFrameTime < FRAME_INTERVAL) return;
      lastFrameTime = timestamp;

      const w = cachedRect.width;
      const h = cachedRect.height;

      ctx.clearRect(0, 0, w, h);

      // 第一层：低多边形地形网格
      const posCache = vertices.map((v) => getVertexPos(v, timestamp));

      for (let row = 0; row < GRID_ROWS - 1; row++) {
        for (let col = 0; col < GRID_COLS - 1; col++) {
          const i = row * GRID_COLS + col;
          const a = posCache[i];
          const b = posCache[i + 1];
          const c = posCache[i + GRID_COLS];
          const d = posCache[i + GRID_COLS + 1];

          const hueShift1 = ((a.x + b.x + c.x) % 30) * 0.5;
          ctx.fillStyle = colorTpl(0.03 + hueShift1 * 0.001);
          ctx.beginPath();
          ctx.moveTo(a.x, a.y);
          ctx.lineTo(b.x, b.y);
          ctx.lineTo(c.x, c.y);
          ctx.fill();

          const hueShift2 = ((b.x + c.x + d.x) % 30) * 0.5;
          ctx.fillStyle = colorTpl(0.03 + hueShift2 * 0.001);
          ctx.beginPath();
          ctx.moveTo(b.x, b.y);
          ctx.lineTo(d.x, d.y);
          ctx.lineTo(c.x, c.y);
          ctx.fill();
        }
      }

      ctx.strokeStyle = colorTpl(0.18);
      ctx.lineWidth = 0.8;
      ctx.beginPath();
      for (let row = 0; row < GRID_ROWS - 1; row++) {
        for (let col = 0; col < GRID_COLS - 1; col++) {
          const i = row * GRID_COLS + col;
          const a = posCache[i];
          const b = posCache[i + 1];
          const c = posCache[i + GRID_COLS];
          const d = posCache[i + GRID_COLS + 1];

          ctx.moveTo(a.x, a.y);
          ctx.lineTo(b.x, b.y);
          ctx.moveTo(a.x, a.y);
          ctx.lineTo(c.x, c.y);
          ctx.moveTo(a.x, a.y);
          ctx.lineTo(d.x, d.y);
        }
      }
      // 补全右边缘和下边缘
      for (let row = 0; row < GRID_ROWS - 1; row++) {
        const i = row * GRID_COLS + GRID_COLS - 1;
        ctx.moveTo(posCache[i].x, posCache[i].y);
        ctx.lineTo(posCache[i + GRID_COLS].x, posCache[i + GRID_COLS].y);
      }
      for (let col = 0; col < GRID_COLS - 1; col++) {
        const i = (GRID_ROWS - 1) * GRID_COLS + col;
        ctx.moveTo(posCache[i].x, posCache[i].y);
        ctx.lineTo(posCache[i + 1].x, posCache[i + 1].y);
      }
      ctx.stroke();

      // 第二层：等高线脉冲波
      if (timestamp - lastPulseTime > PULSE_INTERVAL) {
        pulseWaves.push({ startTime: timestamp });
        lastPulseTime = timestamp;
      }
      const cx = w / 2;
      const cy = h / 2;
      const maxRadius = Math.sqrt(cx * cx + cy * cy);
      for (let i = pulseWaves.length - 1; i >= 0; i--) {
        const wave = pulseWaves[i];
        const elapsed = timestamp - wave.startTime;
        if (elapsed > PULSE_DURATION) {
          pulseWaves.splice(i, 1);
          continue;
        }
        const progress = elapsed / PULSE_DURATION;
        const radius = progress * maxRadius;
        const opacity = 0.18 * (1 - progress);
        ctx.beginPath();
        ctx.arc(cx, cy, radius, 0, Math.PI * 2);
        ctx.strokeStyle = colorTpl(opacity);
        ctx.lineWidth = 1.0;
        ctx.stroke();
      }

      // 第三层：监测信标
      for (const beacon of beacons) {
        // 非线性呼吸节奏：双频率正弦叠加
        const breath =
          Math.sin(timestamp * 0.002 + beacon.phase) *
          Math.sin(timestamp * 0.0045 + beacon.phase * 0.7);
        const intensity = 0.3 + 0.7 * (0.5 + 0.5 * breath);
        const glowRadius = 12 + breath * 6;

        // 外圈辉光
        const gradient = ctx.createRadialGradient(
          beacon.x,
          beacon.y,
          0,
          beacon.x,
          beacon.y,
          glowRadius,
        );
        gradient.addColorStop(0, colorTpl(0.4 * intensity));
        gradient.addColorStop(1, colorTpl(0));
        ctx.fillStyle = gradient;
        ctx.fillRect(beacon.x - glowRadius, beacon.y - glowRadius, glowRadius * 2, glowRadius * 2);

        // 核心点
        ctx.beginPath();
        ctx.arc(beacon.x, beacon.y, 2.5, 0, Math.PI * 2);
        ctx.fillStyle = colorTpl(0.8 * intensity);
        ctx.fill();
      }

      if (beacons.length > 0 && timestamp % 3 < 1) {
        const beacon = beacons[Math.floor(Math.random() * beacons.length)];
        if (particles.length < PARTICLE_COUNT) {
          const angle = Math.random() * Math.PI * 2;
          const speed = 0.3 + Math.random() * 0.5;
          particles.push({
            x: beacon.x,
            y: beacon.y,
            vx: Math.cos(angle) * speed,
            vy: Math.sin(angle) * speed,
            birth: timestamp,
            alpha: 0.6 + Math.random() * 0.3,
          });
        }
      }

      for (let i = particles.length - 1; i >= 0; i--) {
        const pt = particles[i];
        const age = timestamp - pt.birth;
        if (age > PARTICLE_LIFETIME) {
          particles.splice(i, 1);
          continue;
        }
        pt.x += pt.vx;
        pt.y += pt.vy;
        const lifeRatio = age / PARTICLE_LIFETIME;
        const alpha = pt.alpha * (1 - lifeRatio);

        ctx.beginPath();
        ctx.arc(pt.x, pt.y, 1.5, 0, Math.PI * 2);
        ctx.fillStyle = colorTpl(alpha);
        ctx.fill();
      }
    };

    // 页面可见性感知：不可见时暂停 rAF，可见时恢复
    const handleVisibility = () => {
      if (document.visibilityState === 'hidden') {
        stopAnimation();
      } else {
        startAnimation();
      }
    };
    document.addEventListener('visibilitychange', handleVisibility);

    startAnimation();

    return () => {
      stopAnimation();
      window.removeEventListener('resize', resize);
      document.removeEventListener('visibilitychange', handleVisibility);
      themeObserver.disconnect();
    };
  }, [settings.canvasBackground]);

  if (!settings.canvasBackground) {
    return (
      <div className={cn('absolute inset-0 pointer-events-none z-0', className)}>
        <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,rgba(0,255,255,0.05)_0%,transparent_70%)]" />
        <div
          className="absolute bottom-2 left-3 text-[8px] font-mono tracking-widest whitespace-nowrap"
          style={{ color: 'hsl(var(--primary) / 0.3)' }}
        >
          TERRAIN: SWEETWATER // RENDER: OFF
        </div>
      </div>
    );
  }

  return (
    <div className={cn('absolute inset-0 pointer-events-none z-0', className)}>
      <canvas ref={canvasRef} className="absolute inset-0 w-full h-full" aria-hidden="true" />
      <div
        className="absolute bottom-2 left-3 text-[8px] font-mono tracking-widest whitespace-nowrap"
        style={{ color: 'hsl(var(--primary) / 0.3)' }}
      >
        TERRAIN: SWEETWATER // ELEV_SCAN: ACTIVE
      </div>
      <div
        className="absolute bottom-2 right-3 text-[8px] font-mono tracking-widest whitespace-nowrap"
        style={{ color: 'hsl(var(--primary) / 0.3)' }}
      >
        GRID_REF: 34.0522N // 118.2437W
      </div>
    </div>
  );
}
