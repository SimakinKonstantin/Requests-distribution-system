import { useState } from 'react'
import { appealApi } from '../api'
import { useCrud } from '../hooks/useCrud'
import Modal from '../components/Modal'
import type { Appeal } from '../types'

type Form = Omit<Appeal, 'id'>
const empty: Form = { clientId: 0, employeeId: 0, themeId: 0, subthemeId: 0, text: '' }

export default function AppealsPage() {
  const { items, loading, error, create, update, remove } = useCrud<Appeal, Form>(appealApi)
  const [modal, setModal] = useState<'create' | 'edit' | null>(null)
  const [editing, setEditing] = useState<Appeal | null>(null)
  const [form, setForm] = useState<Form>(empty)

  const openCreate = () => { setForm(empty); setModal('create') }
  const openEdit = (item: Appeal) => {
    setEditing(item)
    setForm({ clientId: item.clientId, employeeId: item.employeeId, themeId: item.themeId, subthemeId: item.subthemeId, text: item.text })
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
        <h2 style={{ margin: 0 }}>Обращения</h2>
        <button style={btnPrimary} onClick={openCreate}>+ Добавить</button>
      </div>

      {loading && <p>Загрузка…</p>}
      {error && <p style={{ color: 'red' }}>{error}</p>}

      <table style={table}>
        <thead>
          <tr>{['ID', 'ID клиента', 'ID сотрудника', 'ID темы', 'ID подтемы', 'Текст', ''].map(h => <th key={h} style={th}>{h}</th>)}</tr>
        </thead>
        <tbody>
          {items.map(item => (
            <tr key={item.id}>
              <td style={td}>{item.id}</td>
              <td style={td}>{item.clientId}</td>
              <td style={td}>{item.employeeId}</td>
              <td style={td}>{item.themeId}</td>
              <td style={td}>{item.subthemeId}</td>
              <td style={td} title={item.text}>
                {item.text.length > 40 ? item.text.slice(0, 40) + '…' : item.text}
              </td>
              <td style={td}>
                <button style={btnSm} onClick={() => openEdit(item)}>Изменить</button>
                <button style={{ ...btnSm, ...btnDanger }} onClick={() => remove(item.id)}>Удалить</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {modal && (
        <Modal title={modal === 'create' ? 'Новое обращение' : 'Редактировать обращение'} onClose={close}>
          <form onSubmit={handleSubmit} style={formGrid}>
            <label style={label}>ID клиента
              <input style={input} type="number" value={form.clientId}
                onChange={e => set('clientId', Number(e.target.value))} required />
            </label>
            <label style={label}>ID сотрудника
              <input style={input} type="number" value={form.employeeId}
                onChange={e => set('employeeId', Number(e.target.value))} />
            </label>
            <label style={label}>ID темы
              <input style={input} type="number" value={form.themeId}
                onChange={e => set('themeId', Number(e.target.value))} required />
            </label>
            <label style={label}>ID подтемы
              <input style={input} type="number" value={form.subthemeId}
                onChange={e => set('subthemeId', Number(e.target.value))} required />
            </label>
            <label style={label}>Текст
              <textarea style={{ ...input, minHeight: 80, resize: 'vertical' }} value={form.text}
                onChange={e => set('text', e.target.value)} required />
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
