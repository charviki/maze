import { type ReactNode } from 'react';

export function highlightMatch(text: string, query: string): ReactNode {
  if (!query.trim()) return text;

  const escaped = query.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const regex = new RegExp(`(${escaped})`, 'gi');
  const parts = text.split(regex);

  if (parts.length === 1) return text;

  const testRegex = new RegExp(escaped, 'i');
  return parts.map((part, i) =>
    testRegex.test(part) ? (
      <mark key={i} className="bg-primary/30 text-foreground rounded px-0.5">
        {part}
      </mark>
    ) : (
      part
    ),
  );
}
