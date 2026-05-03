import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

// 读取 CSS 变量 --terminal-foreground 并转换为 hsla 格式，
// 确保 L 分量不低于 50% 以保证前景色在暗背景上可读
export function getPrimaryColor(): string {
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

// 基于预缓存的 hsla 基础色字符串，快速拼接透明度值，
// 避免每帧重复解析模板字符串
export function colorTpl(base: string, alpha: number): string {
  return base.replace(')', `, ${alpha})`);
}

// 生成四角切角 clip-path polygon 字符串，cornerSize 控制切角大小（单位 px）
export function clipPath(cornerSize: number): string {
  const c = cornerSize;
  return `polygon(${c}px 0, calc(100% - ${c}px) 0, 100% ${c}px, 100% calc(100% - ${c}px), calc(100% - ${c}px) 100%, ${c}px 100%, 0 calc(100% - ${c}px), 0 ${c}px)`;
}

// 生成仅右上和左下切角的 clip-path polygon 字符串，
// 与 clipPath 不同，左上和右下保持直角
export function clipPathHalf(cornerSize: number): string {
  const c = cornerSize;
  return `polygon(0 0, calc(100% - ${c}px) 0, 100% ${c}px, 100% 100%, ${c}px 100%, 0 calc(100% - ${c}px))`;
}
