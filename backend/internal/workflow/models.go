package workflow

type GetAllResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Edge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type Node struct {
	ID   string       `json:"id"`
	Data *interface{} `json:"data"`
	Type NodeType     `json:"type"`
}

type BlockResult struct {
	TeamID    int `json:"teamId"`
	ManagerID int `json:"managerId"`
}

type NodeType string

const (
	ActionNode    NodeType = "ActionNode"
	ConditionNode NodeType = "ConditionNode"
	PredicateNode NodeType = "PredicateNode"
)

type ConditionGroupOperator string

const (
	ConditionGroupOperatorAnd ConditionGroupOperator = "and"
	ConditionGroupOperatorOr  ConditionGroupOperator = "or"
)

type ConditionGroup struct {
	Conditions []Predicate            `json:"conditions"`
	Operator   ConditionGroupOperator `json:"operator"`
}

type ActionActionType string

const (
	AssignTeamAction ActionActionType = "assignTeamAction"
)

type PredicateAttribute string

const (
	ClientEmail      PredicateAttribute = "clientEmail"
	MessageCreatedAt PredicateAttribute = "messageCreatedAt"
	Text             PredicateAttribute = "text"
	ThemeId          PredicateAttribute = "themeId"
)

type PredicateComparison string

const (
	All           PredicateComparison = "All"
	Contains      PredicateComparison = "Contains"
	Eq            PredicateComparison = "Eq"
	InInterval    PredicateComparison = "InInterval"
	NotContains   PredicateComparison = "NotContains"
	NotEq         PredicateComparison = "NotEq"
	NotInInterval PredicateComparison = "NotInInterval"
)

type Predicate struct {
	// Поле, по которому будет проводится проверка
	Attribute *PredicateAttribute `json:"attribute,omitempty"`

	// Выражение для сравнения (eq, contains, not_contains, in_interval, not_in_interval, ends_with, all)
	Comparison *PredicateComparison `json:"comparison,omitempty"`

	// Значения, которые нужно сравнить
	Values []string `json:"values"`
}

type Status string

const (
	StatusLive   Status = "active"
	StatusPaused Status = "paused"
)

type Action struct {
	ActionType *ActionActionType `json:"actionType,omitempty"`
	Data       *ActionData       `json:"data,omitempty"`
}

type ActionData struct {
	Values []string `json:"values"`
}

type Workflow struct {
	Edges []Edge `json:"edges"`

	// Идентификатор воркфлоу
	ID int `json:"id"`

	// Имя воркфлоу
	Name   string `json:"name"`
	Nodes  []Node `json:"nodes"`
	Status Status `json:"status"`
}
