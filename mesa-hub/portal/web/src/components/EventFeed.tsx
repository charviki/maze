import { useState, useEffect, useRef } from 'react';
import { DecryptText } from '@maze/fabrication';
import { SYSTEM_EVENTS } from '../data/mock-data';

interface Event {
  time: string;
  message: string;
}

function formatTime(): string {
  const now = new Date();
  return `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`;
}

function randomEvent(): Event {
  return {
    time: formatTime(),
    message: SYSTEM_EVENTS[Math.floor(Math.random() * SYSTEM_EVENTS.length)],
  };
}

export function EventFeed() {
  const [events, setEvents] = useState<Event[]>(() => [randomEvent()]);
  const containerRef = useRef<HTMLDivElement>(null);
  // 用 ref 持有当前递归定时器 ID，确保每次递归都能被 cleanup 清除
  const timerRef = useRef<number | undefined>(undefined);

  useEffect(() => {
    const schedule = () => {
      const delay = 15000 + Math.random() * 15000; // 15-30s
      timerRef.current = window.setTimeout(() => {
        setEvents((prev) => {
          const next = [...prev, randomEvent()];
          return next.length > 20 ? next.slice(-20) : next;
        });
        schedule();
      }, delay);
    };
    schedule();
    return () => {
      if (timerRef.current !== undefined) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  // Auto-scroll to bottom
  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [events]);

  return (
    <div className="flex flex-col gap-1 px-3">
      <span className="text-[8px] font-mono tracking-[0.2em] text-primary/30 uppercase mb-1">
        EVENT LOG
      </span>
      <div
        ref={containerRef}
        className="flex flex-col gap-0.5 max-h-28 overflow-y-auto scrollbar-thin"
      >
        {events.map((evt, i) => (
          <div
            key={`${evt.time}-${i}`}
            className="text-[8px] font-mono text-primary/30 leading-relaxed animate-in fade-in duration-300"
          >
            <span className="text-primary/20">[{evt.time}]</span>{' '}
            {i === events.length - 1 ? (
              <DecryptText text={evt.message} speed={20} maxIterations={2} />
            ) : (
              evt.message
            )}
          </div>
        ))}
      </div>
    </div>
  );
}
