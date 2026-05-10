import { useState, useEffect } from 'react';

export function useClock(intervalMs = 1000): string {
  const [clock, setClock] = useState(() => new Date().toLocaleTimeString());

  useEffect(() => {
    const timer = setInterval(() => setClock(new Date().toLocaleTimeString()), intervalMs);
    return () => clearInterval(timer);
  }, [intervalMs]);

  return clock;
}
