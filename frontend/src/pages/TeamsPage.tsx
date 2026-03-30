import { useState } from 'react'
import { teamApi, themeApi, subthemeApi } from '../api'
import { useCrud } from '../hooks/useCrud'
import Modal from '../components/Modal'
import type { Team, TeamThemeSubtheme, Theme, Subtheme } from '../types'

type Form = Omit<Team, 'id'>
const empty: Form = { name: '', themeSubtheme: [] }

function normalizeRows(team: Team): TeamThemeSubtheme[] {
  const rows = (team as any).themeSubtheme ?? (team as any).ThemeSubtheme
  if (!rows || !Array.isArray(rows)) return []
  return rows
}

export default function TeamsPage() {
  const { items, loading, error, create, update, remove } = useCrud<Team, Form>(teamApi)
  const [modal, setModal] = useState<'create' | 'edit' | null>(null)
  const [editing, setEditing] = useState<Team | null>(null)
  const [form, setForm] = useState<Form>(empty)
  const [row, setRow] = useState<TeamThemeSubtheme>({ theme_id: 0, subtheme_id: 0, for_vip: false })
  const [themes, setThemes] = useState<Theme[]>([])
  const [subthemes, setSubthemes] = useState<Subtheme[]>([])

  // load dictionaries
  useState(() => {
    themeApi.getAll().then(setThemes).catch(() => setThemes([]))
    subthemeApi.getAll().then(setSubthemes).catch(() => setSubthemes([]))
  })

  const openCreate = () => {
    setForm(empty)
    setModal('create')
  }

  const openEdit = (item: Team) => {
    setEditing(item)
    setForm({ name: item.name, themeSubtheme: normalizeRows(item) })
    setModal('edit')
  }

  const close = () => {
    setModal(null)
    setEditing(null)
  }

  const addRow = () => {
    if (!row.theme_id) return
    setForm(f => ({ ...f, themeSubtheme: [...(f.themeSubtheme ?? []), row] }))
    setRow({ theme_id: 0, subtheme_id: 0, for_vip: false })
  }

  const removeRow = (idx: number) =>
    setForm(f => ({ ...f, themeSubtheme: (f.themeSubtheme ?? []).filter((_, i) => i !== idx) }))

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (modal === 'create') await create(form)
    else if (editing) await update(editing.id, form)
    close()
  }

  return (
    <div style={stylePage}>
      <div style={styleTitleRow}>
        <h2 style={{ margin: 0 }}>Команды</h2>
        <button style={styleBtnPrimary} onClick={openCreate}>+ Добавить</button>
      </div>

      {loading && <p>Загрузка…</p>}
      {error && <p style={{ color: 'red' }}>{error}</p>}

      <table style={styleTable}>
        <thead>
          <tr>
            {['ID', 'Название', 'Темы / Подтемы', ''].map(h => (
              <th key={h} style={styleTh}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {items.map(item => (
            <tr key={item.id}>
              <td style={styleTd}>{item.id}</td>
              <td style={styleTd}>{item.name}</td>
              <td style={styleTd}>
                {normalizeRows(item).length === 0 ? (
                  <span style={{ color: '#999' }}>—</span>
                ) : (
                  <ul style={{ margin: 0, paddingLeft: 18 }}>
                    {normalizeRows(item).map((r, i) => (
                      <li key={i}>
                        theme: {r.theme_id} / subtheme: {r.subtheme_id}
                        {r.for_vip ? ' (VIP)' : ''}
                      </li>
                    ))}
                  </ul>
                )}
              </td>
              <td style={styleTd}>
                <button style={styleBtnSm} onClick={() => openEdit(item)}>Изменить</button>
                <button style={{ ...styleBtnSm, ...styleBtnDanger }} onClick={() => remove(item.id)}>Удалить</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {modal && (
        <Modal
          title={modal === 'create' ? 'Новая команда' : 'Редактировать команду'}
          onClose={close}
        >
          <form onSubmit={handleSubmit} style={styleFormGrid}>
            <label style={styleLabel}>
              Название
              <input
                style={styleInput}
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                required
              />
            </label>

            <div style={{ fontWeight: 600, color: '#444' }}>Темы</div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 6 }}>
              {themes.map(t => (
                <label key={t.id} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <input
                    type="radio"
                    name="theme"
                    checked={row.theme_id === t.id}
                    onChange={() => setRow(r => ({ ...r, theme_id: t.id }))}
                  />
                  <span>({t.id}) {t.name}</span>
                </label>
              ))}
            </div>

            <div style={{ fontWeight: 600, color: '#444', marginTop: 8 }}>Подтемы</div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 6 }}>
              <label key="none" style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                <input
                  type="radio"
                  name="subtheme"
                  checked={row.subtheme_id === 0}
                  onChange={() => setRow(r => ({ ...r, subtheme_id: 0 }))}
                />
                <span>— нет —</span>
              </label>
              {subthemes.map(s => (
                <label key={s.id} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
                  <input
                    type="radio"
                    name="subtheme"
                    checked={row.subtheme_id === s.id}
                    onChange={() => setRow(r => ({ ...r, subtheme_id: s.id }))}
                  />
                  <span>({s.id}) {s.name}</span>
                </label>
              ))}
            </div>

            <label style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 8 }}>
              <input
                type="checkbox"
                checked={row.for_vip}
                onChange={e => setRow(r => ({ ...r, for_vip: e.target.checked }))}
              />
              Привязка для VIP
            </label>
            <button
              type="button"
              style={{ ...styleBtnSm, marginTop: 8 }}
              onClick={addRow}
              disabled={!row.theme_id}
            >
              Добавить привязку
            </button>

            {(form.themeSubtheme ?? []).length > 0 && (
              <ul style={{ marginTop: 8, paddingLeft: 18 }}>
                {(form.themeSubtheme ?? []).map((r, i) => (
                  <li key={i}>
                    theme: {r.theme_id} / subtheme: {r.subtheme_id}
                    {r.for_vip ? ' (VIP)' : ''}
                    <button
                      type="button"
                      style={{ ...styleBtnSm, marginLeft: 8 }}
                      onClick={() => removeRow(i)}
                    >
                      Удалить
                    </button>
                  </li>
                ))}
              </ul>
            )}

            <button style={{ ...styleBtnPrimary, marginTop: 12 }} type="submit">Сохранить</button>
          </form>
        </Modal>
      )}
    </div>
  )
}

const stylePage: React.CSSProperties = { padding: '0 4px' }

const styleTitleRow: React.CSSProperties = {
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  marginBottom: 20,
}

const styleTable: React.CSSProperties = {
  width: '100%',
  borderCollapse: 'collapse',
  background: '#fff',
  borderRadius: 8,
  overflow: 'hidden',
  boxShadow: '0 1px 6px rgba(0,0,0,.07)',
}

const styleTh: React.CSSProperties = {
  textAlign: 'left',
  padding: '12px 16px',
  background: '#f5f6fa',
  fontWeight: 600,
  color: '#555',
  fontSize: 13,
}

const styleTd: React.CSSProperties = {
  padding: '11px 16px',
  borderTop: '1px solid #f0f0f0',
  fontSize: 14,
}

const styleBtnPrimary: React.CSSProperties = {
  background: '#4f6ef7',
  color: '#fff',
  border: 'none',
  borderRadius: 6,
  padding: '8px 18px',
  cursor: 'pointer',
  fontWeight: 600,
  fontSize: 14,
}

const styleBtnSm: React.CSSProperties = {
  marginRight: 6,
  padding: '5px 12px',
  border: 'none',
  borderRadius: 5,
  cursor: 'pointer',
  fontSize: 13,
  background: '#e8eaff',
  color: '#4f6ef7',
  fontWeight: 500,
}

const styleBtnDanger: React.CSSProperties = {
  background: '#fff0f0',
  color: '#e53e3e',
}

const styleFormGrid: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 12,
}

const styleLabel: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 5,
  fontSize: 14,
  fontWeight: 500,
  color: '#444',
}

const styleInput: React.CSSProperties = {
  padding: '8px 12px',
  border: '1px solid #ddd',
  borderRadius: 6,
  fontSize: 14,
  outline: 'none',
}
