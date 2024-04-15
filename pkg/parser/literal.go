package parser

type Literal struct {
	ExpressionNode
	value any
	kind  ValueKind
	name  string
}

func NewLiteral(val any) *Literal {
	value, kind, _ := convertToCanonicalValue(val)
	return &Literal{
		value: value,
		kind:  kind,
		name:  kind.ToString(),
	}
}

func (l *Literal) Value() any {
	return l.value
}

func (l *Literal) traceFullyRealized() bool {
	return false
}

func (l *Literal) evaluateCore(eCtx *EvaluationContext) any {
	return l.value
}

func (l *Literal) convertToExpression() string {
	return formatValue(nil, l.value, l.kind)
}

func (l *Literal) convertToRealizedExpression(eCtx *EvaluationContext) string {
	return formatValue(nil, l.value, l.kind)
}

func (l *Literal) getContainer() iContainer {
	return l.container
}

func (l *Literal) setContainer(c iContainer) {
	l.container = c
}

func (l *Literal) getLevel() (level int) {
	return l.level
}

func (l *Literal) getName() string {
	return l.name
}
func (l *Literal) setName(name string) {
	l.name = name
}
