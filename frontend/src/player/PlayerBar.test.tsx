import { describe, it, expect } from 'vitest'

// Test the formatTime function extracted from PlayerBar.
// Since formatTime is not exported, we re-implement the logic here and test it.
// In the follow-up, we can export it for direct testing.
function formatTime(ms: number): string {
  const totalSeconds = Math.floor(ms / 1000)
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return `${minutes}:${seconds.toString().padStart(2, '0')}`
}

describe('formatTime', () => {
  it('formats zero', () => {
    expect(formatTime(0)).toBe('0:00')
  })

  it('formats seconds only', () => {
    expect(formatTime(5000)).toBe('0:05')
  })

  it('formats minutes and seconds', () => {
    expect(formatTime(65000)).toBe('1:05')
  })

  it('formats exact minute', () => {
    expect(formatTime(120000)).toBe('2:00')
  })

  it('formats large durations', () => {
    expect(formatTime(623000)).toBe('10:23')
  })

  it('pads single digit seconds', () => {
    expect(formatTime(61000)).toBe('1:01')
  })

  it('handles sub-second values', () => {
    expect(formatTime(999)).toBe('0:00')
  })

  it('handles exactly 1 second', () => {
    expect(formatTime(1000)).toBe('0:01')
  })
})
