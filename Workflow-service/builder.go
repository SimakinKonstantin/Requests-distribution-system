package workflow

import (
	"encoding/json"
	"workflow-service/gen"
)

type actionBlock interface {
	Do(data map[string]interface{}) actionBlockResult
	GetNext() actionBlock
	End() bool
	SetNext(next actionBlock)
}

func buildNextMap(edges []gen.Edge) map[string]string {
	next := make(map[string]string)
	for _, edge := range edges {
		next[edge.Source] = edge.Target
	}
	return next
}

func findStartNode(nodes []gen.Node, nextMap map[string]string) string {
	incoming := make(map[string]bool)

	for _, target := range nextMap {
		incoming[target] = true
	}

	for _, node := range nodes {
		if !incoming[node.Id] {
			return node.Id
		}
	}

	return ""
}

func (wm *Manager) buildChain(nodes []gen.Node, nextMap map[string]string, startId string) actionBlock {
	nodeMap := make(map[string]actionBlock)

	for _, node := range nodes {
		var block actionBlock

		switch node.Type {
		case gen.ActionNode:
			actionBytes, err := json.Marshal(node.Data)
			if err != nil {
				wm.logger.Error("failed to Marshal action")
				continue
			}

			var action gen.Action
			if err := json.Unmarshal(actionBytes, &action); err != nil {
				wm.logger.Error("failed to Unmarshal action")
				continue
			}

			if action.ActionType == nil || action.Data == nil {
				continue
			}

			if *action.ActionType == gen.AssignTeamAction {
				block = newActionBlockAssignTeam(wm.backEndClient, action.Data.Values, wm.logger)
			}

		case gen.PredicateNode:
			predicateBytes, err := json.Marshal(node.Data)
			if err != nil {
				wm.logger.Error("failed to Marshal predicate")
				continue
			}

			var predicate gen.Predicate
			if err := json.Unmarshal(predicateBytes, &predicate); err != nil {
				wm.logger.Error("failed to Unmarshal predicate")
				continue
			}

			block = newPredicateBlock(predicate)

		case gen.ConditionNode:
			conditionBytes, err := json.Marshal(node.Data)
			if err != nil {
				wm.logger.Error("failed to Marshal condition")
				continue
			}

			var conditionGroup gen.ConditionGroup
			if err := json.Unmarshal(conditionBytes, &conditionGroup); err != nil {
				wm.logger.Error("failed to Unmarshal condition")
				continue
			}

			block = newConditionBlock(conditionGroup, wm.apiSmartChatClient)
		}

		nodeMap[node.Id] = block
	}

	for _, node := range nodes {
		if target, exists := nextMap[node.Id]; exists {
			if current, ok := nodeMap[node.Id]; ok {
				if nextBlock, ok := nodeMap[target]; ok {
					current.SetNext(nextBlock)
				}
			}
		}
	}

	return nodeMap[startId]
}
