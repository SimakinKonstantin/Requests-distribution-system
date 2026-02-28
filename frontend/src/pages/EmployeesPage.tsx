import { useState } from 'react'
import { employeeApi } from '../api'
import { useCrud } from '../hooks/useCrud'
import Modal from '../components/Modal'
import type { Employee } from '../types'

type Form = Omit<Employee, 'id'>
const empty: Form = { name: '', surname: '', limit: 0, teamId: 0 }

export default function EmployeesPage() {
  const { items, loading, error, create, update, remove } = useCrud<Employee, Form>(employeeApi)
  const [modal, setModal] = useState<'create' | 'edit' | null>(null)
  const [editing, setEditing] = useState<Employee | null>(null)
  const [form, setForm] = useState<Form>(empty)

  const openCreate = () => { setForm(empty); setModal('create') }
  const openEdit = (item: Employee) => {
    setEditing(item)
    setForm({ name: item.name, surname: item.surname, limit: item.limit, teamId: item.teamId })
    setModal('edit')
  }
  const close = () => { setModal(null); setEditing(null) }

  const set = <K extends keyof Form>(k: K, v: Form[K]) => setForm(f => ({ ...f, [k]: v }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (modal === 'create') await create(form)
    else if (editing) await update(editing.id, form)
    close()
  }

  return (
    <div style={page}>
      <div style={titleRow}>
        <h2 style={{ margin: 0 }}>Сотрудники</h2>
        <button style={btnPrimary} onClick={openCreate}>+ Добавить</button>
      </div>

      {loading && <p>Загрузка…</p>}
      {error && <p style={{ color: 'red' }}>{error}</p>}

      <table style={table}>
        <thead>
          <tr>{['ID', 'Имя', 'Фамилия', 'Лимит', 'ID команды', ''].map(h => <th key={h} style={th}>{h}</th>)}</tr>
        </thead>
        <tbody>
          {items.map(item => (
            <tr key={item.id}>
              <td style={td}>{item.id}</td>
              <td style={td}>{item.name}</td>
              <td style={td}>{item.surname}</td>
              <td style={td}>{item.limit}</td>
              <td style={td}>{item.teamId}</td>
              <td style={td}>
                <button style={btnSm} onClick={() => openEdit(item)}>Изменить</button>
                <button style={{ ...btnSm, ...btnDanger }} onClick={() => remove(item.id)}>Удалить</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {modal && (
        <Modal title={modal === 'create' ? 'Новый сотрудник' : 'Редактировать сотрудника'} onClose={close}>
          <form onSubmit={handleSubmit} style={formGrid}>
            <label style={label}>Имя
              <input style={input} value={form.name}
                onChange={e => set('name', e.target.value)} required />
            </label>
            <label style={label}>Фамилия
              <input style={input} value={form.surname}
                onChange={e => set('surname', e.target.value)} required />
            </label>
            <label style={label}>Лимит
              <input style={input} type="number" value={form.limit}
                onChange={e => set('limit', Number(e.target.value))} required />
            </label>
            <label style={label}>ID команды
              <input style={input} type="number" value={form.teamId}
                onChange={e => set('teamId', Number(e.target.value))} required />
            </label>
            <button style={{ ...btnPrimary, marginTop: 8 }} type="submit">Сохранить</button>
          </form>
        </Modal>
      )}
    </div>
  )
}

const page: React.CSSProperties = { padding: '0 4px' }
const titleRow: React.CSSProperties = { display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }
const table: React.CSSProperties = { width: '100%', borderCollapse: 'collapse', background: '#fff', borderRadius: 8, overflow: 'hidden', boxShadow: '0 1px 6px rgba(0,0,0,.07)' }
const th: React.CSSProperties = { textAlign: 'left', padding: '12px 16px', background: '#f5f6fa', fontWeight: 600, color: '#555', fontSize: 13 }
const td: React.CSSProperties = { padding: '11px 16px', borderTop: '1px solid #f0f0f0', fontSize: 14 }
const btnPrimary: React.CSSProperties = { background: '#4f6ef7', color: '#fff', border: 'none', borderRadius: 6, padding: '8px 18px', cursor: 'pointer', fontWeight: 600, fontSize: 14 }
const btnSm: React.CSSProperties = { marginRight: 6, padding: '5px 12px', border: 'none', borderRadius: 5, cursor: 'pointer', fontSize: 13, background: '#e8eaff', color: '#4f6ef7', fontWeight: 500 }
const btnDanger: React.CSSProperties = { background: '#fff0f0', color: '#e53e3e' }
const formGrid: React.CSSProperties = { display: 'flex', flexDirection: 'column', gap: 12 }
const label: React.CSSProperties = { display: 'flex', flexDirection: 'column', gap: 5, fontSize: 14, fontWeight: 500, color: '#444' }
const input: React.CSSProperties = { padding: '8px 12px', border: '1px solid #ddd', borderRadius: 6, fontSize: 14, outline: 'none' }
