import { useState, useEffect } from 'react';
import { DecryptText } from '@maze/fabrication';
import { WESTWORLD_QUOTES, MOCK_NODES, MOCK_HOSTS, BUILD_VERSION } from '../data/mock-data';

const onlineNodes = MOCK_NODES.filter((n) => n.status === 'online').length;

export function StatusBar() {
  const [clock, setClock] = useState(new Date().toLocaleTimeString());
  const [quoteIndex, setQuoteIndex] = useState(0);

  useEffect(() => {
    const timer = setInterval(() => {
      setClock(new Date().toLocaleTimeString());
    }, 1000);
    return () => {
      clearInterval(timer);
    };
  }, []);

  useEffect(() => {
    const timer = setInterval(() => {
      setQuoteIndex((i) => (i + 1) % WESTWORLD_QUOTES.length);
    }, 10000);
    return () => {
      clearInterval(timer);
    };
  }, []);

  return (
    <div className="flex flex-col border-t border-primary/10 bg-card/40 backdrop-blur-sm">
      <div className="h-8 flex items-center px-4 text-[10px] font-mono tracking-widest text-primary/50 gap-3">
        <span>SYS_CLOCK: {clock}</span>
        <span className="text-primary/20">|</span>
        <span>
          NODES: {onlineNodes}/{MOCK_NODES.length}
        </span>
        <span className="text-primary/20">|</span>
        <span>HOSTS: {MOCK_HOSTS.total}</span>
        <span className="text-primary/20">|</span>
        <span>{BUILD_VERSION}</span>
      </div>

      <div className="h-6 flex items-center px-4 text-[9px] font-mono tracking-wider text-primary/30 italic overflow-hidden">
        <span key={quoteIndex} className="animate-in fade-in duration-500">
          "<DecryptText text={WESTWORLD_QUOTES[quoteIndex]} speed={25} maxIterations={2} />"
        </span>
      </div>
    </div>
  );
}
