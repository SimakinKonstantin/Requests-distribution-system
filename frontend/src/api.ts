import type {
  Appeal,
  Client,
  Employee,
  Slot,
  Subtheme,
  Theme,
  Team,
  Workflow,
  WorkflowGetAll,
} from './types'

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

export const employeeApi = {
  getAll: () => request<Employee[]>('/employees'),
  getById: (id: number) => request<Employee>(`/employees/${id}`),
  create: (data: Omit<Employee, 'id'>) =>
    request<Employee>('/employees', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Employee, 'id'>) =>
    request<Employee>(`/employees/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/employees/${id}`, { method: 'DELETE' }),
}

export const clientApi = {
  getAll: () => request<Client[]>('/clients'),
  getEmails: () => request<string[]>('/clients/emails'),
  getById: (id: number) => request<Client>(`/clients/${id}`),
  create: (data: Omit<Client, 'id'>) =>
    request<Client>('/clients', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Client, 'id'>) =>
    request<Client>(`/clients/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/clients/${id}`, { method: 'DELETE' }),
}

export const themeApi = {
  getAll: () => request<Theme[]>('/themes'),
  getById: (id: number) => request<Theme>(`/themes/${id}`),
  create: (data: Omit<Theme, 'id'>) =>
    request<Theme>('/themes', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Theme, 'id'>) =>
    request<Theme>(`/themes/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/themes/${id}`, { method: 'DELETE' }),
}

export const slotApi = {
  getAll: () => request<Slot[]>('/slots'),
  getById: (id: number) => request<Slot>(`/slots/${id}`),
  create: (data: Omit<Slot, 'id'>) =>
    request<Slot>('/slots', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Slot, 'id'>) =>
    request<Slot>(`/slots/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/slots/${id}`, { method: 'DELETE' }),
}

export const appealApi = {
  getAll: () => request<Appeal[]>('/appeals'),
  getById: (id: number) => request<Appeal>(`/appeals/${id}`),
  create: (data: Omit<Appeal, 'id'>) =>
    request<Appeal>('/appeals', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Appeal, 'id'>) =>
    request<Appeal>(`/appeals/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/appeals/${id}`, { method: 'DELETE' }),
  close: (id: number) => request<Appeal>(`/appeals/${id}/close`, { method: 'POST' }),
}

export const subthemeApi = {
  getAll: () => request<Subtheme[]>('/subthemes'),
  getById: (id: number) => request<Subtheme>(`/subthemes/${id}`),
  create: (data: Omit<Subtheme, 'id'>) =>
    request<Subtheme>('/subthemes', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Subtheme, 'id'>) =>
    request<Subtheme>(`/subthemes/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/subthemes/${id}`, { method: 'DELETE' }),
}

export const teamApi = {
  getAll: () => request<Team[]>('/teams'),
  getById: (id: number) => request<Team>(`/teams/${id}`),
  create: (data: Omit<Team, 'id'>) =>
    request<Team>('/teams', { method: 'POST', body: JSON.stringify({ name: data.name, themeSubtheme: data.themeSubtheme ?? [] }) }),
  update: (id: number, data: Omit<Team, 'id'>) =>
    request<Team>(`/teams/${id}`, { method: 'PUT', body: JSON.stringify({ name: data.name, themeSubtheme: data.themeSubtheme ?? [] }) }),
  delete: (id: number) => request<void>(`/teams/${id}`, { method: 'DELETE' }),
}

export const workflowApi = {
  getAll: () => request<WorkflowGetAll[]>('/workflows'),
  getById: (id: number) => request<Workflow>(`/workflows/${id}`),
  create: (data: Omit<Workflow, 'id'>) =>
    request<Workflow>('/workflows', { method: 'POST', body: JSON.stringify(data) }),
  update: (id: number, data: Omit<Workflow, 'id'>) =>
    request<Workflow>(`/workflows/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  delete: (id: number) => request<void>(`/workflows/${id}`, { method: 'DELETE' }),
}
