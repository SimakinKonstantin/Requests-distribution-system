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
  teamId: number | null
  themeId: number
  subthemeId: number | null
  text: string
  status: 'active' | 'closed'
}

// Команды (для вкладки «Команды»)
export interface TeamThemeSubtheme {
  themeId: number
  subthemeId: number | null
  forVip: boolean
}
export interface Team {
  id: number
  name: string
  themeSubtheme?: TeamThemeSubtheme[] | null
}

export type WorkflowStatus = 'active' | 'paused'
export type WorkflowNodeType = 'ActionNode' | 'ConditionNode' | 'PredicateNode'
export type WorkflowComparison =
  | 'All'
  | 'Contains'
  | 'EndsWith'
  | 'Eq'
  | 'InInterval'
  | 'NotContains'
  | 'NotEq'
  | 'NotInInterval'
export type WorkflowPredicateAttribute = 'clientEmail' | 'messageCreatedAt' | 'text' | 'themeId'
export type WorkflowLogicOperator = 'and' | 'or'

export interface WorkflowGetAll {
  id: number
  name: string
}

export interface WorkflowEdge {
  id: string
  source: string
  target: string
}

export interface WorkflowPredicate {
  attribute?: WorkflowPredicateAttribute
  comparison?: WorkflowComparison
  values: string[]
}

export interface WorkflowConditionGroup {
  operator: WorkflowLogicOperator
  conditions: WorkflowPredicate[]
}

export interface WorkflowActionData {
  values: string[]
}

export interface WorkflowAction {
  actionType?: 'assignTeamAction'
  data?: WorkflowActionData
}

export interface WorkflowNode {
  id: string
  type: WorkflowNodeType
  data?: WorkflowPredicate | WorkflowConditionGroup | WorkflowAction
}

export interface Workflow {
  id: number
  name: string
  status: WorkflowStatus
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
}
