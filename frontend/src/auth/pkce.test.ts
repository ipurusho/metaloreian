import { describe, it, expect, vi, beforeEach } from 'vitest'
import { getCodeVerifier, clearCodeVerifier, getRedirectUri } from './pkce'

// Mock sessionStorage
const mockSessionStorage: Record<string, string> = {}
vi.stubGlobal('sessionStorage', {
  getItem: (key: string) => mockSessionStorage[key] ?? null,
  setItem: (key: string, value: string) => { mockSessionStorage[key] = value },
  removeItem: (key: string) => { delete mockSessionStorage[key] },
})

beforeEach(() => {
  Object.keys(mockSessionStorage).forEach(key => delete mockSessionStorage[key])
})

describe('getCodeVerifier', () => {
  it('returns null when no verifier stored', () => {
    expect(getCodeVerifier()).toBeNull()
  })

  it('returns stored verifier', () => {
    mockSessionStorage['code_verifier'] = 'test-verifier'
    expect(getCodeVerifier()).toBe('test-verifier')
  })
})

describe('clearCodeVerifier', () => {
  it('removes verifier from storage', () => {
    mockSessionStorage['code_verifier'] = 'test-verifier'
    clearCodeVerifier()
    expect(getCodeVerifier()).toBeNull()
  })
})

describe('getRedirectUri', () => {
  it('returns a string', () => {
    const uri = getRedirectUri()
    expect(typeof uri).toBe('string')
    expect(uri.length).toBeGreaterThan(0)
  })
})
