package evaluator

import (
	"github.com/dungdm93/drasi/pkg/expression"
)

// TODO: next phase

type evaluationMemory struct {
	depths         []int
	maxAmount      int
	node           expression.IExpNode
	maxActiveDepth int
	totalAmount    int
}

func (e *evaluationMemory) addAmount(depth int, bytes int, trimDepth bool) {}

func (e *evaluationMemory) calculateBytes(obj any) int {
	return 0
}

func newEvaluationMemory(maxBytes int, node expression.IExpNode) *evaluationMemory {
	return &evaluationMemory{
		maxAmount: maxBytes,
		node:      node,
	}
}
