import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { useCrud } from './useCrud'

interface Item {
  id: number
  name: string
}

describe('useCrud', () => {
  it('loads items on mount', async () => {
    const api = {
      getAll: vi.fn().mockResolvedValue([{ id: 1, name: 'A' } satisfies Item]),
      create: vi.fn(),
      update: vi.fn(),
      delete: vi.fn(),
    }

    const { result } = renderHook(() => useCrud<Item, Omit<Item, 'id'>>(api))

    await waitFor(() => expect(result.current.loading).toBe(false))
    expect(api.getAll).toHaveBeenCalledTimes(1)
    expect(result.current.items).toEqual([{ id: 1, name: 'A' }])
    expect(result.current.error).toBeNull()
  })

  it('supports create update remove and reload', async () => {
    const api = {
      getAll: vi
        .fn()
        .mockResolvedValueOnce([{ id: 1, name: 'A' } satisfies Item])
        .mockResolvedValueOnce([{ id: 1, name: 'A*' } satisfies Item]),
      create: vi.fn().mockResolvedValue({ id: 2, name: 'B' } satisfies Item),
      update: vi.fn().mockResolvedValue({ id: 1, name: 'A+' } satisfies Item),
      delete: vi.fn().mockResolvedValue(undefined),
    }

    const { result } = renderHook(() => useCrud<Item, Omit<Item, 'id'>>(api))
    await waitFor(() => expect(result.current.loading).toBe(false))

    await act(async () => {
      await result.current.create({ name: 'B' })
    })
    expect(result.current.items).toEqual([
      { id: 1, name: 'A' },
      { id: 2, name: 'B' },
    ])

    await act(async () => {
      await result.current.update(1, { name: 'A+' })
    })
    expect(result.current.items).toEqual([
      { id: 1, name: 'A+' },
      { id: 2, name: 'B' },
    ])

    await act(async () => {
      await result.current.remove(2)
    })
    expect(result.current.items).toEqual([{ id: 1, name: 'A+' }])

    await act(async () => {
      await result.current.reload()
    })
    expect(result.current.items).toEqual([{ id: 1, name: 'A*' }])
  })
})

