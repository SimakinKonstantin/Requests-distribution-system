import { useEffect, useState } from 'react'
import { employeeApi, slotApi } from '../api'
import { useCrud } from '../hooks/useCrud'
import { usePolling } from '../hooks/usePolling'
import type { Employee, Slot } from '../types'

const POLL_MS = 3000

export default function SlotsPage() {
  const { items, loading, error, reload } = useCrud<Slot, Omit<Slot, 'id'>>(slotApi)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [filterEmployeeId, setFilterEmployeeId] = useState<number>(0)

  usePolling(reload, POLL_MS)

  useEffect(() => {
    employeeApi.getAll().then(setEmployees).catch(() => setEmployees([]))
  }, [])

  const employeeEmail = (employeeId: number) =>
    employees.find(emp => emp.id === employeeId)?.email ?? `ID ${employeeId}`

  const filteredItems = items.filter(item =>
    filterEmployeeId === 0 || item.employeeId === filterEmployeeId,
  )

  return (
    <div style={page}>
      <h2 style={{ margin: '0 0 20px' }}>Ячейки</h2>

      {loading && <p>Загрузка…</p>}
      {error && <p style={{ color: 'red' }}>{error}</p>}

      <div style={filterBar}>
        <label style={filterLabel}>
          Фильтр по сотруднику (email):
          <select
            style={filterSelect}
            value={filterEmployeeId}
            onChange={e => setFilterEmployeeId(Number(e.target.value))}
          >
            <option value={0}>Все сотрудники</option>
            {employees.map(emp => (
              <option key={emp.id} value={emp.id}>{emp.email}</option>
            ))}
          </select>
        </label>
      </div>

      <table style={table}>
        <thead>
          <tr>{['ID', 'Email сотрудника', 'ID обращения', 'Будет удален'].map(h => <th key={h} style={th}>{h}</th>)}</tr>
        </thead>
        <tbody>
          {filteredItems.map(item => (
            <tr key={item.id}>
              <td style={td}>{item.id}</td>
              <td style={td}>{employeeEmail(item.employeeId)}</td>
              <td style={td}>{item.appealId ?? 'не назначен'}</td>
              <td style={td}>{item.needToRemove ? 'Да' : 'Нет'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

const page: React.CSSProperties = { padding: '0 4px' }
const table: React.CSSProperties = { width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden', boxShadow: '0 1px 6px rgba(0,0,0,.07)' }
const th: React.CSSProperties = { textAlign: 'left', padding: '12px 16px', background: '#f5f6fa', fontWeight: 600, color: '#555', fontSize: 13 }
const td: React.CSSProperties = { padding: '11px 16px', borderTop: '1px solid #f0f0f0', fontSize: 14 }
const filterBar: React.CSSProperties = {
  display: 'flex', alignItems: 'center', gap: 12, marginBottom: 16,
  background: '#fff', padding: '12px 16px', borderRadius: 8,
  boxShadow: '0 1px 4px rgba(0,0,0,.05)',
}
const filterLabel: React.CSSProperties = {
  display: 'flex', alignItems: 'center', gap: 10, fontSize: 14, fontWeight: 500, color: '#444',
}
const filterSelect: React.CSSProperties = {
  padding: '7px 12px', border: '1px solid #ddd', borderRadius: 6,
  fontSize: 14, outline: 'none', minWidth: 240, background: '#fff',
}
