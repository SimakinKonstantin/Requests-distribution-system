import { renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { usePolling } from './usePolling'

describe('usePolling', () => {
  it('calls function by interval when enabled', () => {
    vi.useFakeTimers()
    const fn = vi.fn()
    renderHook(() => usePolling(fn, 1000, true))

    expect(fn).toHaveBeenCalledTimes(0)
    vi.advanceTimersByTime(999)
    expect(fn).toHaveBeenCalledTimes(0)
    vi.advanceTimersByTime(1)
    expect(fn).toHaveBeenCalledTimes(1)
    vi.advanceTimersByTime(2000)
    expect(fn).toHaveBeenCalledTimes(3)
    vi.useRealTimers()
  })

  it('does not call function when disabled', () => {
    vi.useFakeTimers()
    const fn = vi.fn()
    renderHook(() => usePolling(fn, 500, false))
    vi.advanceTimersByTime(2000)
    expect(fn).toHaveBeenCalledTimes(0)
    vi.useRealTimers()
  })
})

