import { useState } from 'react'
import { slotApi } from '../api'
import { useCrud } from '../hooks/useCrud'
import Modal from '../components/Modal'
import type { Slot } from '../types'

type Form = Omit<Slot, 'id'>
const empty: Form = { employeeId: 0, appealId: 0 }

export default function SlotsPage() {
  const { items, loading, error, create, update, remove } = useCrud<Slot, Form>(slotApi)
  const [modal, setModal] = useState<'create' | 'edit' | null>(null)
  const [editing, setEditing] = useState<Slot | null>(null)
  const [form, setForm] = useState<Form>(empty)

  const openCreate = () => { setForm(empty); setModal('create') }
  const openEdit = (item: Slot) => {
    setEditing(item)
    setForm({ employeeId: item.employeeId, appealId: item.appealId })
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
        <h2 style={{ margin: 0 }}>Слоты</h2>
        <button style={btnPrimary} onClick={openCreate}>+ Добавить</button>
      </div>

      {loading && <p>Загрузка…</p>}
      {error && <p style={{ color: 'red' }}>{error}</p>}

      <table style={table}>
        <thead>
          <tr>{['ID', 'ID сотрудника', 'ID обращения', ''].map(h => <th key={h} style={th}>{h}</th>)}</tr>
        </thead>
        <tbody>
          {items.map(item => (
            <tr key={item.id}>
              <td style={td}>{item.id}</td>
              <td style={td}>{item.employeeId}</td>
              <td style={td}>{item.appealId}</td>
              <td style={td}>
                <button style={btnSm} onClick={() => openEdit(item)}>Изменить</button>
                <button style={{ ...btnSm, ...btnDanger }} onClick={() => remove(item.id)}>Удалить</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {modal && (
        <Modal title={modal === 'create' ? 'Новый слот' : 'Редактировать слот'} onClose={close}>
          <form onSubmit={handleSubmit} style={formGrid}>
            <label style={label}>ID сотрудника
              <input style={input} type="number" value={form.employeeId}
                onChange={e => set('employeeId', Number(e.target.value))} required />
            </label>
            <label style={label}>ID обращения
              <input style={input} type="number" value={form.appealId}
                onChange={e => set('appealId', Number(e.target.value))} required />
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
