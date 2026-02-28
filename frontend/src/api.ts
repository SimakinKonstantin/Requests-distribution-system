import type { Appeal, Employee, Slot, Subtheme } from './types'

const BASE = import.meta.env.VITE_API_URL ?? ''

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })
  if (res.status === 204) return undefined as T
  if (!res.ok) throw new Error(`${res.status} ${res.statusText}`)
  return res.json() as Promise<T>
}

// ── Employees ──────────────────────────────────────────────────────────────
export const employeeApi = {
  getAll: () => request<Employee[]>('/employees'),
  getById: (id: number) => request<Employee>(`/employees/${id}`),
  create: (data: Omit<Employee, 'id'>) =>
    request<Employee>('/employees', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Employee, 'id'>) =>
    request<Employee>(`/employees/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/employees/${id}`, { method: 'DELETE' }),
}

// ── Slots ──────────────────────────────────────────────────────────────────
export const slotApi = {
  getAll: () => request<Slot[]>('/slots'),
  getById: (id: number) => request<Slot>(`/slots/${id}`),
  create: (data: Omit<Slot, 'id'>) =>
    request<Slot>('/slots', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Slot, 'id'>) =>
    request<Slot>(`/slots/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/slots/${id}`, { method: 'DELETE' }),
}

// ── Appeals ────────────────────────────────────────────────────────────────
export const appealApi = {
  getAll: () => request<Appeal[]>('/appeals'),
  getById: (id: number) => request<Appeal>(`/appeals/${id}`),
  create: (data: Omit<Appeal, 'id'>) =>
    request<Appeal>('/appeals', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Appeal, 'id'>) =>
    request<Appeal>(`/appeals/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/appeals/${id}`, { method: 'DELETE' }),
}

// ── Subthemes ──────────────────────────────────────────────────────────────
export const subthemeApi = {
  getAll: () => request<Subtheme[]>('/subthemes'),
  getById: (id: number) => request<Subtheme>(`/subthemes/${id}`),
  create: (data: Omit<Subtheme, 'id'>) =>
    request<Subtheme>('/subthemes', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Subtheme, 'id'>) =>
    request<Subtheme>(`/subthemes/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/subthemes/${id}`, { method: 'DELETE' }),
}
