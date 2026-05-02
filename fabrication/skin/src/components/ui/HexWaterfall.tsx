import { useEffect, useRef } from 'react';
import { cn } from '../../utils';
import { useAnimationSettings } from './AnimationSettings';

interface HexWaterfallProps {
  className?: string;
  opacity?: number;
  color?: string;
}

export function HexWaterfall({ className, opacity = 0.15, color }: HexWaterfallProps) {
  const { settings } = useAnimationSettings();
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    if (!settings.canvasBackground) return;

    const canvas = canvasRef.current;
    if (!canvas) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animationFrameId: number;
    // 跟踪 rAF 循环是否正在运行，避免重复启动
    let isRunning = false;
    let drops: number[] = [];
    let columns = 0;
    const fontSize = 14;

    const resize = () => {
      const dpr = window.devicePixelRatio || 1;
      const rect =
        canvas.parentElement?.getBoundingClientRect() || document.body.getBoundingClientRect();

      canvas.width = rect.width * dpr;
      canvas.height = rect.height * dpr;

      ctx.scale(dpr, dpr);
      canvas.style.width = `${rect.width}px`;
      canvas.style.height = `${rect.height}px`;

      columns = Math.floor(rect.width / fontSize);
      drops = new Array(columns).fill(0).map(() => Math.random() * -100);
    };

    window.addEventListener('resize', resize);
    resize();

    const chars = '0123456789ABCDEF'.split('');
    let lastDrawTime = 0;

    const startAnimation = () => {
      if (isRunning) return;
      isRunning = true;
      animationFrameId = requestAnimationFrame(draw);
    };

    const stopAnimation = () => {
      isRunning = false;
      cancelAnimationFrame(animationFrameId);
    };

    const draw = (timestamp: number) => {
      if (!isRunning) return;
      animationFrameId = requestAnimationFrame(draw);

      if (timestamp - lastDrawTime < 50) return; // 约 20 fps
      lastDrawTime = timestamp;

      const rect =
        canvas.parentElement?.getBoundingClientRect() || document.body.getBoundingClientRect();

      // 半透明黑色背景，形成尾迹效果
      ctx.fillStyle = 'rgba(0, 0, 0, 0.1)';
      ctx.fillRect(0, 0, rect.width, rect.height);

      // 获取主题颜色
      let primaryColor = color;
      if (!primaryColor) {
        const primaryVar = getComputedStyle(document.body).getPropertyValue('--primary').trim();
        primaryColor = primaryVar ? `hsl(${primaryVar})` : '#0f0';
      }

      ctx.font = `${fontSize}px monospace`;

      for (let i = 0; i < drops.length; i++) {
        const text = chars[Math.floor(Math.random() * chars.length)];
        const x = i * fontSize;
        const y = drops[i] * fontSize;

        // 随机高亮部分字符
        if (Math.random() > 0.98) {
          ctx.fillStyle = '#ffffff';
        } else {
          ctx.fillStyle = primaryColor;
        }

        ctx.fillText(text, x, y);

        // 随机重置，让瀑布流更自然
        if (y > rect.height && Math.random() > 0.975) {
          drops[i] = 0;
        }
        drops[i]++;
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
    };
  }, [color, settings.canvasBackground]);

  if (!settings.canvasBackground) {
    return null;
  }

  return (
    <canvas
      ref={canvasRef}
      className={cn('pointer-events-none absolute inset-0 z-0', className)}
      style={{ opacity }}
      aria-hidden="true"
    />
  );
}
