import { useCallback, useEffect, useState } from 'react'

interface CrudApi<T, C> {
  getAll: () => Promise<T[]>
  create: (data: C) => Promise<T>
  update: (id: number, data: C) => Promise<T>
  delete: (id: number) => Promise<void>
}

export function useCrud<T extends { id: number }, C>(api: CrudApi<T, C>) {
  const [items, setItems] = useState<T[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const data = await api.getAll()
      setItems(data ?? [])
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }, [api])

  useEffect(() => { load() }, [load])

  const create = async (data: C) => {
    const item = await api.create(data)
    setItems(prev => [...prev, item])
  }

  const update = async (id: number, data: C) => {
    const item = await api.update(id, data)
    setItems(prev => prev.map(i => (i.id === id ? item : i)))
  }

  const remove = async (id: number) => {
    await api.delete(id)
    setItems(prev => prev.filter(i => i.id !== id))
  }

  return { items, loading, error, create, update, remove, reload: load }
}
