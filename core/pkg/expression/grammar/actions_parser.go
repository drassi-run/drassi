// Code generated from Actions.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // Actions
import (
	"fmt"
	"strconv"
	"sync"

	"github.com/antlr4-go/antlr/v4"
)

// Suppress unused import errors
var _ = fmt.Printf
var _ = strconv.Itoa
var _ = sync.Once{}

type ActionsParser struct {
	*antlr.BaseParser
}

var ActionsParserStaticData struct {
	once                   sync.Once
	serializedATN          []int32
	LiteralNames           []string
	SymbolicNames          []string
	RuleNames              []string
	PredictionContextCache *antlr.PredictionContextCache
	atn                    *antlr.ATN
	decisionToDFA          []*antlr.DFA
}

func actionsParserInit() {
	staticData := &ActionsParserStaticData
	staticData.LiteralNames = []string{
		"", "", "'('", "')'", "'['", "']'", "','", "'.'", "'=='", "'!='", "'<'",
		"'<='", "'>'", "'>='", "'&&'", "'||'", "'!'", "'*'", "'null'",
	}
	staticData.SymbolicNames = []string{
		"", "WS", "LPAREN", "RPAREN", "LBRACK", "RBRACK", "COMMA", "DOT", "EQUAL",
		"NOTEQUAL", "LT", "LTEQ", "GT", "GTEQ", "AND", "OR", "NOT", "WILDCARD",
		"NULL", "BOOLEAN", "INTEGER", "FLOAT", "STRING", "IDENTIFIER",
	}
	staticData.RuleNames = []string{
		"expression", "expr", "exprAccess", "identifier", "literal",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 23, 85, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 1, 0, 1, 0, 1, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 3, 1, 19, 8, 1, 1, 1,
		1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 5, 1,
		33, 8, 1, 10, 1, 12, 1, 36, 9, 1, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 5,
		2, 44, 8, 2, 10, 2, 12, 2, 47, 9, 2, 3, 2, 49, 8, 2, 1, 2, 1, 2, 1, 2,
		1, 2, 1, 2, 1, 2, 1, 2, 3, 2, 58, 8, 2, 1, 2, 1, 2, 1, 2, 4, 2, 63, 8,
		2, 11, 2, 12, 2, 64, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 4, 2, 72, 8, 2, 11,
		2, 12, 2, 73, 5, 2, 76, 8, 2, 10, 2, 12, 2, 79, 9, 2, 1, 3, 1, 3, 1, 4,
		1, 4, 1, 4, 0, 2, 2, 4, 5, 0, 2, 4, 6, 8, 0, 4, 1, 0, 10, 13, 1, 0, 8,
		9, 2, 0, 17, 17, 23, 23, 1, 0, 18, 22, 93, 0, 10, 1, 0, 0, 0, 2, 18, 1,
		0, 0, 0, 4, 57, 1, 0, 0, 0, 6, 80, 1, 0, 0, 0, 8, 82, 1, 0, 0, 0, 10, 11,
		3, 2, 1, 0, 11, 12, 5, 0, 0, 1, 12, 1, 1, 0, 0, 0, 13, 14, 6, 1, -1, 0,
		14, 19, 3, 4, 2, 0, 15, 16, 5, 16, 0, 0, 16, 19, 3, 2, 1, 6, 17, 19, 3,
		8, 4, 0, 18, 13, 1, 0, 0, 0, 18, 15, 1, 0, 0, 0, 18, 17, 1, 0, 0, 0, 19,
		34, 1, 0, 0, 0, 20, 21, 10, 5, 0, 0, 21, 22, 7, 0, 0, 0, 22, 33, 3, 2,
		1, 6, 23, 24, 10, 4, 0, 0, 24, 25, 7, 1, 0, 0, 25, 33, 3, 2, 1, 5, 26,
		27, 10, 3, 0, 0, 27, 28, 5, 14, 0, 0, 28, 33, 3, 2, 1, 4, 29, 30, 10, 2,
		0, 0, 30, 31, 5, 15, 0, 0, 31, 33, 3, 2, 1, 3, 32, 20, 1, 0, 0, 0, 32,
		23, 1, 0, 0, 0, 32, 26, 1, 0, 0, 0, 32, 29, 1, 0, 0, 0, 33, 36, 1, 0, 0,
		0, 34, 32, 1, 0, 0, 0, 34, 35, 1, 0, 0, 0, 35, 3, 1, 0, 0, 0, 36, 34, 1,
		0, 0, 0, 37, 38, 6, 2, -1, 0, 38, 39, 3, 6, 3, 0, 39, 48, 5, 2, 0, 0, 40,
		45, 3, 2, 1, 0, 41, 42, 5, 6, 0, 0, 42, 44, 3, 2, 1, 0, 43, 41, 1, 0, 0,
		0, 44, 47, 1, 0, 0, 0, 45, 43, 1, 0, 0, 0, 45, 46, 1, 0, 0, 0, 46, 49,
		1, 0, 0, 0, 47, 45, 1, 0, 0, 0, 48, 40, 1, 0, 0, 0, 48, 49, 1, 0, 0, 0,
		49, 50, 1, 0, 0, 0, 50, 51, 5, 3, 0, 0, 51, 58, 1, 0, 0, 0, 52, 53, 5,
		2, 0, 0, 53, 54, 3, 2, 1, 0, 54, 55, 5, 3, 0, 0, 55, 58, 1, 0, 0, 0, 56,
		58, 3, 6, 3, 0, 57, 37, 1, 0, 0, 0, 57, 52, 1, 0, 0, 0, 57, 56, 1, 0, 0,
		0, 58, 77, 1, 0, 0, 0, 59, 62, 10, 5, 0, 0, 60, 61, 5, 7, 0, 0, 61, 63,
		7, 2, 0, 0, 62, 60, 1, 0, 0, 0, 63, 64, 1, 0, 0, 0, 64, 62, 1, 0, 0, 0,
		64, 65, 1, 0, 0, 0, 65, 76, 1, 0, 0, 0, 66, 71, 10, 4, 0, 0, 67, 68, 5,
		4, 0, 0, 68, 69, 3, 2, 1, 0, 69, 70, 5, 5, 0, 0, 70, 72, 1, 0, 0, 0, 71,
		67, 1, 0, 0, 0, 72, 73, 1, 0, 0, 0, 73, 71, 1, 0, 0, 0, 73, 74, 1, 0, 0,
		0, 74, 76, 1, 0, 0, 0, 75, 59, 1, 0, 0, 0, 75, 66, 1, 0, 0, 0, 76, 79,
		1, 0, 0, 0, 77, 75, 1, 0, 0, 0, 77, 78, 1, 0, 0, 0, 78, 5, 1, 0, 0, 0,
		79, 77, 1, 0, 0, 0, 80, 81, 5, 23, 0, 0, 81, 7, 1, 0, 0, 0, 82, 83, 7,
		3, 0, 0, 83, 9, 1, 0, 0, 0, 10, 18, 32, 34, 45, 48, 57, 64, 73, 75, 77,
	}
	deserializer := antlr.NewATNDeserializer(nil)
	staticData.atn = deserializer.Deserialize(staticData.serializedATN)
	atn := staticData.atn
	staticData.decisionToDFA = make([]*antlr.DFA, len(atn.DecisionToState))
	decisionToDFA := staticData.decisionToDFA
	for index, state := range atn.DecisionToState {
		decisionToDFA[index] = antlr.NewDFA(state, index)
	}
}

// ActionsParserInit initializes any static state used to implement ActionsParser. By default the
// static state used to implement the parser is lazily initialized during the first call to
// NewActionsParser(). You can call this function if you wish to initialize the static state ahead
// of time.
func ActionsParserInit() {
	staticData := &ActionsParserStaticData
	staticData.once.Do(actionsParserInit)
}

// NewActionsParser produces a new parser instance for the optional input antlr.TokenStream.
func NewActionsParser(input antlr.TokenStream) *ActionsParser {
	ActionsParserInit()
	this := new(ActionsParser)
	this.BaseParser = antlr.NewBaseParser(input)
	staticData := &ActionsParserStaticData
	this.Interpreter = antlr.NewParserATNSimulator(this, staticData.atn, staticData.decisionToDFA, staticData.PredictionContextCache)
	this.RuleNames = staticData.RuleNames
	this.LiteralNames = staticData.LiteralNames
	this.SymbolicNames = staticData.SymbolicNames
	this.GrammarFileName = "Actions.g4"

	return this
}

// ActionsParser tokens.
const (
	ActionsParserEOF        = antlr.TokenEOF
	ActionsParserWS         = 1
	ActionsParserLPAREN     = 2
	ActionsParserRPAREN     = 3
	ActionsParserLBRACK     = 4
	ActionsParserRBRACK     = 5
	ActionsParserCOMMA      = 6
	ActionsParserDOT        = 7
	ActionsParserEQUAL      = 8
	ActionsParserNOTEQUAL   = 9
	ActionsParserLT         = 10
	ActionsParserLTEQ       = 11
	ActionsParserGT         = 12
	ActionsParserGTEQ       = 13
	ActionsParserAND        = 14
	ActionsParserOR         = 15
	ActionsParserNOT        = 16
	ActionsParserWILDCARD   = 17
	ActionsParserNULL       = 18
	ActionsParserBOOLEAN    = 19
	ActionsParserINTEGER    = 20
	ActionsParserFLOAT      = 21
	ActionsParserSTRING     = 22
	ActionsParserIDENTIFIER = 23
)

// ActionsParser rules.
const (
	ActionsParserRULE_expression = 0
	ActionsParserRULE_expr       = 1
	ActionsParserRULE_exprAccess = 2
	ActionsParserRULE_identifier = 3
	ActionsParserRULE_literal    = 4
)

// IExpressionContext is an interface to support dynamic dispatch.
type IExpressionContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetE returns the e rule contexts.
	GetE() IExprContext

	// SetE sets the e rule contexts.
	SetE(IExprContext)

	// Getter signatures
	EOF() antlr.TerminalNode
	Expr() IExprContext

	// IsExpressionContext differentiates from other interfaces.
	IsExpressionContext()
}

type ExpressionContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	e      IExprContext
}

func NewEmptyExpressionContext() *ExpressionContext {
	var p = new(ExpressionContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_expression
	return p
}

func InitEmptyExpressionContext(p *ExpressionContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_expression
}

func (*ExpressionContext) IsExpressionContext() {}

func NewExpressionContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExpressionContext {
	var p = new(ExpressionContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionsParserRULE_expression

	return p
}

func (s *ExpressionContext) GetParser() antlr.Parser { return s.parser }

func (s *ExpressionContext) GetE() IExprContext { return s.e }

func (s *ExpressionContext) SetE(v IExprContext) { s.e = v }

func (s *ExpressionContext) EOF() antlr.TerminalNode {
	return s.GetToken(ActionsParserEOF, 0)
}

func (s *ExpressionContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *ExpressionContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExpressionContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExpressionContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterExpression(s)
	}
}

func (s *ExpressionContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitExpression(s)
	}
}

func (s *ExpressionContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitExpression(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionsParser) Expression() (localctx IExpressionContext) {
	localctx = NewExpressionContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ActionsParserRULE_expression)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(10)

		var _x = p.expr(0)

		localctx.(*ExpressionContext).e = _x
	}
	{
		p.SetState(11)
		p.Match(ActionsParserEOF)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExprContext is an interface to support dynamic dispatch.
type IExprContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// GetOp returns the op token.
	GetOp() antlr.Token

	// SetOp sets the op token.
	SetOp(antlr.Token)

	// Getter signatures
	ExprAccess() IExprAccessContext
	AllExpr() []IExprContext
	Expr(i int) IExprContext
	NOT() antlr.TerminalNode
	Literal() ILiteralContext
	LT() antlr.TerminalNode
	LTEQ() antlr.TerminalNode
	GTEQ() antlr.TerminalNode
	GT() antlr.TerminalNode
	EQUAL() antlr.TerminalNode
	NOTEQUAL() antlr.TerminalNode
	AND() antlr.TerminalNode
	OR() antlr.TerminalNode

	// IsExprContext differentiates from other interfaces.
	IsExprContext()
}

type ExprContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
	op     antlr.Token
}

func NewEmptyExprContext() *ExprContext {
	var p = new(ExprContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_expr
	return p
}

func InitEmptyExprContext(p *ExprContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_expr
}

func (*ExprContext) IsExprContext() {}

func NewExprContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprContext {
	var p = new(ExprContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionsParserRULE_expr

	return p
}

func (s *ExprContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprContext) GetOp() antlr.Token { return s.op }

func (s *ExprContext) SetOp(v antlr.Token) { s.op = v }

func (s *ExprContext) ExprAccess() IExprAccessContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprAccessContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprAccessContext)
}

func (s *ExprContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *ExprContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *ExprContext) NOT() antlr.TerminalNode {
	return s.GetToken(ActionsParserNOT, 0)
}

func (s *ExprContext) Literal() ILiteralContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ILiteralContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(ILiteralContext)
}

func (s *ExprContext) LT() antlr.TerminalNode {
	return s.GetToken(ActionsParserLT, 0)
}

func (s *ExprContext) LTEQ() antlr.TerminalNode {
	return s.GetToken(ActionsParserLTEQ, 0)
}

func (s *ExprContext) GTEQ() antlr.TerminalNode {
	return s.GetToken(ActionsParserGTEQ, 0)
}

func (s *ExprContext) GT() antlr.TerminalNode {
	return s.GetToken(ActionsParserGT, 0)
}

func (s *ExprContext) EQUAL() antlr.TerminalNode {
	return s.GetToken(ActionsParserEQUAL, 0)
}

func (s *ExprContext) NOTEQUAL() antlr.TerminalNode {
	return s.GetToken(ActionsParserNOTEQUAL, 0)
}

func (s *ExprContext) AND() antlr.TerminalNode {
	return s.GetToken(ActionsParserAND, 0)
}

func (s *ExprContext) OR() antlr.TerminalNode {
	return s.GetToken(ActionsParserOR, 0)
}

func (s *ExprContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *ExprContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterExpr(s)
	}
}

func (s *ExprContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitExpr(s)
	}
}

func (s *ExprContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitExpr(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionsParser) Expr() (localctx IExprContext) {
	return p.expr(0)
}

func (p *ActionsParser) expr(_p int) (localctx IExprContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewExprContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExprContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 2
	p.EnterRecursionRule(localctx, 2, ActionsParserRULE_expr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(18)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ActionsParserLPAREN, ActionsParserIDENTIFIER:
		{
			p.SetState(14)
			p.exprAccess(0)
		}

	case ActionsParserNOT:
		{
			p.SetState(15)

			var _m = p.Match(ActionsParserNOT)

			localctx.(*ExprContext).op = _m
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(16)
			p.expr(6)
		}

	case ActionsParserNULL, ActionsParserBOOLEAN, ActionsParserINTEGER, ActionsParserFLOAT, ActionsParserSTRING:
		{
			p.SetState(17)
			p.Literal()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(34)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(32)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 1, p.GetParserRuleContext()) {
			case 1:
				localctx = NewExprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_expr)
				p.SetState(20)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(21)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*ExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&15360) != 0) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*ExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(22)
					p.expr(6)
				}

			case 2:
				localctx = NewExprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_expr)
				p.SetState(23)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(24)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*ExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !(_la == ActionsParserEQUAL || _la == ActionsParserNOTEQUAL) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*ExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(25)
					p.expr(5)
				}

			case 3:
				localctx = NewExprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_expr)
				p.SetState(26)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(27)

					var _m = p.Match(ActionsParserAND)

					localctx.(*ExprContext).op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(28)
					p.expr(4)
				}

			case 4:
				localctx = NewExprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_expr)
				p.SetState(29)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(30)

					var _m = p.Match(ActionsParserOR)

					localctx.(*ExprContext).op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(31)
					p.expr(3)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(36)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 2, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IExprAccessContext is an interface to support dynamic dispatch.
type IExprAccessContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser
	// IsExprAccessContext differentiates from other interfaces.
	IsExprAccessContext()
}

type ExprAccessContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyExprAccessContext() *ExprAccessContext {
	var p = new(ExprAccessContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_exprAccess
	return p
}

func InitEmptyExprAccessContext(p *ExprAccessContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_exprAccess
}

func (*ExprAccessContext) IsExprAccessContext() {}

func NewExprAccessContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *ExprAccessContext {
	var p = new(ExprAccessContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionsParserRULE_exprAccess

	return p
}

func (s *ExprAccessContext) GetParser() antlr.Parser { return s.parser }

func (s *ExprAccessContext) CopyAll(ctx *ExprAccessContext) {
	s.CopyFrom(&ctx.BaseParserRuleContext)
}

func (s *ExprAccessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *ExprAccessContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

type IndexAccessContext struct {
	ExprAccessContext
	_expr   IExprContext
	indexes []IExprContext
}

func NewIndexAccessContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *IndexAccessContext {
	var p = new(IndexAccessContext)

	InitEmptyExprAccessContext(&p.ExprAccessContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprAccessContext))

	return p
}

func (s *IndexAccessContext) Get_expr() IExprContext { return s._expr }

func (s *IndexAccessContext) Set_expr(v IExprContext) { s._expr = v }

func (s *IndexAccessContext) GetIndexes() []IExprContext { return s.indexes }

func (s *IndexAccessContext) SetIndexes(v []IExprContext) { s.indexes = v }

func (s *IndexAccessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IndexAccessContext) ExprAccess() IExprAccessContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprAccessContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprAccessContext)
}

func (s *IndexAccessContext) AllLBRACK() []antlr.TerminalNode {
	return s.GetTokens(ActionsParserLBRACK)
}

func (s *IndexAccessContext) LBRACK(i int) antlr.TerminalNode {
	return s.GetToken(ActionsParserLBRACK, i)
}

func (s *IndexAccessContext) AllRBRACK() []antlr.TerminalNode {
	return s.GetTokens(ActionsParserRBRACK)
}

func (s *IndexAccessContext) RBRACK(i int) antlr.TerminalNode {
	return s.GetToken(ActionsParserRBRACK, i)
}

func (s *IndexAccessContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *IndexAccessContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *IndexAccessContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterIndexAccess(s)
	}
}

func (s *IndexAccessContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitIndexAccess(s)
	}
}

func (s *IndexAccessContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitIndexAccess(s)

	default:
		return t.VisitChildren(s)
	}
}

type FunctionCallContext struct {
	ExprAccessContext
	_expr IExprContext
	args  []IExprContext
}

func NewFunctionCallContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *FunctionCallContext {
	var p = new(FunctionCallContext)

	InitEmptyExprAccessContext(&p.ExprAccessContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprAccessContext))

	return p
}

func (s *FunctionCallContext) Get_expr() IExprContext { return s._expr }

func (s *FunctionCallContext) Set_expr(v IExprContext) { s._expr = v }

func (s *FunctionCallContext) GetArgs() []IExprContext { return s.args }

func (s *FunctionCallContext) SetArgs(v []IExprContext) { s.args = v }

func (s *FunctionCallContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *FunctionCallContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *FunctionCallContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(ActionsParserLPAREN, 0)
}

func (s *FunctionCallContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(ActionsParserRPAREN, 0)
}

func (s *FunctionCallContext) AllExpr() []IExprContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IExprContext); ok {
			len++
		}
	}

	tst := make([]IExprContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IExprContext); ok {
			tst[i] = t.(IExprContext)
			i++
		}
	}

	return tst
}

func (s *FunctionCallContext) Expr(i int) IExprContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			if j == i {
				t = ctx.(antlr.RuleContext)
				break
			}
			j++
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *FunctionCallContext) AllCOMMA() []antlr.TerminalNode {
	return s.GetTokens(ActionsParserCOMMA)
}

func (s *FunctionCallContext) COMMA(i int) antlr.TerminalNode {
	return s.GetToken(ActionsParserCOMMA, i)
}

func (s *FunctionCallContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterFunctionCall(s)
	}
}

func (s *FunctionCallContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitFunctionCall(s)
	}
}

func (s *FunctionCallContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitFunctionCall(s)

	default:
		return t.VisitChildren(s)
	}
}

type VariableContext struct {
	ExprAccessContext
}

func NewVariableContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *VariableContext {
	var p = new(VariableContext)

	InitEmptyExprAccessContext(&p.ExprAccessContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprAccessContext))

	return p
}

func (s *VariableContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *VariableContext) Identifier() IIdentifierContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IIdentifierContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IIdentifierContext)
}

func (s *VariableContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterVariable(s)
	}
}

func (s *VariableContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitVariable(s)
	}
}

func (s *VariableContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitVariable(s)

	default:
		return t.VisitChildren(s)
	}
}

type PropertyAccessContext struct {
	ExprAccessContext
	_IDENTIFIER antlr.Token
	props       []antlr.Token
	_WILDCARD   antlr.Token
	_tset113    antlr.Token
}

func NewPropertyAccessContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *PropertyAccessContext {
	var p = new(PropertyAccessContext)

	InitEmptyExprAccessContext(&p.ExprAccessContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprAccessContext))

	return p
}

func (s *PropertyAccessContext) Get_IDENTIFIER() antlr.Token { return s._IDENTIFIER }

func (s *PropertyAccessContext) Get_WILDCARD() antlr.Token { return s._WILDCARD }

func (s *PropertyAccessContext) Get_tset113() antlr.Token { return s._tset113 }

func (s *PropertyAccessContext) Set_IDENTIFIER(v antlr.Token) { s._IDENTIFIER = v }

func (s *PropertyAccessContext) Set_WILDCARD(v antlr.Token) { s._WILDCARD = v }

func (s *PropertyAccessContext) Set_tset113(v antlr.Token) { s._tset113 = v }

func (s *PropertyAccessContext) GetProps() []antlr.Token { return s.props }

func (s *PropertyAccessContext) SetProps(v []antlr.Token) { s.props = v }

func (s *PropertyAccessContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PropertyAccessContext) ExprAccess() IExprAccessContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprAccessContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprAccessContext)
}

func (s *PropertyAccessContext) AllDOT() []antlr.TerminalNode {
	return s.GetTokens(ActionsParserDOT)
}

func (s *PropertyAccessContext) DOT(i int) antlr.TerminalNode {
	return s.GetToken(ActionsParserDOT, i)
}

func (s *PropertyAccessContext) AllIDENTIFIER() []antlr.TerminalNode {
	return s.GetTokens(ActionsParserIDENTIFIER)
}

func (s *PropertyAccessContext) IDENTIFIER(i int) antlr.TerminalNode {
	return s.GetToken(ActionsParserIDENTIFIER, i)
}

func (s *PropertyAccessContext) AllWILDCARD() []antlr.TerminalNode {
	return s.GetTokens(ActionsParserWILDCARD)
}

func (s *PropertyAccessContext) WILDCARD(i int) antlr.TerminalNode {
	return s.GetToken(ActionsParserWILDCARD, i)
}

func (s *PropertyAccessContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterPropertyAccess(s)
	}
}

func (s *PropertyAccessContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitPropertyAccess(s)
	}
}

func (s *PropertyAccessContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitPropertyAccess(s)

	default:
		return t.VisitChildren(s)
	}
}

type WrapContext struct {
	ExprAccessContext
}

func NewWrapContext(parser antlr.Parser, ctx antlr.ParserRuleContext) *WrapContext {
	var p = new(WrapContext)

	InitEmptyExprAccessContext(&p.ExprAccessContext)
	p.parser = parser
	p.CopyAll(ctx.(*ExprAccessContext))

	return p
}

func (s *WrapContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *WrapContext) LPAREN() antlr.TerminalNode {
	return s.GetToken(ActionsParserLPAREN, 0)
}

func (s *WrapContext) Expr() IExprContext {
	var t antlr.RuleContext
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IExprContext); ok {
			t = ctx.(antlr.RuleContext)
			break
		}
	}

	if t == nil {
		return nil
	}

	return t.(IExprContext)
}

func (s *WrapContext) RPAREN() antlr.TerminalNode {
	return s.GetToken(ActionsParserRPAREN, 0)
}

func (s *WrapContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterWrap(s)
	}
}

func (s *WrapContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitWrap(s)
	}
}

func (s *WrapContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitWrap(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionsParser) ExprAccess() (localctx IExprAccessContext) {
	return p.exprAccess(0)
}

func (p *ActionsParser) exprAccess(_p int) (localctx IExprAccessContext) {
	var _parentctx antlr.ParserRuleContext = p.GetParserRuleContext()

	_parentState := p.GetState()
	localctx = NewExprAccessContext(p, p.GetParserRuleContext(), _parentState)
	var _prevctx IExprAccessContext = localctx
	var _ antlr.ParserRuleContext = _prevctx // TODO: To prevent unused variable warning.
	_startState := 4
	p.EnterRecursionRule(localctx, 4, ActionsParserRULE_exprAccess, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(57)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext()) {
	case 1:
		localctx = NewFunctionCallContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(38)
			p.Identifier()
		}
		{
			p.SetState(39)
			p.Match(ActionsParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(48)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&16580612) != 0 {
			{
				p.SetState(40)

				var _x = p.expr(0)

				localctx.(*FunctionCallContext)._expr = _x
			}
			localctx.(*FunctionCallContext).args = append(localctx.(*FunctionCallContext).args, localctx.(*FunctionCallContext)._expr)
			p.SetState(45)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == ActionsParserCOMMA {
				{
					p.SetState(41)
					p.Match(ActionsParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(42)

					var _x = p.expr(0)

					localctx.(*FunctionCallContext)._expr = _x
				}
				localctx.(*FunctionCallContext).args = append(localctx.(*FunctionCallContext).args, localctx.(*FunctionCallContext)._expr)

				p.SetState(47)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(50)
			p.Match(ActionsParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 2:
		localctx = NewWrapContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(52)
			p.Match(ActionsParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(53)
			p.expr(0)
		}
		{
			p.SetState(54)
			p.Match(ActionsParserRPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}

	case 3:
		localctx = NewVariableContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx
		{
			p.SetState(56)
			p.Identifier()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(77)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(75)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 8, p.GetParserRuleContext()) {
			case 1:
				localctx = NewPropertyAccessContext(p, NewExprAccessContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_exprAccess)
				p.SetState(59)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				p.SetState(62)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = 1
				for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1:
						{
							p.SetState(60)
							p.Match(ActionsParserDOT)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}
						{
							p.SetState(61)

							var _lt = p.GetTokenStream().LT(1)

							localctx.(*PropertyAccessContext)._tset113 = _lt

							_la = p.GetTokenStream().LA(1)

							if !(_la == ActionsParserWILDCARD || _la == ActionsParserIDENTIFIER) {
								var _ri = p.GetErrorHandler().RecoverInline(p)

								localctx.(*PropertyAccessContext)._tset113 = _ri
							} else {
								p.GetErrorHandler().ReportMatch(p)
								p.Consume()
							}
						}
						localctx.(*PropertyAccessContext).props = append(localctx.(*PropertyAccessContext).props, localctx.(*PropertyAccessContext)._tset113)

					default:
						p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
						goto errorExit
					}

					p.SetState(64)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 6, p.GetParserRuleContext())
					if p.HasError() {
						goto errorExit
					}
				}

			case 2:
				localctx = NewIndexAccessContext(p, NewExprAccessContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_exprAccess)
				p.SetState(66)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				p.SetState(71)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = 1
				for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1:
						{
							p.SetState(67)
							p.Match(ActionsParserLBRACK)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}
						{
							p.SetState(68)

							var _x = p.expr(0)

							localctx.(*IndexAccessContext)._expr = _x
						}
						localctx.(*IndexAccessContext).indexes = append(localctx.(*IndexAccessContext).indexes, localctx.(*IndexAccessContext)._expr)
						{
							p.SetState(69)
							p.Match(ActionsParserRBRACK)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}

					default:
						p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
						goto errorExit
					}

					p.SetState(73)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 7, p.GetParserRuleContext())
					if p.HasError() {
						goto errorExit
					}
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(79)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext())
		if p.HasError() {
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.UnrollRecursionContexts(_parentctx)
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IIdentifierContext is an interface to support dynamic dispatch.
type IIdentifierContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	IDENTIFIER() antlr.TerminalNode

	// IsIdentifierContext differentiates from other interfaces.
	IsIdentifierContext()
}

type IdentifierContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyIdentifierContext() *IdentifierContext {
	var p = new(IdentifierContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_identifier
	return p
}

func InitEmptyIdentifierContext(p *IdentifierContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_identifier
}

func (*IdentifierContext) IsIdentifierContext() {}

func NewIdentifierContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *IdentifierContext {
	var p = new(IdentifierContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionsParserRULE_identifier

	return p
}

func (s *IdentifierContext) GetParser() antlr.Parser { return s.parser }

func (s *IdentifierContext) IDENTIFIER() antlr.TerminalNode {
	return s.GetToken(ActionsParserIDENTIFIER, 0)
}

func (s *IdentifierContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *IdentifierContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *IdentifierContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterIdentifier(s)
	}
}

func (s *IdentifierContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitIdentifier(s)
	}
}

func (s *IdentifierContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitIdentifier(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionsParser) Identifier() (localctx IIdentifierContext) {
	localctx = NewIdentifierContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 6, ActionsParserRULE_identifier)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(80)
		p.Match(ActionsParserIDENTIFIER)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// ILiteralContext is an interface to support dynamic dispatch.
type ILiteralContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	STRING() antlr.TerminalNode
	INTEGER() antlr.TerminalNode
	FLOAT() antlr.TerminalNode
	BOOLEAN() antlr.TerminalNode
	NULL() antlr.TerminalNode

	// IsLiteralContext differentiates from other interfaces.
	IsLiteralContext()
}

type LiteralContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyLiteralContext() *LiteralContext {
	var p = new(LiteralContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_literal
	return p
}

func InitEmptyLiteralContext(p *LiteralContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_literal
}

func (*LiteralContext) IsLiteralContext() {}

func NewLiteralContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *LiteralContext {
	var p = new(LiteralContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionsParserRULE_literal

	return p
}

func (s *LiteralContext) GetParser() antlr.Parser { return s.parser }

func (s *LiteralContext) STRING() antlr.TerminalNode {
	return s.GetToken(ActionsParserSTRING, 0)
}

func (s *LiteralContext) INTEGER() antlr.TerminalNode {
	return s.GetToken(ActionsParserINTEGER, 0)
}

func (s *LiteralContext) FLOAT() antlr.TerminalNode {
	return s.GetToken(ActionsParserFLOAT, 0)
}

func (s *LiteralContext) BOOLEAN() antlr.TerminalNode {
	return s.GetToken(ActionsParserBOOLEAN, 0)
}

func (s *LiteralContext) NULL() antlr.TerminalNode {
	return s.GetToken(ActionsParserNULL, 0)
}

func (s *LiteralContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *LiteralContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *LiteralContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterLiteral(s)
	}
}

func (s *LiteralContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitLiteral(s)
	}
}

func (s *LiteralContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitLiteral(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionsParser) Literal() (localctx ILiteralContext) {
	localctx = NewLiteralContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 8, ActionsParserRULE_literal)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(82)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&8126464) != 0) {
			p.GetErrorHandler().RecoverInline(p)
		} else {
			p.GetErrorHandler().ReportMatch(p)
			p.Consume()
		}
	}

errorExit:
	if p.HasError() {
		v := p.GetError()
		localctx.SetException(v)
		p.GetErrorHandler().ReportError(p, v)
		p.GetErrorHandler().Recover(p, v)
		p.SetError(nil)
	}
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

func (p *ActionsParser) Sempred(localctx antlr.RuleContext, ruleIndex, predIndex int) bool {
	switch ruleIndex {
	case 1:
		var t *ExprContext = nil
		if localctx != nil {
			t = localctx.(*ExprContext)
		}
		return p.Expr_Sempred(t, predIndex)

	case 2:
		var t *ExprAccessContext = nil
		if localctx != nil {
			t = localctx.(*ExprAccessContext)
		}
		return p.ExprAccess_Sempred(t, predIndex)

	default:
		panic("No predicate with index: " + fmt.Sprint(ruleIndex))
	}
}

func (p *ActionsParser) Expr_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 0:
		return p.Precpred(p.GetParserRuleContext(), 5)

	case 1:
		return p.Precpred(p.GetParserRuleContext(), 4)

	case 2:
		return p.Precpred(p.GetParserRuleContext(), 3)

	case 3:
		return p.Precpred(p.GetParserRuleContext(), 2)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}

func (p *ActionsParser) ExprAccess_Sempred(localctx antlr.RuleContext, predIndex int) bool {
	switch predIndex {
	case 4:
		return p.Precpred(p.GetParserRuleContext(), 5)

	case 5:
		return p.Precpred(p.GetParserRuleContext(), 4)

	default:
		panic("No predicate with index: " + fmt.Sprint(predIndex))
	}
}
