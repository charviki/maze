import { describe, it, expect, vi, beforeEach } from 'vitest';
import { formatRelativeTime, formatAbsoluteTime } from './time';

describe('formatRelativeTime', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-01-15T12:00:00Z'));
  });

  it('returns empty string for empty input', () => {
    expect(formatRelativeTime('')).toBe('');
  });

  it('returns original string for invalid date', () => {
    expect(formatRelativeTime('not-a-date')).toBe('not-a-date');
  });

  it('returns "刚刚" for less than 1 minute ago', () => {
    const d = new Date('2026-01-15T11:59:30Z').toISOString();
    expect(formatRelativeTime(d)).toBe('刚刚');
  });

  it('returns minutes ago', () => {
    const d = new Date('2026-01-15T11:30:00Z').toISOString();
    expect(formatRelativeTime(d)).toBe('30 分钟前');
  });

  it('returns hours ago', () => {
    const d = new Date('2026-01-15T10:00:00Z').toISOString();
    expect(formatRelativeTime(d)).toBe('2 小时前');
  });

  it('returns days ago', () => {
    const d = new Date('2026-01-13T12:00:00Z').toISOString();
    expect(formatRelativeTime(d)).toBe('2 天前');
  });

  it('returns formatted date for older than a week', () => {
    const d = new Date('2026-01-01T00:00:00Z').toISOString();
    expect(formatRelativeTime(d)).toBe('2026-01-01');
  });
});

describe('formatAbsoluteTime', () => {
  it('returns empty string for empty input', () => {
    expect(formatAbsoluteTime('')).toBe('');
  });

  it('returns original string for invalid date', () => {
    expect(formatAbsoluteTime('not-a-date')).toBe('not-a-date');
  });

  it('formats date and time correctly', () => {
    const d = new Date('2026-01-15T14:30:00Z');
    const expected = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
    expect(formatAbsoluteTime('2026-01-15T14:30:00Z')).toBe(expected);
  });
});
