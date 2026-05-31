import { describe, expect, it, vi, beforeEach } from 'vitest'
import { clientApi, employeeApi } from './api'

describe('api', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
  })

  it('calls fetch with expected url and headers', async () => {
    const fetchMock = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(JSON.stringify([{ id: 1, name: 'Ivan' }]), { status: 200, statusText: 'OK' }),
    )

    const result = await employeeApi.getAll()
    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(fetchMock).toHaveBeenCalledWith(
      '/employees',
      expect.objectContaining({
        headers: { 'Content-Type': 'application/json' },
      }),
    )
    expect(result).toEqual([{ id: 1, name: 'Ivan' }])
  })

  it('throws error when response is not ok', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(
      new Response(null, { status: 500, statusText: 'Internal Server Error' }),
    )

    await expect(clientApi.getAll()).rejects.toThrow('500 Internal Server Error')
  })

  it('returns undefined for 204 responses', async () => {
    vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(null, { status: 204, statusText: 'No Content' }))
    await expect(employeeApi.delete(1)).resolves.toBeUndefined()
  })
})

