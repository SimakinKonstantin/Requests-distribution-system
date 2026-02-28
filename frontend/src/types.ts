export interface Employee {
  id: number
  name: string
  surname: string
  limit: number
  teamId: number
}

export interface Slot {
  id: number
  employeeId: number
  appealId: number
}

export interface Appeal {
  id: number
  clientId: number
  employeeId: number
  themeId: number
  subthemeId: number
  text: string
}

export interface Subtheme {
  id: number
  name: string
}
