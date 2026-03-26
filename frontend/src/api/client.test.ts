import { describe, it, expect, vi, beforeEach } from 'vitest'
import { searchBands, getBand, getAlbum, getSimilarAlbums } from './client'

// Mock global fetch
const mockFetch = vi.fn()
vi.stubGlobal('fetch', mockFetch)

beforeEach(() => {
  mockFetch.mockReset()
})

describe('searchBands', () => {
  it('fetches search results with encoded query', async () => {
    const results = [{ ma_id: 125, name: 'Metallica', genre: 'Thrash Metal', country: 'United States' }]
    mockFetch.mockResolvedValueOnce({
      ok: true,
      json: () => Promise.resolve(results),
    })

    const data = await searchBands('Metallica')
    expect(data).toEqual(results)
    expect(mockFetch).toHaveBeenCalledWith(
      '/api/bands/search?q=Metallica',
      expect.objectContaining({ headers: expect.objectContaining({ 'Content-Type': 'application/json' }) }),
    )
  })

  it('encodes special characters in query', async () => {
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) })
    await searchBands('AC/DC')
    expect(mockFetch.mock.calls[0][0]).toBe('/api/bands/search?q=AC%2FDC')
  })

  it('throws on HTTP error', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 500,
      statusText: 'Internal Server Error',
      json: () => Promise.resolve({ error: 'search failed' }),
    })

    await expect(searchBands('test')).rejects.toThrow('search failed')
  })

  it('throws with status text when error body is not JSON', async () => {
    mockFetch.mockResolvedValueOnce({
      ok: false,
      status: 502,
      statusText: 'Bad Gateway',
      json: () => Promise.reject(new Error('not json')),
    })

    await expect(searchBands('test')).rejects.toThrow('Bad Gateway')
  })
})

describe('getBand', () => {
  it('fetches band by ID', async () => {
    const band = { ma_id: 482, name: 'Opeth', genre: 'Progressive Metal' }
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(band) })

    const data = await getBand(482)
    expect(data).toEqual(band)
    expect(mockFetch.mock.calls[0][0]).toBe('/api/bands/482')
  })
})

describe('getAlbum', () => {
  it('fetches album by ID', async () => {
    const album = { album_id: 1234, name: 'Blackwater Park' }
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(album) })

    const data = await getAlbum(1234)
    expect(data).toEqual(album)
    expect(mockFetch.mock.calls[0][0]).toBe('/api/albums/1234')
  })
})

describe('getSimilarAlbums', () => {
  it('fetches similar albums by album ID', async () => {
    const similar = [
      { album_id: 100, name: 'Blackwater Park', band_name: 'Opeth', type: 'Full-length', year: '2001', cover_url: '', score: 0.95 },
      { album_id: 200, name: 'Still Life', band_name: 'Opeth', type: 'Full-length', year: '1999', cover_url: '', score: 0.88 },
    ]
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve(similar) })

    const data = await getSimilarAlbums(42)
    expect(data).toEqual(similar)
    expect(mockFetch.mock.calls[0][0]).toBe('/api/albums/42/similar')
  })

  it('returns empty array when no similar albums', async () => {
    mockFetch.mockResolvedValueOnce({ ok: true, json: () => Promise.resolve([]) })

    const data = await getSimilarAlbums(999)
    expect(data).toEqual([])
  })
})
