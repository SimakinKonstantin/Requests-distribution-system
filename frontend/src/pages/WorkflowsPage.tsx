import { useEffect, useState } from 'react'
import { clientApi, teamApi, themeApi, workflowApi } from '../api'
import Modal from '../components/Modal'
import type {
  Team,
  Theme,
  Workflow,
  WorkflowAction,
  WorkflowComparison,
  WorkflowConditionGroup,
  WorkflowGetAll,
  WorkflowLogicOperator,
  WorkflowNode,
  WorkflowPredicate,
  WorkflowPredicateAttribute,
  WorkflowStatus,
} from '../types'

// ─── Словари ──────────────────────────────────────────────────────────────────

const ATTRIBUTE_LABELS: Record<WorkflowPredicateAttribute, string> = {
  clientEmail: 'Email клиента',
  messageCreatedAt: 'Время создания обращения',
  text: 'Текст обращения',
  themeId: 'Тема обращения',
}

// Допустимые операции для каждого атрибута
const ATTRIBUTE_COMPARISONS: Record<WorkflowPredicateAttribute, WorkflowComparison[]> = {
  clientEmail:      ['Eq', 'NotEq'],
  messageCreatedAt: ['InInterval', 'NotInInterval'],
  text:             ['Contains', 'NotContains'],
  themeId:          ['Eq', 'NotEq'],
}

const COMPARISON_LABELS: Record<WorkflowComparison, string> = {
  All:           'Все',
  Contains:      'Содержит',
  EndsWith:      'Заканчивается на',
  Eq:            'Равно',
  InInterval:    'В интервале',
  NotContains:   'Не содержит',
  NotEq:         'Не равно',
  NotInInterval: 'Не в интервале',
}

const OPERATOR_LABELS: Record<WorkflowLogicOperator, string> = {
  and: 'И',
  or:  'ИЛИ',
}

// ─── Типы черновиков ──────────────────────────────────────────────────────────

type StartComparison = Extract<WorkflowComparison, 'Eq' | 'All'>

interface ActionNodeDraft {
  kind: 'ActionNode'
  id: string
  teamId: string
}

/** Черновик предиката внутри ConditionNode */
interface ConditionPredicateDraft {
  attribute: WorkflowPredicateAttribute
  comparison: WorkflowComparison
  // для clientEmail / themeId - список выбранных значений
  selectedValues: string[]
  // для text - произвольный текст
  valuesText: string
  // для messageCreatedAt - два времени (HH:MM, Moscow)
  intervalStart: string
  intervalEnd:   string
}

/** ConditionNode хранит плоский список предикатов + оператор (соответствует backend ConditionGroup) */
interface ConditionNodeDraft {
  kind: 'ConditionNode'
  id: string
  operator: WorkflowLogicOperator
  predicates: ConditionPredicateDraft[]
}

type ExtraNodeDraft = ActionNodeDraft | ConditionNodeDraft

// ─── Вспомогательные функции ──────────────────────────────────────────────────

const predicateAttributes: WorkflowPredicateAttribute[] = [
  'clientEmail',
  'messageCreatedAt',
  'text',
  'themeId',
]

function defaultComparisonFor(attr: WorkflowPredicateAttribute): WorkflowComparison {
  return ATTRIBUTE_COMPARISONS[attr][0]
}

function emptyConditionPredicate(): ConditionPredicateDraft {
  return {
    attribute:      'text',
    comparison:     'Contains',
    selectedValues: [],
    valuesText:     '',
    intervalStart:  '',
    intervalEnd:    '',
  }
}

/**
 * Конвертирует время пользователя (HH:MM, Europe/Moscow = UTC+3)
 * в RFC3339 UTC строку на сегодняшнюю дату по московскому времени.
 */
function moscowTimeToUtcIso(timeHHMM: string): string {
  const [hhStr, mmStr] = timeHHMM.split(':')
  const hours = Number(hhStr)
  const minutes = Number(mmStr)
  if (!Number.isFinite(hours) || !Number.isFinite(minutes)) {
    throw new Error('Неверный формат времени (ожидается HH:MM)')
  }

  // Получаем сегодняшнюю дату в московском времени (UTC+3)
  const now = new Date()
  const MOSCOW_OFFSET_MS = 3 * 60 * 60 * 1000
  const nowMoscow = new Date(now.getTime() + MOSCOW_OFFSET_MS)
  const y = nowMoscow.getUTCFullYear()
  const mo = nowMoscow.getUTCMonth()
  const d = nowMoscow.getUTCDate()

  // Строим момент в UTC: московская дата + введённое время - смещение UTC+3
  const mskMidnightUtcMs = Date.UTC(y, mo, d, 0, 0, 0, 0) - MOSCOW_OFFSET_MS
  const totalMinutes = hours * 60 + minutes
  const resultMs = mskMidnightUtcMs + totalMinutes * 60_000
  return new Date(resultMs).toISOString()
}

/**
 * Обратное преобразование: UTC ISO → HH:MM в московском времени.
 * Используется при заполнении формы редактирования из сохранённой автоматизации.
 */
function utcIsoToMoscowHHMM(iso: string): string {
  const date = new Date(iso)
  if (isNaN(date.getTime())) return ''
  const MOSCOW_OFFSET_MS = 3 * 60 * 60 * 1000
  const msk = new Date(date.getTime() + MOSCOW_OFFSET_MS)
  const hh = String(msk.getUTCHours()).padStart(2, '0')
  const mm = String(msk.getUTCMinutes()).padStart(2, '0')
  return `${hh}:${mm}`
}

/**
 * Десериализует сохранённый Workflow обратно в черновик формы.
 * Возвращает null если структура не распознана.
 */
function workflowToDraft(wf: Workflow): {
  name: string
  status: WorkflowStatus
  startComparison: StartComparison
  selectedClientEmails: string[]
  extraNodes: ExtraNodeDraft[]
} {
  let startComparison: StartComparison = 'All'
  let selectedClientEmails: string[] = []
  const extraNodes: ExtraNodeDraft[] = []

  for (const node of wf.nodes) {
    if (node.type === 'PredicateNode') {
      const data = node.data as WorkflowPredicate | undefined
      if (data?.comparison === 'Eq' || data?.comparison === 'All') {
        startComparison = data.comparison as StartComparison
        selectedClientEmails = data.values ?? []
      }
      continue
    }

    if (node.type === 'ActionNode') {
      const data = node.data as WorkflowAction | undefined
      const teamId = data?.data?.values?.[0] ?? ''
      extraNodes.push({ kind: 'ActionNode', id: node.id, teamId })
      continue
    }

    if (node.type === 'ConditionNode') {
      const data = node.data as WorkflowConditionGroup | undefined
      const operator: WorkflowLogicOperator = data?.operator ?? 'and'
      const predicates: ConditionPredicateDraft[] = (data?.conditions ?? []).map(pred => {
        const attr = (pred.attribute ?? 'text') as WorkflowPredicateAttribute
        const cmp = (pred.comparison ?? defaultComparisonFor(attr)) as WorkflowComparison
        const values = pred.values ?? []

        if (attr === 'messageCreatedAt') {
          return {
            attribute:      attr,
            comparison:     cmp,
            selectedValues: [],
            valuesText:     '',
            intervalStart:  values[0] ? utcIsoToMoscowHHMM(values[0]) : '',
            intervalEnd:    values[1] ? utcIsoToMoscowHHMM(values[1]) : '',
          }
        }
        if (attr === 'clientEmail' || attr === 'themeId') {
          return {
            attribute:      attr,
            comparison:     cmp,
            selectedValues: values,
            valuesText:     '',
            intervalStart:  '',
            intervalEnd:    '',
          }
        }
        // text
        return {
          attribute:      attr,
          comparison:     cmp,
          selectedValues: [],
          valuesText:     values.join('\n'),
          intervalStart:  '',
          intervalEnd:    '',
        }
      })
      extraNodes.push({
        kind: 'ConditionNode',
        id: node.id,
        operator,
        predicates: predicates.length > 0 ? predicates : [emptyConditionPredicate()],
      })
    }
  }

  return {
    name:                 wf.name,
    status:               wf.status,
    startComparison,
    selectedClientEmails,
    extraNodes,
  }
}

// ─── Вспомогательный компонент: редактор значений предиката ──────────────────

interface PredicateValueEditorProps {
  predicate: ConditionPredicateDraft
  clientEmails: string[]
  themes: Theme[]
  onChange: (updated: Partial<ConditionPredicateDraft>) => void
}

function PredicateValueEditor({ predicate, clientEmails, themes, onChange }: PredicateValueEditorProps) {
  const { attribute, selectedValues, valuesText, intervalStart, intervalEnd } = predicate

  if (attribute === 'clientEmail') {
    const toggle = (v: string) =>
      onChange({
        selectedValues: selectedValues.includes(v)
          ? selectedValues.filter(x => x !== v)
          : [...selectedValues, v],
      })
    return (
      <div style={labelStyle}>
        Email клиентов
        {clientEmails.length === 0 ? (
          <p style={hintStyle}>Список email пуст.</p>
        ) : (
          <div style={checkboxList}>
            {clientEmails.map(email => (
              <label key={email} style={checkboxItem}>
                <input
                  type="checkbox"
                  checked={selectedValues.includes(email)}
                  onChange={() => toggle(email)}
                />
                <span>{email}</span>
              </label>
            ))}
          </div>
        )}
      </div>
    )
  }

  if (attribute === 'themeId') {
    const toggle = (id: string) =>
      onChange({
        selectedValues: selectedValues.includes(id)
          ? selectedValues.filter(x => x !== id)
          : [...selectedValues, id],
      })
    return (
      <div style={labelStyle}>
        Темы обращений
        {themes.length === 0 ? (
          <p style={hintStyle}>Список тем пуст.</p>
        ) : (
          <div style={checkboxList}>
            {themes.map(t => (
              <label key={t.id} style={checkboxItem}>
                <input
                  type="checkbox"
                  checked={selectedValues.includes(String(t.id))}
                  onChange={() => toggle(String(t.id))}
                />
                <span>({t.id}) {t.name}</span>
              </label>
            ))}
          </div>
        )}
      </div>
    )
  }

  if (attribute === 'messageCreatedAt') {
    return (
      <div style={{ display: 'flex', gap: 8 }}>
        <label style={{ ...labelStyle, flex: 1 }}>
          Начало интервала (МСК)
          <input
            type="time"
            style={inputStyle}
            value={intervalStart}
            onChange={e => onChange({ intervalStart: e.target.value })}
          />
        </label>
        <label style={{ ...labelStyle, flex: 1 }}>
          Конец интервала (МСК)
          <input
            type="time"
            style={inputStyle}
            value={intervalEnd}
            onChange={e => onChange({ intervalEnd: e.target.value })}
          />
        </label>
      </div>
    )
  }

  return (
    <label style={labelStyle}>
      Текст
      <textarea
        style={{ ...inputStyle, minHeight: 48 }}
        placeholder="Через запятую или новую строку"
        value={valuesText}
        onChange={e => onChange({ valuesText: e.target.value })}
      />
    </label>
  )
}

// ─── Компонент формы автоматизации (создание и редактирование) ────────────────

interface WorkflowFormProps {
  initialName: string
  initialStatus: WorkflowStatus
  initialStartComparison: StartComparison
  initialSelectedClientEmails: string[]
  initialExtraNodes: ExtraNodeDraft[]
  clientEmails: string[]
  themes: Theme[]
  teams: Team[]
  submitLabel: string
  onSubmit: (payload: Omit<Workflow, 'id'>) => Promise<void>
}

function WorkflowForm({
  initialName,
  initialStatus,
  initialStartComparison,
  initialSelectedClientEmails,
  initialExtraNodes,
  clientEmails,
  themes,
  teams,
  submitLabel,
  onSubmit,
}: WorkflowFormProps) {
  const [name, setName]                                   = useState(initialName)
  const [status, setStatus]                               = useState<WorkflowStatus>(initialStatus)
  const [startComparison, setStartComparison]             = useState<StartComparison>(initialStartComparison)
  const [selectedClientEmails, setSelectedClientEmails]   = useState<string[]>(initialSelectedClientEmails)
  const [extraNodes, setExtraNodes]                       = useState<ExtraNodeDraft[]>(initialExtraNodes)
  const [formError, setFormError]                         = useState<string | null>(null)
  const [submitting, setSubmitting]                       = useState(false)

  // ── Конвертация и валидация ────────────────────────────────────────────────

  const resolveValues = (p: ConditionPredicateDraft, label: string): string[] => {
    if (p.attribute === 'clientEmail' || p.attribute === 'themeId') {
      if (p.selectedValues.length === 0) throw new Error(`${label}: выберите хотя бы одно значение`)
      return p.selectedValues
    }
    if (p.attribute === 'messageCreatedAt') {
      if (!p.intervalStart || !p.intervalEnd)
        throw new Error(`${label}: укажите начало и конец интервала`)
      return [moscowTimeToUtcIso(p.intervalStart), moscowTimeToUtcIso(p.intervalEnd)]
    }
    const values = p.valuesText.split(/[\n,]/).map(v => v.trim()).filter(Boolean)
    if (values.length === 0) throw new Error(`${label}: заполните текст`)
    return values
  }

  const buildPredicateFromDraft = (p: ConditionPredicateDraft, label: string): WorkflowPredicate => ({
    attribute:  p.attribute,
    comparison: p.comparison,
    values:     resolveValues(p, label),
  })

  const buildPayload = (): Omit<Workflow, 'id'> => {
    const workflowName = name.trim()
    if (!workflowName) throw new Error('Укажите название автоматизации')
    if (startComparison === 'Eq' && selectedClientEmails.length === 0)
      throw new Error('Выберите хотя бы один email клиента')

    const nodes: WorkflowNode[] = [
      {
        id:   'node-start-client-email',
        type: 'PredicateNode',
        data: {
          attribute:  'clientEmail',
          comparison: startComparison,
          values:     startComparison === 'All' ? [] : selectedClientEmails,
        },
      },
    ]

    for (let i = 0; i < extraNodes.length; i++) {
      const node = extraNodes[i]

      if (node.kind === 'ActionNode') {
        const tid = Number(node.teamId)
        if (!node.teamId || !Number.isInteger(tid) || tid <= 0)
          throw new Error(`Действие #${i + 1}: выберите команду из списка`)
        const actionData: WorkflowAction = {
          actionType: 'assignTeamAction',
          data: { values: [String(tid)] },
        }
        nodes.push({ id: node.id, type: 'ActionNode', data: actionData })
        continue
      }

      if (node.predicates.length === 0)
        throw new Error(`Условие #${i + 1}: добавьте хотя бы один предикат`)

      const conditionData: WorkflowConditionGroup = {
        operator:   node.operator,
        conditions: node.predicates.map((p, pi) =>
          buildPredicateFromDraft(p, `Условие #${i + 1}, предикат #${pi + 1}`)
        ),
      }
      nodes.push({ id: node.id, type: 'ConditionNode', data: conditionData })
    }

    const edges = nodes.slice(0, -1).map((n, i) => ({
      id: `edge-${i + 1}`, source: n.id, target: nodes[i + 1].id,
    }))

    return { name: workflowName, status, nodes, edges }
  }

  // ── Управление узлами ──────────────────────────────────────────────────────

  const addExtraNode = (kind: ExtraNodeDraft['kind']) => {
    if (kind === 'ActionNode') {
      setExtraNodes(prev => [...prev, { kind, id: uid(), teamId: '' }])
    } else {
      setExtraNodes(prev => [...prev, { kind, id: uid(), operator: 'and', predicates: [emptyConditionPredicate()] }])
    }
  }

  const updateExtraNode = (idx: number, updater: (n: ExtraNodeDraft) => ExtraNodeDraft) =>
    setExtraNodes(prev => prev.map((n, i) => i === idx ? updater(n) : n))

  const updatePredicate = (
    nodeIdx: number,
    predIdx: number,
    patch: Partial<ConditionPredicateDraft>,
  ) => {
    updateExtraNode(nodeIdx, node => {
      if (node.kind !== 'ConditionNode') return node
      return {
        ...node,
        predicates: node.predicates.map((p, pi) => {
          if (pi !== predIdx) return p
          if (patch.attribute && patch.attribute !== p.attribute) {
            return {
              ...emptyConditionPredicate(),
              attribute:  patch.attribute,
              comparison: defaultComparisonFor(patch.attribute),
            }
          }
          return { ...p, ...patch }
        }),
      }
    })
  }

  const toggleClientEmail = (email: string) =>
    setSelectedClientEmails(prev =>
      prev.includes(email) ? prev.filter(x => x !== email) : [...prev, email],
    )

  // ── Submit ─────────────────────────────────────────────────────────────────

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)
    setSubmitting(true)
    try {
      await onSubmit(buildPayload())
    } catch (err) {
      setFormError(String(err))
    } finally {
      setSubmitting(false)
    }
  }

  // ── Рендер ─────────────────────────────────────────────────────────────────

  return (
    <form onSubmit={handleSubmit} style={formStyle}>
      <label style={labelStyle}>
        Название
        <input style={inputStyle} value={name} onChange={e => setName(e.target.value)} required />
      </label>

      <label style={labelStyle}>
        Статус
        <select style={inputStyle} value={status} onChange={e => setStatus(e.target.value as WorkflowStatus)}>
          <option value="active">Активный</option>
          <option value="paused">Неактивный</option>
        </select>
      </label>

      {/* ── Блок клиентов ── */}
      <div style={sectionCard}>
        <h4 style={sectionTitle}>Клиенты</h4>
        <p style={hintStyle}>Укажите клиентов, для которых будет срабатывать автоматизация.</p>
        <label style={labelStyle}>
          Режим
          <select
            style={inputStyle}
            value={startComparison}
            onChange={e => { setStartComparison(e.target.value as StartComparison); setSelectedClientEmails([]) }}
          >
            <option value="All">Все клиенты</option>
            <option value="Eq">Любой из списка</option>
          </select>
        </label>
        {startComparison === 'Eq' && (
          <div style={labelStyle}>
            Email-адреса клиентов
            {clientEmails.length === 0 ? (
              <p style={hintStyle}>Список email пуст.</p>
            ) : (
              <div style={checkboxList}>
                {clientEmails.map(email => (
                  <label key={email} style={checkboxItem}>
                    <input
                      type="checkbox"
                      checked={selectedClientEmails.includes(email)}
                      onChange={() => toggleClientEmail(email)}
                    />
                    <span>{email}</span>
                  </label>
                ))}
              </div>
            )}
          </div>
        )}
      </div>

      {/* ── Шаги ── */}
      <div style={sectionCard}>
        <h4 style={sectionTitle}>Добавить шаг</h4>
        <div style={actionsRow}>
          <button type="button" style={smallBtn} onClick={() => addExtraNode('ConditionNode')}>+ Условие</button>
          <button type="button" style={smallBtn} onClick={() => addExtraNode('ActionNode')}>+ Действие</button>
        </div>

        {extraNodes.length === 0 && (
          <p style={hintStyle}>Шаги не добавлены. Цепочка состоит только из фильтра клиентов.</p>
        )}

        {extraNodes.map((node, nodeIdx) => (
          <div key={node.id} style={nodeCard}>
            <div style={nodeHeader}>
              <strong>Шаг {nodeIdx + 2}. {node.kind === 'ActionNode' ? 'Действие' : 'Условие'}</strong>
              <button
                type="button"
                style={{ ...smallBtn, ...dangerBtn }}
                onClick={() => setExtraNodes(prev => prev.filter((_, i) => i !== nodeIdx))}
              >
                Удалить
              </button>
            </div>

            {/* ActionNode */}
            {node.kind === 'ActionNode' && (
              <div style={labelStyle}>
                Команда, на которую назначить обращение
                {teams.length === 0 ? (
                  <p style={hintStyle}>Список команд пуст.</p>
                ) : (
                  <div style={checkboxList}>
                    {teams.map(team => (
                      <label key={team.id} style={checkboxItem}>
                        <input
                          type="radio"
                          name={`team-${node.id}`}
                          value={String(team.id)}
                          checked={node.teamId === String(team.id)}
                          onChange={() =>
                            updateExtraNode(nodeIdx, n =>
                              n.kind === 'ActionNode' ? { ...n, teamId: String(team.id) } : n,
                            )
                          }
                        />
                        <span>({team.id}) {team.name}</span>
                      </label>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* ConditionNode */}
            {node.kind === 'ConditionNode' && (
              <div style={conditionCard}>
                <label style={labelStyle}>
                  Оператор объединения
                  <select
                    style={inputStyle}
                    value={node.operator}
                    onChange={e =>
                      updateExtraNode(nodeIdx, n =>
                        n.kind === 'ConditionNode'
                          ? { ...n, operator: e.target.value as WorkflowLogicOperator }
                          : n,
                      )
                    }
                  >
                    {Object.entries(OPERATOR_LABELS).map(([v, l]) => (
                      <option key={v} value={v}>{l}</option>
                    ))}
                  </select>
                </label>

                {node.predicates.map((pred, predIdx) => (
                  <div key={predIdx} style={predicateCard}>
                    <div style={nodeHeader}>
                      <span style={{ fontSize: 13, color: '#666' }}>Предикат #{predIdx + 1}</span>
                      <button
                        type="button"
                        style={{ ...smallBtn, ...dangerBtn }}
                        disabled={node.predicates.length === 1}
                        onClick={() =>
                          updateExtraNode(nodeIdx, n =>
                            n.kind === 'ConditionNode'
                              ? { ...n, predicates: n.predicates.filter((_, j) => j !== predIdx) }
                              : n,
                          )
                        }
                      >
                        Удалить
                      </button>
                    </div>

                    <label style={labelStyle}>
                      Атрибут
                      <select
                        style={inputStyle}
                        value={pred.attribute}
                        onChange={e =>
                          updatePredicate(nodeIdx, predIdx, {
                            attribute: e.target.value as WorkflowPredicateAttribute,
                          })
                        }
                      >
                        {predicateAttributes.map(attr => (
                          <option key={attr} value={attr}>{ATTRIBUTE_LABELS[attr]}</option>
                        ))}
                      </select>
                    </label>

                    <label style={labelStyle}>
                      Операция
                      <select
                        style={inputStyle}
                        value={pred.comparison}
                        onChange={e =>
                          updatePredicate(nodeIdx, predIdx, {
                            comparison: e.target.value as WorkflowComparison,
                          })
                        }
                      >
                        {ATTRIBUTE_COMPARISONS[pred.attribute].map(cmp => (
                          <option key={cmp} value={cmp}>{COMPARISON_LABELS[cmp]}</option>
                        ))}
                      </select>
                    </label>

                    <PredicateValueEditor
                      predicate={pred}
                      clientEmails={clientEmails}
                      themes={themes}
                      onChange={patch => updatePredicate(nodeIdx, predIdx, patch)}
                    />
                  </div>
                ))}

                <button
                  type="button"
                  style={smallBtn}
                  onClick={() =>
                    updateExtraNode(nodeIdx, n =>
                      n.kind === 'ConditionNode'
                        ? { ...n, predicates: [...n.predicates, emptyConditionPredicate()] }
                        : n,
                    )
                  }
                >
                  + Предикат
                </button>
              </div>
            )}
          </div>
        ))}
      </div>

      {formError && <p style={{ color: '#e53e3e', margin: 0 }}>{formError}</p>}

      <button style={primaryBtn} type="submit" disabled={submitting}>
        {submitting ? 'Сохраняем...' : submitLabel}
      </button>
    </form>
  )
}

// ─── Главный компонент ────────────────────────────────────────────────────────

export default function WorkflowsPage() {
  const [items, setItems]     = useState<WorkflowGetAll[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError]     = useState<string | null>(null)

  // Справочники (загружаются один раз)
  const [clientEmails, setClientEmails] = useState<string[]>([])
  const [themes, setThemes]             = useState<Theme[]>([])
  const [teams, setTeams]               = useState<Team[]>([])

  // Состояние модалок
  const [createOpen, setCreateOpen]           = useState(false)
  const [editWorkflow, setEditWorkflow]       = useState<Workflow | null>(null)
  const [editLoading, setEditLoading]         = useState(false)

  // ── Загрузка ──────────────────────────────────────────────────────────────

  const loadList = async () => {
    setLoading(true); setError(null)
    try { setItems((await workflowApi.getAll()) ?? []) }
    catch (e) { setError(String(e)) }
    finally { setLoading(false) }
  }

  useEffect(() => { void loadList() }, [])

  useEffect(() => {
    clientApi.getEmails().then(d => setClientEmails(d ?? [])).catch(() => setClientEmails([]))
    themeApi.getAll().then(d => setThemes(d ?? [])).catch(() => setThemes([]))
    teamApi.getAll().then(d => setTeams(d ?? [])).catch(() => setTeams([]))
  }, [])

  // ── Открытие редактора ────────────────────────────────────────────────────

  const openEdit = async (id: number) => {
    setEditLoading(true)
    try {
      const wf = await workflowApi.getById(id)
      setEditWorkflow(wf)
    } catch (e) {
      setError(String(e))
    } finally {
      setEditLoading(false)
    }
  }

  // ── Удаление ──────────────────────────────────────────────────────────────

  const removeWorkflow = async (id: number) => {
    if (!window.confirm('Удалить автоматизацию?')) return
    try { await workflowApi.delete(id); setItems(prev => prev.filter(x => x.id !== id)) }
    catch (e) { setError(String(e)) }
  }

  // ── Рендер ───────────────────────────────────────────────────────────────

  // Черновик для редактирования (вычисляется из editWorkflow)
  const editDraft = editWorkflow ? workflowToDraft(editWorkflow) : null

  return (
    <div style={pageStyle}>
      <div style={titleRow}>
        <h2 style={{ margin: 0 }}>Автоматизация</h2>
        <button style={primaryBtn} onClick={() => setCreateOpen(true)}>Создать</button>
      </div>

      {loading && <p>Загрузка...</p>}
      {error   && <p style={{ color: '#e53e3e' }}>{error}</p>}

      <table style={tableStyle}>
        <thead>
          <tr>
            <th style={thStyle}>ID</th>
            <th style={thStyle}>Название</th>
            <th style={thStyle}></th>
          </tr>
        </thead>
        <tbody>
          {items.map(item => (
            <tr key={item.id}>
              <td style={tdStyle}>{item.id}</td>
              <td style={tdStyle}>{item.name}</td>
              <td style={tdStyle}>
                <button style={smallBtn} onClick={() => void openEdit(item.id)}>
                  {editLoading ? '...' : 'Открыть'}
                </button>
                <button style={{ ...smallBtn, ...dangerBtn }} onClick={() => void removeWorkflow(item.id)}>
                  Удалить
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {/* ── Модалка создания ── */}
      {createOpen && (
        <Modal title="Создание автоматизации" onClose={() => setCreateOpen(false)}>
          <WorkflowForm
            initialName=""
            initialStatus="active"
            initialStartComparison="Eq"
            initialSelectedClientEmails={[]}
            initialExtraNodes={[]}
            clientEmails={clientEmails}
            themes={themes}
            teams={teams}
            submitLabel="Создать автоматизацию"
            onSubmit={async payload => {
              await workflowApi.create(payload)
              setCreateOpen(false)
              await loadList()
            }}
          />
        </Modal>
      )}

      {/* ── Модалка редактирования ── */}
      {editWorkflow && editDraft && (
        <Modal
          title={`ID автоматизации: ${editWorkflow.id}`}
          onClose={() => setEditWorkflow(null)}
        >
          <WorkflowForm
            key={editWorkflow.id}
            initialName={editDraft.name}
            initialStatus={editDraft.status}
            initialStartComparison={editDraft.startComparison}
            initialSelectedClientEmails={editDraft.selectedClientEmails}
            initialExtraNodes={editDraft.extraNodes}
            clientEmails={clientEmails}
            themes={themes}
            teams={teams}
            submitLabel="Сохранить"
            onSubmit={async payload => {
              await workflowApi.update(editWorkflow.id, payload)
              setEditWorkflow(null)
              await loadList()
            }}
          />
        </Modal>
      )}
    </div>
  )
}

// ─── Утилиты ──────────────────────────────────────────────────────────────────

function uid() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  return `node-${Date.now()}-${Math.random().toString(16).slice(2)}`
}

// ─── Стили ────────────────────────────────────────────────────────────────────

const pageStyle: React.CSSProperties = { padding: '0 4px' }
const titleRow: React.CSSProperties = {
  display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20,
}
const tableStyle: React.CSSProperties = {
  width: '100%', borderCollapse: 'collapse', background: '#fff',
  borderRadius: 8, overflow: 'hidden', boxShadow: '0 1px 6px rgba(0,0,0,.07)',
}
const thStyle: React.CSSProperties = {
  textAlign: 'left', padding: '12px 16px', background: '#f5f6fa',
  fontWeight: 600, color: '#555', fontSize: 13,
}
const tdStyle: React.CSSProperties = {
  padding: '11px 16px', borderTop: '1px solid #f0f0f0', fontSize: 14,
}
const primaryBtn: React.CSSProperties = {
  background: '#4f6ef7', color: '#fff', border: 'none', borderRadius: 6,
  padding: '8px 18px', cursor: 'pointer', fontWeight: 600, fontSize: 14,
}
const smallBtn: React.CSSProperties = {
  marginRight: 6, padding: '5px 12px', border: 'none', borderRadius: 5,
  cursor: 'pointer', fontSize: 13, background: '#e8eaff', color: '#4f6ef7', fontWeight: 500,
}
const dangerBtn: React.CSSProperties = { background: '#fff0f0', color: '#e53e3e' }
const formStyle: React.CSSProperties = { display: 'flex', flexDirection: 'column', gap: 12 }
const labelStyle: React.CSSProperties = {
  display: 'flex', flexDirection: 'column', gap: 5, fontSize: 14, fontWeight: 500, color: '#444',
}
const inputStyle: React.CSSProperties = {
  padding: '8px 12px', border: '1px solid #ddd', borderRadius: 6,
  fontSize: 14, outline: 'none', fontFamily: 'inherit',
}
const sectionCard: React.CSSProperties = {
  border: '1px solid #edf0f8', borderRadius: 8, padding: 10,
  background: '#fafbff', display: 'flex', flexDirection: 'column', gap: 10,
}
const sectionTitle: React.CSSProperties = { margin: 0, fontSize: 15 }
const hintStyle: React.CSSProperties = { margin: 0, color: '#666', fontSize: 12 }
const checkboxList: React.CSSProperties = {
  border: '1px solid #e7e9f4', borderRadius: 8, padding: 8,
  maxHeight: 180, overflowY: 'auto', display: 'flex', flexDirection: 'column', gap: 6,
  background: '#fff',
}
const checkboxItem: React.CSSProperties = {
  display: 'flex', alignItems: 'center', gap: 8, fontWeight: 400,
}
const actionsRow: React.CSSProperties = { display: 'flex', flexWrap: 'wrap', gap: 8 }
const nodeCard: React.CSSProperties = {
  border: '1px solid #e7e9f4', borderRadius: 8, padding: 10,
  display: 'flex', flexDirection: 'column', gap: 8, background: '#fff',
}
const nodeHeader: React.CSSProperties = {
  display: 'flex', justifyContent: 'space-between', alignItems: 'center', gap: 10,
}
const conditionCard: React.CSSProperties = {
  border: '1px dashed #dfe3f0', borderRadius: 8, padding: 10,
  display: 'flex', flexDirection: 'column', gap: 8,
}
const predicateCard: React.CSSProperties = {
  border: '1px solid #edf0f8', borderRadius: 8, padding: 10,
  display: 'flex', flexDirection: 'column', gap: 8,
}
