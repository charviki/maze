import { useState, useEffect, useRef } from 'react';
import { cn } from '../../utils';

export interface RadarNode {
  name: string;
  status: 'online' | 'offline';
}

const GOLDEN_ANGLE = 137.508;

function computeBlipPosition(index: number): { cx: number; cy: number } {
  const angle = (index * GOLDEN_ANGLE * Math.PI) / 180;
  const radius = 20 + (index % 3) * 10;
  return {
    cx: 50 + radius * Math.cos(angle),
    cy: 50 + radius * Math.sin(angle),
  };
}

interface PulseLink {
  id: number;
  from: { cx: number; cy: number };
  to: { cx: number; cy: number };
}

interface RadarViewProps {
  className?: string;
  nodes?: RadarNode[];
}

export function RadarView({ className, nodes }: RadarViewProps) {
  const [links, setLinks] = useState<PulseLink[]>([]);
  const linkIdRef = useRef(0);

  // 多个 online 节点时，随机生成连接脉冲线
  useEffect(() => {
    if (!nodes || nodes.length < 2) return;

    const onlineNodes = nodes.filter((n) => n.status === 'online');
    if (onlineNodes.length < 2) return;

    const onlineIndices = nodes
      .map((n, i) => (n.status === 'online' ? i : -1))
      .filter((i) => i >= 0);

    const scheduleNext = () => {
      const delay = 3000 + Math.random() * 5000;
      return setTimeout(() => {
        const a = onlineIndices[Math.floor(Math.random() * onlineIndices.length)];
        let b = a;
        while (b === a && onlineIndices.length > 1) {
          b = onlineIndices[Math.floor(Math.random() * onlineIndices.length)];
        }
        if (a !== b) {
          const posA = computeBlipPosition(a);
          const posB = computeBlipPosition(b);
          const id = ++linkIdRef.current;
          setLinks((prev) => [...prev, { id, from: posA, to: posB }]);
          // 1.5 秒后移除脉冲线
          setTimeout(() => {
            setLinks((prev) => prev.filter((l) => l.id !== id));
          }, 1500);
        }
        scheduleNext();
      }, delay);
    };

    const timer = scheduleNext();
    return () => {
      clearTimeout(timer);
    };
  }, [nodes]);

  return (
    <div
      className={cn(
        'relative aspect-square w-full h-full rounded-full border border-primary/30 bg-background/40 overflow-hidden shadow-[0_0_15px_rgba(0,255,255,0.15)]',
        className,
      )}
    >
      <svg viewBox="0 0 100 100" className="absolute inset-0 w-full h-full">
        <circle
          cx="50"
          cy="50"
          r="45"
          fill="none"
          stroke="hsl(var(--primary) / 0.2)"
          strokeWidth="0.5"
        />
        <circle
          cx="50"
          cy="50"
          r="30"
          fill="none"
          stroke="hsl(var(--primary) / 0.2)"
          strokeWidth="0.5"
        />
        <circle
          cx="50"
          cy="50"
          r="15"
          fill="none"
          stroke="hsl(var(--primary) / 0.2)"
          strokeWidth="0.5"
        />

        <line
          x1="50"
          y1="5"
          x2="50"
          y2="95"
          stroke="hsl(var(--primary) / 0.15)"
          strokeWidth="0.3"
        />
        <line
          x1="5"
          y1="50"
          x2="95"
          y2="50"
          stroke="hsl(var(--primary) / 0.15)"
          strokeWidth="0.3"
        />

        <g className="radar-sweep-svg">
          <line
            x1="50"
            y1="50"
            x2="95"
            y2="50"
            stroke="hsl(var(--primary) / 0.8)"
            strokeWidth="0.8"
          />
          <path d="M50,50 L95,50 A45,45 0 0,0 84,27 Z" fill="hsl(var(--primary))" opacity="0.18" />
          <path d="M50,50 L84,27 A45,45 0 0,0 72,11 Z" fill="hsl(var(--primary))" opacity="0.1" />
          <path d="M50,50 L72,11 A45,45 0 0,0 50,5 Z" fill="hsl(var(--primary))" opacity="0.04" />
        </g>

        {/* 节点间连接脉冲线 */}
        {links.map((link) => (
          <line
            key={link.id}
            x1={link.from.cx}
            y1={link.from.cy}
            x2={link.to.cx}
            y2={link.to.cy}
            stroke="hsl(var(--primary))"
            strokeWidth="0.3"
            strokeDasharray="4 2"
            style={{
              animation: 'radar-link-pulse 1.5s ease-out forwards',
              filter: 'drop-shadow(0 0 1px hsl(var(--primary)))',
            }}
          />
        ))}

        {nodes && nodes.length > 0 ? (
          nodes.map((node, i) => {
            const pos = computeBlipPosition(i);
            const isOnline = node.status === 'online';
            // 每个节点的呼吸相位偏移，避免同步呼吸
            const breathDelay = `${(i * 0.7) % 3}s`;
            return (
              <circle
                key={node.name}
                cx={pos.cx}
                cy={pos.cy}
                r="1.5"
                className={cn(isOnline ? 'fill-cyan-400' : 'fill-red-500 animate-pulse')}
                style={
                  isOnline
                    ? {
                        filter: 'drop-shadow(0 0 2px #22d3ee) drop-shadow(0 0 4px #22d3ee)',
                        animation: `radar-node-breathe 3s ${breathDelay} ease-in-out infinite`,
                      }
                    : {
                        filter: 'drop-shadow(0 0 2px #ef4444) drop-shadow(0 0 4px #ef4444)',
                      }
                }
              />
            );
          })
        ) : (
          <>
            <circle
              cx="60"
              cy="35"
              r="1.5"
              className="radar-blip-svg"
              style={{ animationDelay: '0.5s' }}
            />
            <circle
              cx="40"
              cy="65"
              r="1.5"
              className="radar-blip-svg"
              style={{ animationDelay: '2.1s' }}
            />
            <circle
              cx="25"
              cy="45"
              r="1.5"
              className="radar-blip-svg"
              style={{ animationDelay: '3.2s' }}
            />
          </>
        )}

        <circle
          cx="50"
          cy="50"
          r="1.5"
          className="fill-primary"
          style={{ filter: 'drop-shadow(0 0 3px hsl(var(--primary)))' }}
        />
      </svg>
    </div>
  );
}
