import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { appealApi, clientApi, employeeApi, subthemeApi, themeApi } from '../api'
import { useCrud } from '../hooks/useCrud'
import { usePolling } from '../hooks/usePolling'
import Modal from '../components/Modal'
import type { Appeal, Client, Employee, Subtheme, Theme } from '../types'
import {
  applyTextFieldValidity,
  validateAppealText,
} from '../validation'

type CreateForm = { clientId: number; themeId: number; subthemeId: number | null; text: string }
const emptyCreate: CreateForm = { clientId: 0, themeId: 0, subthemeId: null, text: '' }
type EditForm = Omit<Appeal, 'id'>

const POLL_MS = 3000

export default function AppealsPage() {
  const navigate = useNavigate()
  const { items, loading, error, create, update, remove, reload } =
    useCrud<Appeal, EditForm>(appealApi)

  // Периодическое обновление списка каждые 3 с
  usePolling(reload, POLL_MS)

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

  // Создание
  const [createModal, setCreateModal] = useState(false)
  const [createForm, setCreateForm] = useState<CreateForm>(emptyCreate)
  const [createError, setCreateError] = useState<string | null>(null)
  const openCreate = () => { setCreateForm(emptyCreate); setCreateError(null); setCreateModal(true) }
  const closeCreate = () => { setCreateModal(false); setCreateError(null) }
  const handleCreate = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    applyTextFieldValidity(e.currentTarget)
    if (!e.currentTarget.checkValidity()) {
      e.currentTarget.reportValidity()
      return
    }
    const validationError = validateAppealText(createForm.text, 'Текст обращения')
    if (validationError) {
      setCreateError(validationError)
      return
    }
    await create({
      ...createForm,
      text: createForm.text.trim(),
      employeeId: null,
      teamId: null,
      status: 'active',
    })
    closeCreate()
  }

  // Редактирование
  const [editModal, setEditModal] = useState(false)
  const [editing, setEditing] = useState<Appeal | null>(null)
  const [editError, setEditError] = useState<string | null>(null)
  const [editForm, setEditForm] = useState<EditForm>({
    clientId: 0, employeeId: null, teamId: null, themeId: 0, subthemeId: null, text: '', status: 'active',
  })
  const openEdit = (item: Appeal) => {
    setEditError(null)
    setEditing(item)
    setEditForm({
      clientId: item.clientId, employeeId: item.employeeId, teamId: item.teamId,
      themeId: item.themeId, subthemeId: item.subthemeId,
      text: item.text, status: item.status,
    })
    setEditModal(true)
  }
  const closeEdit = () => { setEditModal(false); setEditing(null); setEditError(null) }
  const handleEdit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    applyTextFieldValidity(e.currentTarget)
    if (!e.currentTarget.checkValidity()) {
      e.currentTarget.reportValidity()
      return
    }
    const validationError = validateAppealText(editForm.text, 'Текст обращения')
    if (validationError) {
      setEditError(validationError)
      return
    }
    if (editing) await update(editing.id, { ...editForm, text: editForm.text.trim() })
    closeEdit()
  }

  // Фильтр по клиенту
  const [filterClientId, setFilterClientId] = useState<number>(0)

  const setC = <K extends keyof CreateForm>(k: K, v: CreateForm[K]) =>
    setCreateForm(f => ({ ...f, [k]: v }))
  const setE = <K extends keyof EditForm>(k: K, v: EditForm[K]) =>
    setEditForm(f => ({ ...f, [k]: v }))

  const clientEmail = (id: number) =>
    clients.find(c => c.id === id)?.email ?? String(id)
  const employeeEmail = (id: number) =>
    employees.find(e => e.id === id)?.email ?? String(id)
  const themeName = (id: number) =>
    themes.find(t => t.id === id)?.name ?? String(id)

  return (
    <div style={page}>
      <div style={titleRow}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <h2 style={{ margin: 0 }}>Обращения</h2>
        </div>
        <button style={btnPrimary} onClick={openCreate}>+ Добавить</button>
      </div>

      {/* Панель фильтра */}
      <div style={filterBar}>
        <label style={filterLabel}>
          Фильтр по клиенту:
          <select
            style={filterSelect}
            value={filterClientId}
            onChange={e => setFilterClientId(Number(e.target.value))}
          >
            <option value={0}>Все клиенты</option>
            {clients.map(c => (
              <option key={c.id} value={c.id}>{c.email}</option>
            ))}
          </select>
        </label>
        {filterClientId !== 0 && (
          <button style={clearBtn} onClick={() => setFilterClientId(0)}>
            Сбросить
          </button>
        )}
      </div>

      {loading && <p>Загрузка...</p>}
      {error && <p style={{ color: 'red' }}>{error}</p>}

      <table style={table}>
        <thead>
          <tr>
            {['ID', 'Статус', 'Клиент', 'Сотрудник', 'Тема', 'Текст', ''].map(h =>
              <th key={h} style={th}>{h}</th>
            )}
          </tr>
        </thead>
        <tbody>
          {items.filter(item => filterClientId === 0 || item.clientId === filterClientId).map(item => (
            <tr key={item.id}>
              <td style={td}>{item.id}</td>
              <td style={td}>
                <span style={item.status === 'closed' ? closedBadge : activeBadge}>
                  {item.status === 'closed' ? 'закрыто' : 'активно'}
                </span>
              </td>
              <td style={td}>{clientEmail(item.clientId)}</td>
              <td style={td}>
                {item.employeeId
                  ? <span style={assignedBadge}>{employeeEmail(item.employeeId)}</span>
                  : <span style={pendingBadge}>не назначен</span>}
              </td>
              <td style={td}>{themeName(item.themeId)}</td>
              <td style={td} title={item.text}>
                {item.text.length > 35 ? item.text.slice(0, 35) + '...' : item.text}
              </td>
              <td style={td}>
                <button style={btnSm} onClick={() => navigate('/appeals/' + String(item.id))}>
                  Детали
                </button>
                <button style={btnSm} onClick={() => openEdit(item)}>Изменить</button>
                <button style={{ ...btnSm, ...btnDanger }} onClick={() => remove(item.id)}>
                  Удалить
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {createModal && (
        <Modal title="Новое обращение" onClose={closeCreate}>
          <form onSubmit={handleCreate} style={formGrid}>
            <label style={labelS}>Клиент (email)
              <select style={input} value={createForm.clientId}
                onChange={e => setC('clientId', Number(e.target.value))} required>
                <option value={0} disabled>- выберите клиента -</option>
                {clients.map(c => <option key={c.id} value={c.id}>{c.email}</option>)}
              </select>
            </label>
            <label style={labelS}>Тема
              <select style={input} value={createForm.themeId}
                onChange={e => setC('themeId', Number(e.target.value))} required>
                <option value={0} disabled>- выберите тему -</option>
                {themes.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
            </label>
            <label style={labelS}>Подтема
              <select style={input} value={createForm.subthemeId ?? 0}
                onChange={e => {
                  const v = Number(e.target.value)
                  setC('subthemeId', v === 0 ? null : v)
                }}>
                <option value={0}>- нет -</option>
                {subthemes.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </label>
            <label style={labelS}>Текст
              <textarea style={{ ...input, minHeight: 80, resize: 'vertical' }}
                value={createForm.text}
                onChange={e => setC('text', e.target.value)}
                data-text-field="true"
                required />
            </label>
            {createError && <p style={{ margin: 0, color: '#e53e3e' }}>{createError}</p>}
            <p style={hint}>Сотрудник будет назначен системой автоматически.</p>
            <button style={{ ...btnPrimary, marginTop: 4 }} type="submit">Сохранить</button>
          </form>
        </Modal>
      )}

      {editModal && editing && (
        <Modal title="Редактировать обращение" onClose={closeEdit}>
          <form onSubmit={handleEdit} style={formGrid}>
            <label style={labelS}>Клиент (email)
              <select style={input} value={editForm.clientId}
                onChange={e => setE('clientId', Number(e.target.value))} required>
                <option value={0} disabled>- выберите клиента -</option>
                {clients.map(c => <option key={c.id} value={c.id}>{c.email}</option>)}
              </select>
            </label>
            <label style={labelS}>Сотрудник (email)
              <select
                style={input}
                value={editForm.employeeId ?? 0}
                onChange={e => {
                  const v = Number(e.target.value)
                  setE('employeeId', v === 0 ? null : v)
                }}
              >
                <option value={0}>- не назначен -</option>
                {employees.map(emp =>
                  <option key={emp.id} value={emp.id}>{emp.email}</option>
                )}
              </select>
            </label>
            <label style={labelS}>Тема
              <select style={input} value={editForm.themeId}
                onChange={e => setE('themeId', Number(e.target.value))} required>
                <option value={0} disabled>- выберите тему -</option>
                {themes.map(t => <option key={t.id} value={t.id}>{t.name}</option>)}
              </select>
            </label>
            <label style={labelS}>Подтема
              <select style={input} value={editForm.subthemeId ?? 0}
                onChange={e => {
                  const v = Number(e.target.value)
                  setE('subthemeId', v === 0 ? null : v)
                }}>
                <option value={0}>- нет -</option>
                {subthemes.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </label>
            <label style={labelS}>Текст
              <textarea style={{ ...input, minHeight: 80, resize: 'vertical' }}
                value={editForm.text}
                onChange={e => setE('text', e.target.value)}
                data-text-field="true"
                required />
            </label>
            {editError && <p style={{ margin: 0, color: '#e53e3e' }}>{editError}</p>}
            <button style={{ ...btnPrimary, marginTop: 8 }} type="submit">Сохранить</button>
          </form>
        </Modal>
      )}
    </div>
  )
}

const page: React.CSSProperties = { padding: '0 4px' }
const titleRow: React.CSSProperties = {
  display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20,
}
const table: React.CSSProperties = {
  width: '100%', borderCollapse: 'collapse', background: '#fff',
  borderRadius: 8, overflow: 'hidden', boxShadow: '0 1px 6px rgba(0,0,0,.07)',
}
const th: React.CSSProperties = {
  textAlign: 'left', padding: '12px 16px', background: '#f5f6fa',
  fontWeight: 600, color: '#555', fontSize: 13,
}
const td: React.CSSProperties = {
  padding: '11px 16px', borderTop: '1px solid #f0f0f0', fontSize: 14,
}
const btnPrimary: React.CSSProperties = {
  background: '#4f6ef7', color: '#fff', border: 'none',
  borderRadius: 6, padding: '8px 18px', cursor: 'pointer', fontWeight: 600, fontSize: 14,
}
const btnSm: React.CSSProperties = {
  marginRight: 6, padding: '5px 10px', border: 'none',
  borderRadius: 5, cursor: 'pointer', fontSize: 12,
  background: '#e8eaff', color: '#4f6ef7', fontWeight: 500,
}
const btnDanger: React.CSSProperties = { background: '#fff0f0', color: '#e53e3e' }
const formGrid: React.CSSProperties = { display: 'flex', flexDirection: 'column', gap: 12 }
const labelS: React.CSSProperties = {
  display: 'flex', flexDirection: 'column', gap: 5, fontSize: 14, fontWeight: 500, color: '#444',
}
const input: React.CSSProperties = {
  padding: '8px 12px', border: '1px solid #ddd', borderRadius: 6, fontSize: 14, outline: 'none',
}
const hint: React.CSSProperties = { margin: 0, fontSize: 12, color: '#888', fontStyle: 'italic' }
const activeBadge: React.CSSProperties = {
  background: '#e6ffed', color: '#276749',
  borderRadius: 4, padding: '2px 8px', fontSize: 12, fontWeight: 600,
}
const closedBadge: React.CSSProperties = {
  background: '#f0f0f0', color: '#666',
  borderRadius: 4, padding: '2px 8px', fontSize: 12, fontWeight: 600,
}
const assignedBadge: React.CSSProperties = {
  background: '#e6ffed', color: '#276749',
  borderRadius: 4, padding: '2px 8px', fontSize: 13, fontWeight: 500,
}
const pendingBadge: React.CSSProperties = {
  background: '#fff8e1', color: '#92600a',
  borderRadius: 4, padding: '2px 8px', fontSize: 13, fontWeight: 500,
}
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
  fontSize: 14, outline: 'none', minWidth: 220, background: '#fff',
}
const clearBtn: React.CSSProperties = {
  padding: '7px 14px', border: '1px solid #e0e0e0', borderRadius: 6,
  cursor: 'pointer', fontSize: 13, background: '#f5f5f5', color: '#555',
}
