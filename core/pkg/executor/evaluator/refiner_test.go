package evaluator

import (
	"drassi.run/core/pkg/expression/ast"
	"drassi.run/core/pkg/expression/ast/operators"
	"github.com/stretchr/testify/assert"
	"testing"
)

var r = new(refiner)

func TestRefiner(t *testing.T) {
	t.Run("status", testStatusCondition)
	t.Run("refined", testRefinedCondition)
}

func testStatusCondition(t *testing.T) {
	tests := []string{
		"always()",
		"cancelled()",
		"success()",
		"failure()",

		"var && always()",
		"cancelled() || 1234",
		"x.y == success()",
		"failure() > x['y']",

		"always().how",
		"what[cancelled()]",
		"contains(array, success())",
		"format('job {0} status {1}', job, failure())",
	}
	for _, test := range tests {
		originNode, err := env.Parse(test, true)
		assert.NoError(t, err)

		refinedNode := r.Refine(originNode)
		assert.EqualValuesf(t, originNode, refinedNode, test)
	}
}

func testRefinedCondition(t *testing.T) {
	tests := []string{
		"var",
		"'str'",
		"3.14",
		"true",

		"x.y",
		"a['b']",
		"var && 'str'",
		"false || 1234",
		"x.y && a['b']",

		"contains(array, 'str')",
		"format('Hello {0} {1} {2}', 'Mona', 'the', 'Octocat')",

		// status function in the string
		"contains(array, 'always()')",
		"format('always() {0}', var)",
		"foo['success()']",
		"'failure()'",
	}
	for _, test := range tests {
		originNode, err := env.Parse(test, true)
		assert.NoError(t, err)

		expected := &ast.OperatorNode{
			Operator: operators.LogicalAnd,
			Operands: []ast.Node{
				&ast.FunctionNode{Name: "success"},
				originNode,
			},
		}

		refinedNode := r.Refine(originNode)
		assert.EqualValuesf(t, expected, refinedNode, test)
	}
}
