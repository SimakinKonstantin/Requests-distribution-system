export interface Employee {
  id: number
  name: string
  surname: string
  limit: number
  teamId: number
  email: string
}

export interface Client {
  id: number
  email: string
}

export interface Theme {
  id: number
  name: string
}

export interface Subtheme {
  id: number
  name: string
}

export interface Slot {
  id: number
  employeeId: number
  appealId: number
}

export interface Appeal {
  id: number
  clientId: number
  employeeId: number | null  // null = not yet assigned
  themeId: number
  subthemeId: number
  text: string
  status: 'active' | 'closed'
}
