package evaluator

import (
	"github.com/dungdm93/drasi/pkg/expression/interfaces"
)

// TODO: next phase

type evaluationMemory struct {
	depths         []int
	maxAmount      int
	node           interfaces.IExpressionNode
	maxActiveDepth int
	totalAmount    int
}

func (e *evaluationMemory) addAmount(depth int, bytes int, trimDepth bool) {}

func (e *evaluationMemory) calculateBytes(obj any) int {
	return 0
}

func newEvaluationMemory(maxBytes int, node interfaces.IExpressionNode) *evaluationMemory {
	return &evaluationMemory{
		maxAmount: maxBytes,
		node:      node,
	}
}
