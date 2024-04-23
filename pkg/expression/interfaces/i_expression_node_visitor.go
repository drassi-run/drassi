package interfaces

// TODO: plan to add bi-cond interface

type IExpressionNodeVisitor interface {
	VisitAlwaysFn(eCtx IEvaluationContext, c IExpressionNode) any
	VisitAnd(eCtx IEvaluationContext, c IContainer) any
	VisitCancelledFn(eCtx IEvaluationContext, c IExpressionNode) any
	VisitContainsFn(eCtx IEvaluationContext, c IContainer) any
	VisitContextValueNode(eCtx IEvaluationContext, c IExpressionNode) any
	VisitEndsWithFn(eCtx IEvaluationContext, c IContainer) any
	VisitEqual(eCtx IEvaluationContext, c IContainer) any
	VisitFailureFn(eCtx IEvaluationContext, c IExpressionNode) any
	VisitFormatFn(eCtx IEvaluationContext, c IContainer) any
	VisitFromJsonFn(eCtx IEvaluationContext, c IContainer) any
	VisitGreaterThan(eCtx IEvaluationContext, c IContainer) any
	VisitGreaterThanOrEqual(eCtx IEvaluationContext, c IContainer) any
	VisitIndex(eCtx IEvaluationContext, c IContainer) any
	VisitJoinFn(eCtx IEvaluationContext, c IContainer) any
	VisitLessThan(eCtx IEvaluationContext, c IContainer) any
	VisitLessThanOrEqual(eCtx IEvaluationContext, c IContainer) any
	VisitLiteral(eCtx IEvaluationContext, c IExpressionNode) any
	VisitNoopFn(eCtx IEvaluationContext, c IExpressionNode) any
	VisitNoopNamedValue(eCtx IEvaluationContext, c IExpressionNode) any
	VisitNot(eCtx IEvaluationContext, c IContainer) any
	VisitNotEqual(eCtx IEvaluationContext, c IContainer) any
	VisitOr(eCtx IEvaluationContext, c IContainer) any
	VisitStartsWithFn(eCtx IEvaluationContext, c IContainer) any
	VisitSuccessFn(eCtx IEvaluationContext, c IExpressionNode) any
	VisitWildCard(eCtx IEvaluationContext, c IExpressionNode) any
}
