package workflow

type Edge struct {
	Id     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type Node struct {
	Id   string       `json:"id"`
	Data *interface{} `json:"data"`
	// Position *struct {
	// 	X float64 `json:"x"`
	// 	Y float64 `json:"y"`
	// } `json:"position,omitempty"`
	Type NodeType `json:"type"`
}

type BlockResult struct {
	TeamId    string `json:"teamId"`
	ManagerId string `json:"managerId"`
}

type NodeType string

const (
	ActionNode    NodeType = "ActionNode"
	ConditionNode NodeType = "ConditionNode"
	PredicateNode NodeType = "PredicateNode"
)
