import { useEffect, useRef } from 'react'

// Вызывает fn периодически с заданным интервалом.
// Очищает таймер при размонтировании или изменении зависимостей.
export function usePolling(fn: () => void, intervalMs = 3000, enabled = true) {
  // Храним актуальную функцию в ref, чтобы не пересоздавать эффект при каждом рендере
  const fnRef = useRef(fn)
  useEffect(() => {
    fnRef.current = fn
  })

  useEffect(() => {
    if (!enabled) return

    // Немедленный первый вызов не нужен — начальную загрузку уже выполняет useCrud.
    // Поллинг начинается с заданного интервала.
    const id = setInterval(() => fnRef.current(), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs, enabled])
}
