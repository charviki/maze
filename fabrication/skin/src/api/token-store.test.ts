import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
  clearTokens,
  getAccessToken,
  getAccessTokenExpiresAt,
  getRefreshToken,
  isAccessTokenExpired,
  isTokenAuthenticated,
  storeTokens,
  willAccessTokenExpireSoon,
} from '../index';

describe('token-store', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-09T12:00:00Z'));
    localStorage.clear();
  });

  afterEach(() => {
    vi.useRealTimers();
    localStorage.clear();
  });

  it('storeTokens 会同时保存 token 与 access token 绝对过期时间', () => {
    storeTokens({
      accessToken: 'access-1',
      refreshToken: 'refresh-1',
      expiresIn: 120,
    });

    expect(getAccessToken()).toBe('access-1');
    expect(getRefreshToken()).toBe('refresh-1');
    expect(getAccessTokenExpiresAt()).toBe(Date.now() + 120_000);
  });

  it('接近阈值时应视为需要预刷新', () => {
    storeTokens({
      accessToken: 'access-1',
      refreshToken: 'refresh-1',
      expiresIn: 45,
    });

    expect(willAccessTokenExpireSoon()).toBe(true);
    expect(isAccessTokenExpired()).toBe(false);
  });

  it('过期后应判定为已过期', () => {
    storeTokens({
      accessToken: 'access-1',
      refreshToken: 'refresh-1',
      expiresIn: 1,
    });

    vi.advanceTimersByTime(2_000);

    expect(isAccessTokenExpired()).toBe(true);
    expect(willAccessTokenExpireSoon()).toBe(true);
  });

  it('仅剩 refresh token 时仍视为存在可恢复会话', () => {
    localStorage.setItem('maze:refresh_token', 'refresh-only');

    expect(isTokenAuthenticated()).toBe(true);
  });

  it('clearTokens 会清理绝对过期时间', () => {
    storeTokens({
      accessToken: 'access-1',
      refreshToken: 'refresh-1',
      expiresIn: 60,
    });

    clearTokens();

    expect(getAccessToken()).toBeNull();
    expect(getRefreshToken()).toBeNull();
    expect(getAccessTokenExpiresAt()).toBeNull();
  });
});
