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
	// Position *struct {
	// 	X float64 `json:"x"`
	// 	Y float64 `json:"y"`
	// } `json:"position,omitempty"`
	Type NodeType `json:"type"`
}

type BlockResult struct {
	TeamID    string `json:"teamId"`
	ManagerID string `json:"managerId"`
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

type Condition struct {
	Predicates []Predicate `json:"predicates"`
}

type ConditionGroup struct {
	Conditions []Condition            `json:"conditions"`
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
	EndsWith      PredicateComparison = "EndsWith"
	Eq            PredicateComparison = "Eq"
	InInterval    PredicateComparison = "InInterval"
	NotContains   PredicateComparison = "NotContains"
	NotEq         PredicateComparison = "NotEq"
	NotInInterval PredicateComparison = "NotInInterval"
)

type Predicate struct {
	// Attribute Поле, по которому будет проводится проверка
	Attribute *PredicateAttribute `json:"attribute,omitempty"`

	// Comparison Выражение для сравнения (eq, contains, not_contains, in_interval, not_in_interval, ends_with, all)
	Comparison *PredicateComparison `json:"comparison,omitempty"`

	// Values Значения, которые нужно сравнить
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

	// Id Id воркфлоу
	ID int `json:"id"`

	// Name Имя воркфлоу
	Name   string `json:"name"`
	Nodes  []Node `json:"nodes"`
	Status Status `json:"status"`
}
