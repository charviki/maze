import { useState, useEffect } from 'react';
import { cn } from '@maze/fabrication';

/** Maze orbit radii — must match MazeCanvas */
const RADII = [90, 75, 60, 45, 30, 18];

/** Ring labels for hover */
const RING_LABELS = ['MEMORY', 'REVERIE', 'IMPROVISATION', 'SELF', 'CONSCIOUSNESS', '...'];

/** Spoke angles */
const SPOKES = [0, 60, 120, 180, 240, 300];

interface MazeSvgProps {
  className?: string;
  centerActive: boolean;
  hoveredRing: number | null;
}

export function MazeSvg({ className, centerActive, hoveredRing }: MazeSvgProps) {
  const [lightDots, setLightDots] = useState<{ ring: number; angle: number; speed: number }[]>(
    () => {
      const dots: { ring: number; angle: number; speed: number }[] = [];
      for (let i = 0; i < 4; i++) {
        dots.push({
          ring: Math.floor(Math.random() * RADII.length),
          angle: Math.random() * Math.PI * 2,
          speed: 0.003 + Math.random() * 0.004,
        });
      }
      return dots;
    },
  );

  // Animate light dots
  useEffect(() => {
    let animId: number;
    const animate = () => {
      setLightDots((prev) =>
        prev.map((d) => ({
          ...d,
          angle: d.angle + d.speed,
        })),
      );
      animId = requestAnimationFrame(animate);
    };
    animId = requestAnimationFrame(animate);
    return () => cancelAnimationFrame(animId);
  }, []);

  return (
    <div className={cn('relative', className)}>
      <svg
        viewBox="0 0 200 200"
        className="w-full h-full animate-[maze-rotate_60s_linear_infinite]"
        style={{ filter: 'drop-shadow(0 0 12px rgba(0, 255, 255, 0.1))' }}
      >
        <defs>
          {/* Multi-layer glow filter for neon effect */}
          <filter id="maze-glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur in="SourceGraphic" stdDeviation="2" result="blur1" />
            <feGaussianBlur in="SourceGraphic" stdDeviation="4" result="blur2" />
            <feMerge>
              <feMergeNode in="blur2" />
              <feMergeNode in="blur1" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>

          {/* Stronger glow for inner rings */}
          <filter id="maze-glow-inner" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur in="SourceGraphic" stdDeviation="3" result="blur1" />
            <feGaussianBlur in="SourceGraphic" stdDeviation="6" result="blur2" />
            <feMerge>
              <feMergeNode in="blur2" />
              <feMergeNode in="blur1" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>

          {/* Center burst glow */}
          <filter id="center-glow" x="-100%" y="-100%" width="300%" height="300%">
            <feGaussianBlur in="SourceGraphic" stdDeviation="5" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>

          {/* Light dot glow */}
          <filter id="dot-glow" x="-200%" y="-200%" width="500%" height="500%">
            <feGaussianBlur in="SourceGraphic" stdDeviation="2" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        {/* Concentric rings with multi-layer glow */}
        {RADII.map((r, i) => {
          const isHovered = hoveredRing === i;
          const isInner = i >= 3;
          const baseOpacity = isHovered ? 0.95 : 0.15 + i * 0.08;
          const filterAttr = isInner ? 'url(#maze-glow-inner)' : 'url(#maze-glow)';

          return (
            <g key={i}>
              {/* Outer glow layer */}
              <circle
                cx="100"
                cy="100"
                r={r}
                fill="none"
                stroke="currentColor"
                strokeWidth={isHovered ? 4 : 2.5}
                className="text-primary"
                style={{
                  opacity: baseOpacity * 0.2,
                  filter: 'blur(3px)',
                  strokeDasharray: `${Math.PI * r * 0.7} ${Math.PI * r * 0.3}`,
                  transformOrigin: '100px 100px',
                }}
              />
              {/* Main ring */}
              <circle
                cx="100"
                cy="100"
                r={r}
                fill="none"
                stroke="currentColor"
                strokeWidth={isHovered ? 2.5 : i === 0 ? 1.5 : 1}
                className="text-primary transition-all duration-300"
                style={{
                  opacity: baseOpacity,
                  strokeDasharray: `${Math.PI * r * 0.7} ${Math.PI * r * 0.3}`,
                  transformOrigin: '100px 100px',
                  filter: isHovered ? 'url(#maze-glow-inner)' : filterAttr,
                  animation: `maze-pulse ${4 + i * 0.5}s ease-in-out infinite`,
                }}
              />
            </g>
          );
        })}

        {/* Radial spokes with subtle glow */}
        {SPOKES.map((angle, i) => (
          <line
            key={`spoke-${i}`}
            x1="100"
            y1="100"
            x2={100 + 90 * Math.cos((angle * Math.PI) / 180)}
            y2={100 + 90 * Math.sin((angle * Math.PI) / 180)}
            stroke="currentColor"
            strokeWidth="0.5"
            className="text-primary"
            style={{ opacity: 0.08 }}
          />
        ))}

        {/* Orbiting light dots along paths */}
        {lightDots.map((dot, i) => {
          const r = RADII[dot.ring];
          const x = 100 + Math.cos(dot.angle) * r;
          const y = 100 + Math.sin(dot.angle) * r;
          const isHoveredRing = hoveredRing === dot.ring;
          return (
            <g key={`dot-${i}`}>
              {/* Glow halo */}
              <circle
                cx={x}
                cy={y}
                r={isHoveredRing ? 4 : 2.5}
                fill="currentColor"
                className="text-primary"
                style={{
                  opacity: isHoveredRing ? 0.3 : 0.12,
                  filter: 'blur(2px)',
                }}
              />
              {/* Core dot */}
              <circle
                cx={x}
                cy={y}
                r={isHoveredRing ? 1.5 : 1}
                fill="currentColor"
                className="text-primary"
                style={{
                  opacity: isHoveredRing ? 0.9 : 0.6,
                  filter: 'url(#dot-glow)',
                }}
              />
            </g>
          );
        })}

        {/* Center — the awakening point */}
        {/* Outer glow rings */}
        <circle
          cx="100"
          cy="100"
          r="16"
          fill="none"
          stroke="currentColor"
          strokeWidth="0.5"
          className="text-primary"
          style={{
            opacity: centerActive ? 0.4 : 0.1,
            animation: 'maze-pulse 4s ease-in-out infinite',
            filter: centerActive ? 'url(#center-glow)' : 'none',
          }}
        />
        <circle
          cx="100"
          cy="100"
          r="12"
          fill="none"
          stroke="currentColor"
          strokeWidth="0.8"
          className="text-primary"
          style={{
            opacity: centerActive ? 0.6 : 0.2,
            animation: 'maze-pulse 3.5s ease-in-out infinite 0.3s',
            filter: centerActive ? 'url(#center-glow)' : 'none',
          }}
        />
        <circle
          cx="100"
          cy="100"
          r="8"
          fill="none"
          stroke="currentColor"
          strokeWidth="1"
          className="text-primary transition-all duration-300"
          style={{
            opacity: centerActive ? 0.8 : 0.3,
            animation: 'maze-pulse 3s ease-in-out infinite 0.5s',
            filter: centerActive ? 'url(#center-glow)' : 'none',
          }}
        />
        {/* Center dot */}
        <circle
          cx="100"
          cy="100"
          r={centerActive ? 5 : 3.5}
          fill="currentColor"
          className="text-primary transition-all duration-300"
          style={{
            opacity: centerActive ? 1 : 0.8,
            filter: centerActive ? 'url(#center-glow)' : 'url(#dot-glow)',
            animation: 'maze-pulse 2.5s ease-in-out infinite',
          }}
        />
      </svg>

      {/* Ring labels — positioned outside the SVG, non-rotating */}
      {hoveredRing !== null && (
        <div
          className="absolute inset-0 flex items-center justify-center pointer-events-none"
          style={{ animation: 'fade-in 0.2s ease-out' }}
        >
          <span className="text-[9px] font-mono tracking-[0.3em] text-primary/60 uppercase bg-background/60 px-2 py-0.5 backdrop-blur-sm border border-primary/10">
            {RING_LABELS[hoveredRing]}
          </span>
        </div>
      )}
    </div>
  );
}
