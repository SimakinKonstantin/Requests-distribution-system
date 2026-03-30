export interface Employee {
  id: number
  name: string
  surname: string
  limit: number
  teamId: number
  email: string
  // расширение формы: множественная привязка команд
  teamIds?: number[]
  // статус сотрудника
  status: 'working' | 'break'
}

export interface Client {
  id: number
  email: string
  name: string
  surname: string
  isVip: boolean
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
  appealId: number | null
  needToRemove: boolean
}

export interface Appeal {
  id: number
  clientId: number
  employeeId: number | null
  themeId: number
  subthemeId: number
  text: string
  status: 'active' | 'closed'
}

// Teams (для вкладки Teams)
export interface TeamThemeSubtheme {
  theme_id: number
  subtheme_id: number
  for_vip: boolean
}
export interface Team {
  id: number
  name: string
  // backend может присылать themeSubtheme/null
  themeSubtheme?: TeamThemeSubtheme[] | null
  // и/или "ThemeSubtheme" (без json-тэга) — обработаем на уровне UI
  // ThemeSubtheme?: TeamThemeSubtheme[] | null
}
