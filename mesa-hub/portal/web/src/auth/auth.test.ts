import { describe, it, expect, beforeEach } from 'vitest'
import { login, logout, isAuthenticated, getCurrentUser } from './auth'

const STORAGE_KEY = 'maze:auth'

describe('auth', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  describe('login', () => {
    it('accepts valid credentials', () => {
      expect(login('admin', 'admin')).toBe(true)
    })

    it('rejects invalid username', () => {
      expect(login('wrong', 'admin')).toBe(false)
    })

    it('rejects invalid password', () => {
      expect(login('admin', 'wrong')).toBe(false)
    })

    it('rejects empty credentials', () => {
      expect(login('', '')).toBe(false)
    })

    it('persists session to localStorage on success', () => {
      login('admin', 'admin')
      const raw = localStorage.getItem(STORAGE_KEY)
      expect(raw).toBeTruthy()
      const session = JSON.parse(raw!)
      expect(session.user).toBe('admin')
      expect(session.loginAt).toBeTypeOf('number')
    })

    it('does not persist on failure', () => {
      login('wrong', 'wrong')
      expect(localStorage.getItem(STORAGE_KEY)).toBeNull()
    })
  })

  describe('logout', () => {
    it('clears localStorage', () => {
      login('admin', 'admin')
      expect(isAuthenticated()).toBe(true)
      logout()
      expect(isAuthenticated()).toBe(false)
    })
  })

  describe('isAuthenticated', () => {
    it('returns false when not logged in', () => {
      expect(isAuthenticated()).toBe(false)
    })

    it('returns true after login', () => {
      login('admin', 'admin')
      expect(isAuthenticated()).toBe(true)
    })

    it('returns true with valid stored session', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ user: 'admin', loginAt: Date.now() }))
      expect(isAuthenticated()).toBe(true)
    })

    it('returns false with invalid JSON', () => {
      localStorage.setItem(STORAGE_KEY, 'not-json')
      expect(isAuthenticated()).toBe(false)
    })

    it('returns false with empty user', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ user: '', loginAt: Date.now() }))
      expect(isAuthenticated()).toBe(false)
    })
  })

  describe('getCurrentUser', () => {
    it('returns null when not logged in', () => {
      expect(getCurrentUser()).toBeNull()
    })

    it('returns username after login', () => {
      login('admin', 'admin')
      expect(getCurrentUser()).toBe('admin')
    })

    it('returns stored username from valid session', () => {
      localStorage.setItem(STORAGE_KEY, JSON.stringify({ user: 'admin', loginAt: Date.now() }))
      expect(getCurrentUser()).toBe('admin')
    })

    it('returns null with invalid JSON', () => {
      localStorage.setItem(STORAGE_KEY, 'not-json')
      expect(getCurrentUser()).toBeNull()
    })
  })
})
