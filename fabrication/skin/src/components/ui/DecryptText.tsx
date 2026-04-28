import { useState, useEffect, useRef } from 'react';
import { cn } from '../../utils';
import { useAnimationSettings } from './AnimationSettings';

interface DecryptTextProps {
  text: string;
  speed?: number;
  maxIterations?: number;
  characters?: string;
  className?: string;
  animateOnHover?: boolean;
}

export function DecryptText({
  text,
  speed = 30,
  maxIterations = 3,
  characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789@#$%&*!?',
  className = '',
  animateOnHover = false,
}: DecryptTextProps) {
  const { settings } = useAnimationSettings();
  const [displayText, setDisplayText] = useState(text);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const isAnimatingRef = useRef(false);

  const startAnimation = () => {
    if (isAnimatingRef.current) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
    }
    isAnimatingRef.current = true;
    let iteration = 0;

    intervalRef.current = setInterval(() => {
      setDisplayText(() => {
        return text
          .split('')
          .map((letter, index) => {
            if (letter === ' ') return letter;

            if (index < iteration) {
              return text[index];
            }

            return characters[Math.floor(Math.random() * characters.length)];
          })
          .join('');
      });

      if (iteration >= text.length) {
        if (intervalRef.current) clearInterval(intervalRef.current);
        intervalRef.current = null;
        isAnimatingRef.current = false;
      }

      iteration += 1 / maxIterations;
    }, speed);
  };

  useEffect(() => {
    if (!animateOnHover && settings.decryptText) {
      startAnimation();
    }
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      isAnimatingRef.current = false;
    };
  }, [text, animateOnHover, settings.decryptText]);

  return (
    <span 
      className={cn("font-mono tabular-nums inline-block", className)}
      onMouseEnter={animateOnHover ? startAnimation : undefined}
    >
      {displayText}
    </span>
  );
}
