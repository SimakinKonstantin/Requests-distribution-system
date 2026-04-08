import { useCallback, useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { appealApi, clientApi, employeeApi, subthemeApi, themeApi } from '../api'
import { usePolling } from '../hooks/usePolling'
import type { Appeal, Client, Employee, Subtheme, Theme } from '../types'

const POLL_MS = 3000

export default function AppealDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const appealId = Number(id)

  const [appeal, setAppeal] = useState<Appeal | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [closing, setClosing] = useState(false)

  // Справочники
  const [clients, setClients] = useState<Client[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [themes, setThemes] = useState<Theme[]>([])
  const [subthemes, setSubthemes] = useState<Subtheme[]>([])

  useEffect(() => {
    clientApi.getAll().then(setClients).catch(() => {})
    employeeApi.getAll().then(setEmployees).catch(() => {})
    themeApi.getAll().then(setThemes).catch(() => {})
    subthemeApi.getAll().then(setSubthemes).catch(() => {})
  }, [])

  // Загрузка конкретного обращения
  const fetchAppeal = useCallback(async () => {
    try {
      const a = await appealApi.getById(appealId)
      setAppeal(a)
      setError(null)
    } catch (e) {
      setError(String(e))
    } finally {
      setLoading(false)
    }
  }, [appealId])

  useEffect(() => { void fetchAppeal() }, [fetchAppeal])

  // Short polling — активен всегда, пока страница открыта
  usePolling(fetchAppeal, POLL_MS)

  // Закрытие обращения
  const handleClose = async () => {
    setClosing(true)
    try {
      const updated = await appealApi.close(appealId)
      setAppeal(updated)
    } catch (e) {
      alert(`Ошибка: ${e}`)
    } finally {
      setClosing(false)
    }
  }

  // Helpers
  const clientEmail   = (id: number) => clients.find(c => c.id === id)?.email   ?? `#${id}`
  const employeeEmail = (id: number) => employees.find(e => e.id === id)?.email ?? `#${id}`
  const themeName     = (id: number) => themes.find(t => t.id === id)?.name     ?? `#${id}`
  const subthemeName  = (id: number | null) => id == null ? '—' : (subthemes.find(s => s.id === id)?.name  ?? `#${id}`)

  if (loading) return <div style={page}><p>Загрузка...</p></div>
  if (error)   return <div style={page}><p style={{ color: 'red' }}>{error}</p></div>
  if (!appeal) return null

  const isClosed = appeal.status === 'closed'

  return (
    <div style={page}>
      {/* Шапка */}
      <div style={header}>
        <button style={backBtn} onClick={() => navigate('/appeals')}>
          ← Назад
        </button>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <h2 style={{ margin: 0 }}>Обращение #{appeal.id}</h2>
          <span style={isClosed ? closedBadge : activeBadge}>
            {isClosed ? 'закрыто' : 'активно'}
          </span>
          <span style={pollBadge}>&#x21BB; {POLL_MS / 1000} с</span>
        </div>
        {!isClosed && (
          <button style={closeBtn} onClick={handleClose} disabled={closing}>
            {closing ? 'Закрываем...' : 'Закрыть обращение'}
          </button>
        )}
      </div>

      {/* Карточка */}
      <div style={card}>
        <Row label="Клиент">{clientEmail(appeal.clientId)}</Row>

        <Row label="Сотрудник">
          {appeal.employeeId
            ? <span style={assignedBadge}>{employeeEmail(appeal.employeeId)}</span>
            : <span style={pendingBadge}>&#x23F3; ожидает назначения...</span>}
        </Row>

        <Row label="Тема">{themeName(appeal.themeId)}</Row>
        <Row label="Подтема">{subthemeName(appeal.subthemeId)}</Row>

        <div style={textSection}>
          <span style={rowLabel}>Текст</span>
          <p style={textBody}>{appeal.text}</p>
        </div>
      </div>

      {/* Подсказка про поллинг */}
      <p style={pollHint}>
        Страница автоматически обновляется каждые {POLL_MS / 1000} секунды.
        Как только сторонний сервис назначит сотрудника, изменение отобразится здесь.
      </p>
    </div>
  )
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div style={rowWrap}>
      <span style={rowLabel}>{label}</span>
      <span style={rowValue}>{children}</span>
    </div>
  )
}

// ── styles ───────────────────────────────────────────────────────────────────
const page: React.CSSProperties = { padding: '0 4px', maxWidth: 720 }
const header: React.CSSProperties = {
  display: 'flex', alignItems: 'center', gap: 16, marginBottom: 24, flexWrap: 'wrap',
}
const backBtn: React.CSSProperties = {
  background: 'none', border: '1px solid #dde', borderRadius: 6,
  padding: '6px 14px', cursor: 'pointer', fontSize: 13, color: '#555',
}
const closeBtn: React.CSSProperties = {
  marginLeft: 'auto', background: '#e53e3e', color: '#fff', border: 'none',
  borderRadius: 6, padding: '8px 18px', cursor: 'pointer', fontWeight: 600, fontSize: 14,
}
const card: React.CSSProperties = {
  background: '#fff', borderRadius: 10, padding: '24px 28px',
  boxShadow: '0 1px 8px rgba(0,0,0,.08)', display: 'flex', flexDirection: 'column', gap: 18,
}
const rowWrap: React.CSSProperties = { display: 'flex', alignItems: 'center', gap: 16 }
const rowLabel: React.CSSProperties = {
  width: 110, fontWeight: 600, color: '#555', fontSize: 13, flexShrink: 0,
}
const rowValue: React.CSSProperties = { fontSize: 15, color: '#222' }
const textSection: React.CSSProperties = { display: 'flex', flexDirection: 'column', gap: 6 }
const textBody: React.CSSProperties = {
  margin: 0, fontSize: 15, color: '#222', whiteSpace: 'pre-wrap',
  lineHeight: 1.6, background: '#f8f9ff', padding: '12px 16px', borderRadius: 6,
}
const activeBadge: React.CSSProperties = {
  background: '#e6ffed', color: '#276749', borderRadius: 6,
  padding: '3px 12px', fontSize: 13, fontWeight: 600,
}
const closedBadge: React.CSSProperties = {
  background: '#f0f0f0', color: '#666', borderRadius: 6,
  padding: '3px 12px', fontSize: 13, fontWeight: 600,
}
const assignedBadge: React.CSSProperties = {
  background: '#e6ffed', color: '#276749', borderRadius: 4,
  padding: '2px 10px', fontSize: 14, fontWeight: 500,
}
const pendingBadge: React.CSSProperties = {
  background: '#fff8e1', color: '#92600a', borderRadius: 4,
  padding: '2px 10px', fontSize: 14, fontWeight: 500,
}
const pollBadge: React.CSSProperties = {
  fontSize: 12, color: '#7c8cf8', background: '#eef0ff',
  borderRadius: 4, padding: '2px 8px', fontWeight: 500,
}
const pollHint: React.CSSProperties = {
  marginTop: 16, fontSize: 12, color: '#aaa', fontStyle: 'italic',
}
