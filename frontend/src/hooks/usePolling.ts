import { useEffect, useRef } from 'react'

/**
 * usePolling — вызывает `fn` сразу при монтировании, затем каждые `intervalMs` мс.
 * Очищает таймер при размонтировании или изменении зависимостей.
 *
 * @param fn          Функция, которую нужно вызывать периодически (должна быть стабильной
 *                    ссылкой — оборачивайте в useCallback при необходимости)
 * @param intervalMs  Интервал в миллисекундах (по умолчанию 3000)
 * @param enabled     Флаг активации поллинга (по умолчанию true)
 */
export function usePolling(fn: () => void, intervalMs = 3000, enabled = true) {
  // Храним актуальную функцию в ref, чтобы не пересоздавать эффект при каждом рендере
  const fnRef = useRef(fn)
  useEffect(() => {
    fnRef.current = fn
  })

  useEffect(() => {
    if (!enabled) return

    // Немедленный первый вызов не нужен — useCrud уже делает начальную загрузку.
    // Поллинг начинается через intervalMs.
    const id = setInterval(() => fnRef.current(), intervalMs)
    return () => clearInterval(id)
  }, [intervalMs, enabled])
}
