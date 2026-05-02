import { useEffect, useRef, useCallback } from 'react';
import { cn } from '@maze/fabrication';

/** Maze orbit radii — must match MazeSvg */
const RADII = [90, 75, 60, 45, 30, 18];

/** Particles per orbit ring */
const PARTICLES_PER_RING = 35;
/** Floating particles between rings */
const FLOATING_COUNT = 30;
/** Max trail length */
const TRAIL_LENGTH = 4;
/** Ripple lifetime (ms) */
const RIPPLE_LIFETIME = 2000;
/** Ripple max radius */
const RIPPLE_MAX_RADIUS = 100;
/** Mouse repulsion strength */
const REPULSION_STRENGTH = 25;
/** Mouse repulsion radius */
const REPULSION_RADIUS = 35;
/** Frame rate cap */
const FRAME_INTERVAL = 1000 / 30;

interface OrbitalParticle {
  angle: number;
  radius: number;
  targetRadius: number;
  speed: number;
  size: number;
  alpha: number;
  trail: { x: number; y: number }[];
  isFloating: boolean;
}

interface Ripple {
  radius: number;
  alpha: number;
  birth: number;
}

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

function createParticles(): OrbitalParticle[] {
  const particles: OrbitalParticle[] = [];

  // Orbital particles
  for (let ring = 0; ring < RADII.length; ring++) {
    const r = RADII[ring];
    const speed = 0.0003 + (RADII.length - ring) * 0.00012; // outer = slower
    for (let i = 0; i < PARTICLES_PER_RING; i++) {
      particles.push({
        angle: (Math.PI * 2 * i) / PARTICLES_PER_RING + Math.random() * 0.3,
        radius: r,
        targetRadius: r,
        speed: speed * (0.7 + Math.random() * 0.6),
        size: 1 + Math.random() * 1,
        alpha: 0.3 + Math.random() * 0.5,
        trail: [],
        isFloating: false,
      });
    }
  }

  // Floating particles
  for (let i = 0; i < FLOATING_COUNT; i++) {
    const r = 10 + Math.random() * 85;
    particles.push({
      angle: Math.random() * Math.PI * 2,
      radius: r,
      targetRadius: r,
      speed: 0.0001 + Math.random() * 0.0003,
      size: 0.5 + Math.random() * 1,
      alpha: 0.15 + Math.random() * 0.25,
      trail: [],
      isFloating: true,
    });
  }

  return particles;
}

interface MazeCanvasProps {
  className?: string;
  mousePos: { x: number; y: number } | null;
  centerActive: boolean;
}

export function MazeCanvas({ className, mousePos, centerActive }: MazeCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const mousePosRef = useRef(mousePos);
  const centerActiveRef = useRef(centerActive);
  const prevCenterActive = useRef(false);

  // Keep refs in sync without re-running effect
  useEffect(() => {
    mousePosRef.current = mousePos;
  }, [mousePos]);
  useEffect(() => {
    centerActiveRef.current = centerActive;
  }, [centerActive]);

  const initCanvas = useCallback(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animFrameId: number;
    let isRunning = false;
    let lastFrameTime = 0;
    const particles = createParticles();
    const ripples: Ripple[] = [];
    let cachedPrimary = '';

    const updateColor = () => {
      cachedPrimary = getPrimaryColor();
    };
    const colorTpl = (alpha: number) => cachedPrimary.replace(')', `, ${alpha})`);

    // Fixed coordinate space: 200x200 (matches SVG viewBox)
    const SIZE = 200;

    const resize = () => {
      const dpr = window.devicePixelRatio || 1;
      const parent = canvas.parentElement;
      if (!parent) return;
      const rect = parent.getBoundingClientRect();
      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      canvas.style.width = `${rect.width}px`;
      canvas.style.height = `${rect.height}px`;
      updateColor();
    };

    const themeObserver = new MutationObserver(updateColor);
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

      const parent = canvas.parentElement;
      if (!parent) return;
      const rect = parent.getBoundingClientRect();
      const w = rect.width;
      const h = rect.height;

      // Scale context to map 200x200 SVG space to actual pixels
      ctx.save();
      ctx.clearRect(0, 0, w, h);
      ctx.scale(w / SIZE, h / SIZE);

      const cx = SIZE / 2;
      const cy = SIZE / 2;
      const mp = mousePosRef.current;

      // Trigger ripple on center activation
      if (centerActiveRef.current && !prevCenterActive.current) {
        ripples.push({ radius: 5, alpha: 0.8, birth: timestamp });
      }
      prevCenterActive.current = centerActiveRef.current;

      // Update and draw particles
      for (const p of particles) {
        // Store trail
        const px = cx + Math.cos(p.angle) * p.radius;
        const py = cy + Math.sin(p.angle) * p.radius;
        p.trail.push({ x: px, y: py });
        if (p.trail.length > TRAIL_LENGTH) p.trail.shift();

        // Update angle
        p.angle += p.speed;

        // Floating particles drift in radius
        if (p.isFloating) {
          p.targetRadius += Math.sin(timestamp * 0.001 + p.angle * 3) * 0.02;
          p.targetRadius = Math.max(8, Math.min(95, p.targetRadius));
        }

        // Mouse repulsion
        if (mp) {
          const dx = px - mp.x;
          const dy = py - mp.y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < REPULSION_RADIUS && dist > 0) {
            const force = (1 - dist / REPULSION_RADIUS) * REPULSION_STRENGTH;
            const pushAngle = Math.atan2(dy, dx);
            p.radius += Math.cos(pushAngle) * force * 0.05;
          }
        }

        // Center attraction when active
        if (centerActiveRef.current && !p.isFloating) {
          p.radius += (p.targetRadius * 0.3 - p.radius) * 0.02;
        } else {
          // Return to target radius
          p.radius += (p.targetRadius - p.radius) * 0.05;
        }

        // Check ripple intersection
        let rippleBoost = 0;
        for (const ripple of ripples) {
          const distToRipple = Math.abs(p.radius - ripple.radius);
          if (distToRipple < 8) {
            rippleBoost = Math.max(rippleBoost, ripple.alpha * (1 - distToRipple / 8));
          }
        }

        // Draw trail
        if (p.trail.length > 1) {
          ctx.beginPath();
          ctx.moveTo(p.trail[0].x, p.trail[0].y);
          for (let i = 1; i < p.trail.length; i++) {
            ctx.lineTo(p.trail[i].x, p.trail[i].y);
          }
          ctx.strokeStyle = colorTpl(p.alpha * 0.2);
          ctx.lineWidth = p.size * 0.5;
          ctx.stroke();
        }

        // Draw particle
        const finalAlpha = Math.min(1, p.alpha + rippleBoost);
        ctx.beginPath();
        ctx.arc(px, py, p.size, 0, Math.PI * 2);
        ctx.fillStyle = colorTpl(finalAlpha);
        ctx.fill();

        // Glow for bright particles
        if (finalAlpha > 0.6) {
          ctx.beginPath();
          ctx.arc(px, py, p.size * 3, 0, Math.PI * 2);
          ctx.fillStyle = colorTpl(finalAlpha * 0.15);
          ctx.fill();
        }
      }

      // Update and draw ripples
      for (let i = ripples.length - 1; i >= 0; i--) {
        const ripple = ripples[i];
        const age = timestamp - ripple.birth;
        if (age > RIPPLE_LIFETIME) {
          ripples.splice(i, 1);
          continue;
        }
        const progress = age / RIPPLE_LIFETIME;
        ripple.radius = 5 + progress * RIPPLE_MAX_RADIUS;
        ripple.alpha = 0.8 * (1 - progress * progress); // quadratic fade

        ctx.beginPath();
        ctx.arc(cx, cy, ripple.radius, 0, Math.PI * 2);
        ctx.strokeStyle = colorTpl(ripple.alpha);
        ctx.lineWidth = 1.5 * (1 - progress);
        ctx.stroke();

        // Inner glow ring
        ctx.beginPath();
        ctx.arc(cx, cy, ripple.radius * 0.95, 0, Math.PI * 2);
        ctx.strokeStyle = colorTpl(ripple.alpha * 0.3);
        ctx.lineWidth = 4 * (1 - progress);
        ctx.stroke();
      }

      ctx.restore();
    };

    const handleVisibility = () => {
      if (document.visibilityState === 'hidden') {
        stopAnimation();
      } else {
        startAnimation();
      }
    };

    document.addEventListener('visibilitychange', handleVisibility);
    window.addEventListener('resize', resize);
    resize();
    startAnimation();

    return () => {
      stopAnimation();
      window.removeEventListener('resize', resize);
      document.removeEventListener('visibilitychange', handleVisibility);
      themeObserver.disconnect();
    };
  }, []);

  useEffect(initCanvas, [initCanvas]);

  return (
    <canvas
      ref={canvasRef}
      className={cn('absolute inset-0 w-full h-full pointer-events-none', className)}
      aria-hidden="true"
    />
  );
}
