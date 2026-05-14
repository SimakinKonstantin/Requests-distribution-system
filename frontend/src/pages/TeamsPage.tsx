import { useState } from 'react'
import { teamApi, themeApi, subthemeApi } from '../api'
import { useCrud } from '../hooks/useCrud'
import Modal from '../components/Modal'
import type { Team, TeamThemeSubtheme, Theme, Subtheme } from '../types'
import {
  applyTextFieldValidity,
  PERSON_NAME_PATTERN,
} from '../validation'

type Form = { name: string; themeSubtheme: TeamThemeSubtheme[] }
const emptyRow = (): TeamThemeSubtheme => ({ themeId: 0, subthemeId: null, forVip: false })
const emptyForm = (): Form => ({ name: '', themeSubtheme: [] })

function normalizeRows(team: Team): TeamThemeSubtheme[] {
  const rows = (team as any).themeSubtheme ?? (team as any).ThemeSubtheme
  if (!rows || !Array.isArray(rows)) return []
  return rows.map((r: any) => ({
    themeId:    r.themeId    ?? r.theme_id    ?? 0,
    subthemeId: (r.subthemeId ?? r.subtheme_id ?? null) === 0 ? null : (r.subthemeId ?? r.subtheme_id ?? null),
    forVip:     r.forVip     ?? r.for_vip     ?? false,
  }))
}

export default function TeamsPage() {
  const { items, loading, error, create, update, remove } = useCrud<Team, Form>(teamApi)
  const [modal, setModal] = useState<'create' | 'edit' | null>(null)
  const [editing, setEditing] = useState<Team | null>(null)
  const [form, setForm] = useState<Form>(emptyForm())
  const [editLoading, setEditLoading] = useState(false)
  const [themes, setThemes] = useState<Theme[]>([])
  const [subthemes, setSubthemes] = useState<Subtheme[]>([])
  const [formError, setFormError] = useState<string | null>(null)

  // load dictionaries once
  useState(() => {
    themeApi.getAll().then(setThemes).catch(() => setThemes([]))
    subthemeApi.getAll().then(setSubthemes).catch(() => setSubthemes([]))
  })

  const openCreate = () => {
    setForm(emptyForm())
    setFormError(null)
    setModal('create')
  }

  const openEdit = async (item: Team) => {
    setFormError(null)
    setModal('edit')
    setEditLoading(true)
    try {
      const full = await teamApi.getById(item.id)
      setEditing(full)
      setForm({ name: full.name, themeSubtheme: normalizeRows(full) })
    } catch {
      // fallback to list item data if request fails
      setEditing(item)
      setForm({ name: item.name, themeSubtheme: normalizeRows(item) })
    } finally {
      setEditLoading(false)
    }
  }

  const close = () => {
    setModal(null)
    setEditing(null)
    setEditLoading(false)
    setFormError(null)
  }

  // ── row helpers ──────────────────────────────────────────────────────────────
  const addRow = () =>
    setForm(f => ({ ...f, themeSubtheme: [...f.themeSubtheme, emptyRow()] }))

  const removeRow = (idx: number) =>
    setForm(f => ({ ...f, themeSubtheme: f.themeSubtheme.filter((_, i) => i !== idx) }))

  const updateRow = (idx: number, patch: Partial<TeamThemeSubtheme>) =>
    setForm(f => ({
      ...f,
      themeSubtheme: f.themeSubtheme.map((r, i) => i === idx ? { ...r, ...patch } : r),
    }))

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    applyTextFieldValidity(e.currentTarget)
    if (!e.currentTarget.checkValidity()) {
      e.currentTarget.reportValidity()
      return
    }
    const normalized: Form = { ...form, name: form.name.trim() }
    const hasInvalidTheme = normalized.themeSubtheme.some(r => !Number.isInteger(r.themeId) || r.themeId <= 0)
    const validationError =
      (normalized.themeSubtheme.length === 0 ? 'Добавьте хотя бы одну привязку темы/подтемы' : null)
      ?? (hasInvalidTheme ? 'В каждой привязке должна быть выбрана тема' : null)
    if (validationError) {
      setFormError(validationError)
      return
    }
    if (modal === 'create') await create(normalized)
    else if (editing) await update(editing.id, normalized)
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
            {['ID', 'Название', ''].map(h => (
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
                data-text-field="true"
                pattern={PERSON_NAME_PATTERN}
                title="Только русские буквы, пробел и дефис"
                required
              />
            </label>

            {/* ── Bindings table ── */}
            <div style={{ fontWeight: 600, color: '#444', marginTop: 4 }}>Привязки тем / подтем</div>

            {editLoading && modal === 'edit' && <p>Загрузка…</p>}

            {!editLoading && form.themeSubtheme.length > 0 && (
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                {form.themeSubtheme.map((row, idx) => (
                  <div key={idx} style={styleRowCard}>
                    {/* Theme */}
                    <label style={styleInlineLabel}>
                      Тема
                      <select
                        style={styleSelect}
                        value={row.themeId}
                        onChange={e => updateRow(idx, { themeId: Number(e.target.value) })}
                        required
                      >
                        <option value={0} disabled>- выберите -</option>
                        {themes.map(t => (
                          <option key={t.id} value={t.id}>({t.id}) {t.name}</option>
                        ))}
                      </select>
                    </label>

                    {/* Subtheme */}
                    <label style={styleInlineLabel}>
                      Подтема
                      <select
                        style={styleSelect}
                        value={row.subthemeId ?? 0}
                        onChange={e => {
                          const v = Number(e.target.value)
                          updateRow(idx, { subthemeId: v === 0 ? null : v })
                        }}
                      >
                        <option value={0}>- нет -</option>
                        {subthemes.map(s => (
                          <option key={s.id} value={s.id}>({s.id}) {s.name}</option>
                        ))}
                      </select>
                    </label>

                    {/* VIP */}
                    <label style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 13 }}>
                      <input
                        type="checkbox"
                        checked={row.forVip}
                        onChange={e => updateRow(idx, { forVip: e.target.checked })}
                      />
                      VIP
                    </label>

                    {/* Remove row */}
                    <button
                      type="button"
                      style={{ ...styleBtnSm, ...styleBtnDanger, marginLeft: 'auto' }}
                      onClick={() => removeRow(idx)}
                    >
                      ✕
                    </button>
                  </div>
                ))}
              </div>
            )}

            <button
              type="button"
              style={{ ...styleBtnSm, alignSelf: 'flex-start' }}
              onClick={addRow}
            >
              + Добавить привязку
            </button>

            {formError && <p style={{ margin: 0, color: '#e53e3e' }}>{formError}</p>}
            <button style={{ ...styleBtnPrimary, marginTop: 8 }} type="submit" disabled={editLoading}>
              Сохранить
            </button>
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

const styleSelect: React.CSSProperties = {
  padding: '6px 10px',
  border: '1px solid #ddd',
  borderRadius: 6,
  fontSize: 13,
  outline: 'none',
  background: '#fff',
}

const styleRowCard: React.CSSProperties = {
  display: 'flex',
  alignItems: 'flex-end',
  gap: 10,
  background: '#f8f9ff',
  borderRadius: 7,
  padding: '10px 12px',
  border: '1px solid #e8eaff',
}

const styleInlineLabel: React.CSSProperties = {
  display: 'flex',
  flexDirection: 'column',
  gap: 3,
  fontSize: 12,
  fontWeight: 500,
  color: '#555',
}
