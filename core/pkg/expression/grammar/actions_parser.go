/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

// Code generated from ActionsParser.g4 by ANTLR 4.13.1. DO NOT EDIT.

package grammar // ActionsParser
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

func actionsparserParserInit() {
	staticData := &ActionsParserStaticData
	staticData.LiteralNames = []string{
		"", "'${{'", "", "'}}'", "", "'('", "')'", "'['", "']'", "','", "'.'",
		"'=='", "'!='", "'<'", "'<='", "'>'", "'>='", "'&&'", "'||'", "'!'",
		"'*'", "'null'",
	}
	staticData.SymbolicNames = []string{
		"", "EXPRESSION_OPEN", "TEXT", "EXPRESSION_CLOSE", "WS", "LPAREN", "RPAREN",
		"LBRACK", "RBRACK", "COMMA", "DOT", "EQUAL", "NOTEQUAL", "LT", "LTEQ",
		"GT", "GTEQ", "AND", "OR", "NOT", "WILDCARD", "NULL", "BOOLEAN", "INTEGER",
		"FLOAT", "STRING", "IDENTIFIER",
	}
	staticData.RuleNames = []string{
		"template", "text", "placeholder", "expression", "expr", "exprAccess",
		"identifier", "literal",
	}
	staticData.PredictionContextCache = antlr.NewPredictionContextCache()
	staticData.serializedATN = []int32{
		4, 1, 26, 108, 2, 0, 7, 0, 2, 1, 7, 1, 2, 2, 7, 2, 2, 3, 7, 3, 2, 4, 7,
		4, 2, 5, 7, 5, 2, 6, 7, 6, 2, 7, 7, 7, 1, 0, 1, 0, 4, 0, 19, 8, 0, 11,
		0, 12, 0, 20, 1, 0, 1, 0, 1, 1, 4, 1, 26, 8, 1, 11, 1, 12, 1, 27, 1, 2,
		1, 2, 1, 2, 1, 2, 1, 3, 1, 3, 1, 3, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 3, 4,
		42, 8, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1, 4, 1,
		4, 1, 4, 5, 4, 56, 8, 4, 10, 4, 12, 4, 59, 9, 4, 1, 5, 1, 5, 1, 5, 1, 5,
		1, 5, 1, 5, 5, 5, 67, 8, 5, 10, 5, 12, 5, 70, 9, 5, 3, 5, 72, 8, 5, 1,
		5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 3, 5, 81, 8, 5, 1, 5, 1, 5, 1, 5,
		4, 5, 86, 8, 5, 11, 5, 12, 5, 87, 1, 5, 1, 5, 1, 5, 1, 5, 1, 5, 4, 5, 95,
		8, 5, 11, 5, 12, 5, 96, 5, 5, 99, 8, 5, 10, 5, 12, 5, 102, 9, 5, 1, 6,
		1, 6, 1, 7, 1, 7, 1, 7, 0, 2, 8, 10, 8, 0, 2, 4, 6, 8, 10, 12, 14, 0, 4,
		1, 0, 13, 16, 1, 0, 11, 12, 2, 0, 20, 20, 26, 26, 1, 0, 21, 25, 116, 0,
		18, 1, 0, 0, 0, 2, 25, 1, 0, 0, 0, 4, 29, 1, 0, 0, 0, 6, 33, 1, 0, 0, 0,
		8, 41, 1, 0, 0, 0, 10, 80, 1, 0, 0, 0, 12, 103, 1, 0, 0, 0, 14, 105, 1,
		0, 0, 0, 16, 19, 3, 2, 1, 0, 17, 19, 3, 4, 2, 0, 18, 16, 1, 0, 0, 0, 18,
		17, 1, 0, 0, 0, 19, 20, 1, 0, 0, 0, 20, 18, 1, 0, 0, 0, 20, 21, 1, 0, 0,
		0, 21, 22, 1, 0, 0, 0, 22, 23, 5, 0, 0, 1, 23, 1, 1, 0, 0, 0, 24, 26, 5,
		2, 0, 0, 25, 24, 1, 0, 0, 0, 26, 27, 1, 0, 0, 0, 27, 25, 1, 0, 0, 0, 27,
		28, 1, 0, 0, 0, 28, 3, 1, 0, 0, 0, 29, 30, 5, 1, 0, 0, 30, 31, 3, 8, 4,
		0, 31, 32, 5, 3, 0, 0, 32, 5, 1, 0, 0, 0, 33, 34, 3, 8, 4, 0, 34, 35, 5,
		0, 0, 1, 35, 7, 1, 0, 0, 0, 36, 37, 6, 4, -1, 0, 37, 42, 3, 10, 5, 0, 38,
		39, 5, 19, 0, 0, 39, 42, 3, 8, 4, 6, 40, 42, 3, 14, 7, 0, 41, 36, 1, 0,
		0, 0, 41, 38, 1, 0, 0, 0, 41, 40, 1, 0, 0, 0, 42, 57, 1, 0, 0, 0, 43, 44,
		10, 5, 0, 0, 44, 45, 7, 0, 0, 0, 45, 56, 3, 8, 4, 6, 46, 47, 10, 4, 0,
		0, 47, 48, 7, 1, 0, 0, 48, 56, 3, 8, 4, 5, 49, 50, 10, 3, 0, 0, 50, 51,
		5, 17, 0, 0, 51, 56, 3, 8, 4, 4, 52, 53, 10, 2, 0, 0, 53, 54, 5, 18, 0,
		0, 54, 56, 3, 8, 4, 3, 55, 43, 1, 0, 0, 0, 55, 46, 1, 0, 0, 0, 55, 49,
		1, 0, 0, 0, 55, 52, 1, 0, 0, 0, 56, 59, 1, 0, 0, 0, 57, 55, 1, 0, 0, 0,
		57, 58, 1, 0, 0, 0, 58, 9, 1, 0, 0, 0, 59, 57, 1, 0, 0, 0, 60, 61, 6, 5,
		-1, 0, 61, 62, 3, 12, 6, 0, 62, 71, 5, 5, 0, 0, 63, 68, 3, 8, 4, 0, 64,
		65, 5, 9, 0, 0, 65, 67, 3, 8, 4, 0, 66, 64, 1, 0, 0, 0, 67, 70, 1, 0, 0,
		0, 68, 66, 1, 0, 0, 0, 68, 69, 1, 0, 0, 0, 69, 72, 1, 0, 0, 0, 70, 68,
		1, 0, 0, 0, 71, 63, 1, 0, 0, 0, 71, 72, 1, 0, 0, 0, 72, 73, 1, 0, 0, 0,
		73, 74, 5, 6, 0, 0, 74, 81, 1, 0, 0, 0, 75, 76, 5, 5, 0, 0, 76, 77, 3,
		8, 4, 0, 77, 78, 5, 6, 0, 0, 78, 81, 1, 0, 0, 0, 79, 81, 3, 12, 6, 0, 80,
		60, 1, 0, 0, 0, 80, 75, 1, 0, 0, 0, 80, 79, 1, 0, 0, 0, 81, 100, 1, 0,
		0, 0, 82, 85, 10, 5, 0, 0, 83, 84, 5, 10, 0, 0, 84, 86, 7, 2, 0, 0, 85,
		83, 1, 0, 0, 0, 86, 87, 1, 0, 0, 0, 87, 85, 1, 0, 0, 0, 87, 88, 1, 0, 0,
		0, 88, 99, 1, 0, 0, 0, 89, 94, 10, 4, 0, 0, 90, 91, 5, 7, 0, 0, 91, 92,
		3, 8, 4, 0, 92, 93, 5, 8, 0, 0, 93, 95, 1, 0, 0, 0, 94, 90, 1, 0, 0, 0,
		95, 96, 1, 0, 0, 0, 96, 94, 1, 0, 0, 0, 96, 97, 1, 0, 0, 0, 97, 99, 1,
		0, 0, 0, 98, 82, 1, 0, 0, 0, 98, 89, 1, 0, 0, 0, 99, 102, 1, 0, 0, 0, 100,
		98, 1, 0, 0, 0, 100, 101, 1, 0, 0, 0, 101, 11, 1, 0, 0, 0, 102, 100, 1,
		0, 0, 0, 103, 104, 5, 26, 0, 0, 104, 13, 1, 0, 0, 0, 105, 106, 7, 3, 0,
		0, 106, 15, 1, 0, 0, 0, 13, 18, 20, 27, 41, 55, 57, 68, 71, 80, 87, 96,
		98, 100,
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
	staticData.once.Do(actionsparserParserInit)
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
	this.GrammarFileName = "ActionsParser.g4"

	return this
}

// ActionsParser tokens.
const (
	ActionsParserEOF              = antlr.TokenEOF
	ActionsParserEXPRESSION_OPEN  = 1
	ActionsParserTEXT             = 2
	ActionsParserEXPRESSION_CLOSE = 3
	ActionsParserWS               = 4
	ActionsParserLPAREN           = 5
	ActionsParserRPAREN           = 6
	ActionsParserLBRACK           = 7
	ActionsParserRBRACK           = 8
	ActionsParserCOMMA            = 9
	ActionsParserDOT              = 10
	ActionsParserEQUAL            = 11
	ActionsParserNOTEQUAL         = 12
	ActionsParserLT               = 13
	ActionsParserLTEQ             = 14
	ActionsParserGT               = 15
	ActionsParserGTEQ             = 16
	ActionsParserAND              = 17
	ActionsParserOR               = 18
	ActionsParserNOT              = 19
	ActionsParserWILDCARD         = 20
	ActionsParserNULL             = 21
	ActionsParserBOOLEAN          = 22
	ActionsParserINTEGER          = 23
	ActionsParserFLOAT            = 24
	ActionsParserSTRING           = 25
	ActionsParserIDENTIFIER       = 26
)

// ActionsParser rules.
const (
	ActionsParserRULE_template    = 0
	ActionsParserRULE_text        = 1
	ActionsParserRULE_placeholder = 2
	ActionsParserRULE_expression  = 3
	ActionsParserRULE_expr        = 4
	ActionsParserRULE_exprAccess  = 5
	ActionsParserRULE_identifier  = 6
	ActionsParserRULE_literal     = 7
)

// ITemplateContext is an interface to support dynamic dispatch.
type ITemplateContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EOF() antlr.TerminalNode
	AllText() []ITextContext
	Text(i int) ITextContext
	AllPlaceholder() []IPlaceholderContext
	Placeholder(i int) IPlaceholderContext

	// IsTemplateContext differentiates from other interfaces.
	IsTemplateContext()
}

type TemplateContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTemplateContext() *TemplateContext {
	var p = new(TemplateContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_template
	return p
}

func InitEmptyTemplateContext(p *TemplateContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_template
}

func (*TemplateContext) IsTemplateContext() {}

func NewTemplateContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TemplateContext {
	var p = new(TemplateContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionsParserRULE_template

	return p
}

func (s *TemplateContext) GetParser() antlr.Parser { return s.parser }

func (s *TemplateContext) EOF() antlr.TerminalNode {
	return s.GetToken(ActionsParserEOF, 0)
}

func (s *TemplateContext) AllText() []ITextContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(ITextContext); ok {
			len++
		}
	}

	tst := make([]ITextContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(ITextContext); ok {
			tst[i] = t.(ITextContext)
			i++
		}
	}

	return tst
}

func (s *TemplateContext) Text(i int) ITextContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(ITextContext); ok {
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

	return t.(ITextContext)
}

func (s *TemplateContext) AllPlaceholder() []IPlaceholderContext {
	children := s.GetChildren()
	len := 0
	for _, ctx := range children {
		if _, ok := ctx.(IPlaceholderContext); ok {
			len++
		}
	}

	tst := make([]IPlaceholderContext, len)
	i := 0
	for _, ctx := range children {
		if t, ok := ctx.(IPlaceholderContext); ok {
			tst[i] = t.(IPlaceholderContext)
			i++
		}
	}

	return tst
}

func (s *TemplateContext) Placeholder(i int) IPlaceholderContext {
	var t antlr.RuleContext
	j := 0
	for _, ctx := range s.GetChildren() {
		if _, ok := ctx.(IPlaceholderContext); ok {
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

	return t.(IPlaceholderContext)
}

func (s *TemplateContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TemplateContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TemplateContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterTemplate(s)
	}
}

func (s *TemplateContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitTemplate(s)
	}
}

func (s *TemplateContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitTemplate(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionsParser) Template() (localctx ITemplateContext) {
	localctx = NewTemplateContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 0, ActionsParserRULE_template)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(18)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_la = p.GetTokenStream().LA(1)

	for ok := true; ok; ok = _la == ActionsParserEXPRESSION_OPEN || _la == ActionsParserTEXT {
		p.SetState(18)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}

		switch p.GetTokenStream().LA(1) {
		case ActionsParserTEXT:
			{
				p.SetState(16)
				p.Text()
			}

		case ActionsParserEXPRESSION_OPEN:
			{
				p.SetState(17)
				p.Placeholder()
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(20)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)
	}
	{
		p.SetState(22)
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

// ITextContext is an interface to support dynamic dispatch.
type ITextContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	AllTEXT() []antlr.TerminalNode
	TEXT(i int) antlr.TerminalNode

	// IsTextContext differentiates from other interfaces.
	IsTextContext()
}

type TextContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyTextContext() *TextContext {
	var p = new(TextContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_text
	return p
}

func InitEmptyTextContext(p *TextContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_text
}

func (*TextContext) IsTextContext() {}

func NewTextContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *TextContext {
	var p = new(TextContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionsParserRULE_text

	return p
}

func (s *TextContext) GetParser() antlr.Parser { return s.parser }

func (s *TextContext) AllTEXT() []antlr.TerminalNode {
	return s.GetTokens(ActionsParserTEXT)
}

func (s *TextContext) TEXT(i int) antlr.TerminalNode {
	return s.GetToken(ActionsParserTEXT, i)
}

func (s *TextContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *TextContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *TextContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterText(s)
	}
}

func (s *TextContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitText(s)
	}
}

func (s *TextContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitText(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionsParser) Text() (localctx ITextContext) {
	localctx = NewTextContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 2, ActionsParserRULE_text)
	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(25)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = 1
	for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		switch _alt {
		case 1:
			{
				p.SetState(24)
				p.Match(ActionsParserTEXT)
				if p.HasError() {
					// Recognition error - abort rule
					goto errorExit
				}
			}

		default:
			p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
			goto errorExit
		}

		p.SetState(27)
		p.GetErrorHandler().Sync(p)
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
	p.ExitRule()
	return localctx
	goto errorExit // Trick to prevent compiler error if the label is not used
}

// IPlaceholderContext is an interface to support dynamic dispatch.
type IPlaceholderContext interface {
	antlr.ParserRuleContext

	// GetParser returns the parser.
	GetParser() antlr.Parser

	// Getter signatures
	EXPRESSION_OPEN() antlr.TerminalNode
	Expr() IExprContext
	EXPRESSION_CLOSE() antlr.TerminalNode

	// IsPlaceholderContext differentiates from other interfaces.
	IsPlaceholderContext()
}

type PlaceholderContext struct {
	antlr.BaseParserRuleContext
	parser antlr.Parser
}

func NewEmptyPlaceholderContext() *PlaceholderContext {
	var p = new(PlaceholderContext)
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_placeholder
	return p
}

func InitEmptyPlaceholderContext(p *PlaceholderContext) {
	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, nil, -1)
	p.RuleIndex = ActionsParserRULE_placeholder
}

func (*PlaceholderContext) IsPlaceholderContext() {}

func NewPlaceholderContext(parser antlr.Parser, parent antlr.ParserRuleContext, invokingState int) *PlaceholderContext {
	var p = new(PlaceholderContext)

	antlr.InitBaseParserRuleContext(&p.BaseParserRuleContext, parent, invokingState)

	p.parser = parser
	p.RuleIndex = ActionsParserRULE_placeholder

	return p
}

func (s *PlaceholderContext) GetParser() antlr.Parser { return s.parser }

func (s *PlaceholderContext) EXPRESSION_OPEN() antlr.TerminalNode {
	return s.GetToken(ActionsParserEXPRESSION_OPEN, 0)
}

func (s *PlaceholderContext) Expr() IExprContext {
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

func (s *PlaceholderContext) EXPRESSION_CLOSE() antlr.TerminalNode {
	return s.GetToken(ActionsParserEXPRESSION_CLOSE, 0)
}

func (s *PlaceholderContext) GetRuleContext() antlr.RuleContext {
	return s
}

func (s *PlaceholderContext) ToStringTree(ruleNames []string, recog antlr.Recognizer) string {
	return antlr.TreesStringTree(s, ruleNames, recog)
}

func (s *PlaceholderContext) EnterRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.EnterPlaceholder(s)
	}
}

func (s *PlaceholderContext) ExitRule(listener antlr.ParseTreeListener) {
	if listenerT, ok := listener.(ActionsListener); ok {
		listenerT.ExitPlaceholder(s)
	}
}

func (s *PlaceholderContext) Accept(visitor antlr.ParseTreeVisitor) any {
	switch t := visitor.(type) {
	case ActionsVisitor:
		return t.VisitPlaceholder(s)

	default:
		return t.VisitChildren(s)
	}
}

func (p *ActionsParser) Placeholder() (localctx IPlaceholderContext) {
	localctx = NewPlaceholderContext(p, p.GetParserRuleContext(), p.GetState())
	p.EnterRule(localctx, 4, ActionsParserRULE_placeholder)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(29)
		p.Match(ActionsParserEXPRESSION_OPEN)
		if p.HasError() {
			// Recognition error - abort rule
			goto errorExit
		}
	}
	{
		p.SetState(30)
		p.expr(0)
	}
	{
		p.SetState(31)
		p.Match(ActionsParserEXPRESSION_CLOSE)
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
	p.EnterRule(localctx, 6, ActionsParserRULE_expression)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(33)

		var _x = p.expr(0)

		localctx.(*ExpressionContext).e = _x
	}
	{
		p.SetState(34)
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
	_startState := 8
	p.EnterRecursionRule(localctx, 8, ActionsParserRULE_expr, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(41)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetTokenStream().LA(1) {
	case ActionsParserLPAREN, ActionsParserIDENTIFIER:
		{
			p.SetState(37)
			p.exprAccess(0)
		}

	case ActionsParserNOT:
		{
			p.SetState(38)

			var _m = p.Match(ActionsParserNOT)

			localctx.(*ExprContext).op = _m
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(39)
			p.expr(6)
		}

	case ActionsParserNULL, ActionsParserBOOLEAN, ActionsParserINTEGER, ActionsParserFLOAT, ActionsParserSTRING:
		{
			p.SetState(40)
			p.Literal()
		}

	default:
		p.SetError(antlr.NewNoViableAltException(p, nil, nil, nil, nil, nil))
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(57)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(55)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 4, p.GetParserRuleContext()) {
			case 1:
				localctx = NewExprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_expr)
				p.SetState(43)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				{
					p.SetState(44)

					var _lt = p.GetTokenStream().LT(1)

					localctx.(*ExprContext).op = _lt

					_la = p.GetTokenStream().LA(1)

					if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&122880) != 0) {
						var _ri = p.GetErrorHandler().RecoverInline(p)

						localctx.(*ExprContext).op = _ri
					} else {
						p.GetErrorHandler().ReportMatch(p)
						p.Consume()
					}
				}
				{
					p.SetState(45)
					p.expr(6)
				}

			case 2:
				localctx = NewExprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_expr)
				p.SetState(46)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				{
					p.SetState(47)

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
					p.SetState(48)
					p.expr(5)
				}

			case 3:
				localctx = NewExprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_expr)
				p.SetState(49)

				if !(p.Precpred(p.GetParserRuleContext(), 3)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 3)", ""))
					goto errorExit
				}
				{
					p.SetState(50)

					var _m = p.Match(ActionsParserAND)

					localctx.(*ExprContext).op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(51)
					p.expr(4)
				}

			case 4:
				localctx = NewExprContext(p, _parentctx, _parentState)
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_expr)
				p.SetState(52)

				if !(p.Precpred(p.GetParserRuleContext(), 2)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 2)", ""))
					goto errorExit
				}
				{
					p.SetState(53)

					var _m = p.Match(ActionsParserOR)

					localctx.(*ExprContext).op = _m
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(54)
					p.expr(3)
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(59)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 5, p.GetParserRuleContext())
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
	_startState := 10
	p.EnterRecursionRule(localctx, 10, ActionsParserRULE_exprAccess, _p)
	var _la int

	var _alt int

	p.EnterOuterAlt(localctx, 1)
	p.SetState(80)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}

	switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 8, p.GetParserRuleContext()) {
	case 1:
		localctx = NewFunctionCallContext(p, localctx)
		p.SetParserRuleContext(localctx)
		_prevctx = localctx

		{
			p.SetState(61)
			p.Identifier()
		}
		{
			p.SetState(62)
			p.Match(ActionsParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		p.SetState(71)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_la = p.GetTokenStream().LA(1)

		if (int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&132644896) != 0 {
			{
				p.SetState(63)

				var _x = p.expr(0)

				localctx.(*FunctionCallContext)._expr = _x
			}
			localctx.(*FunctionCallContext).args = append(localctx.(*FunctionCallContext).args, localctx.(*FunctionCallContext)._expr)
			p.SetState(68)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}
			_la = p.GetTokenStream().LA(1)

			for _la == ActionsParserCOMMA {
				{
					p.SetState(64)
					p.Match(ActionsParserCOMMA)
					if p.HasError() {
						// Recognition error - abort rule
						goto errorExit
					}
				}
				{
					p.SetState(65)

					var _x = p.expr(0)

					localctx.(*FunctionCallContext)._expr = _x
				}
				localctx.(*FunctionCallContext).args = append(localctx.(*FunctionCallContext).args, localctx.(*FunctionCallContext)._expr)

				p.SetState(70)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_la = p.GetTokenStream().LA(1)
			}

		}
		{
			p.SetState(73)
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
			p.SetState(75)
			p.Match(ActionsParserLPAREN)
			if p.HasError() {
				// Recognition error - abort rule
				goto errorExit
			}
		}
		{
			p.SetState(76)
			p.expr(0)
		}
		{
			p.SetState(77)
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
			p.SetState(79)
			p.Identifier()
		}

	case antlr.ATNInvalidAltNumber:
		goto errorExit
	}
	p.GetParserRuleContext().SetStop(p.GetTokenStream().LT(-1))
	p.SetState(100)
	p.GetErrorHandler().Sync(p)
	if p.HasError() {
		goto errorExit
	}
	_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext())
	if p.HasError() {
		goto errorExit
	}
	for _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
		if _alt == 1 {
			if p.GetParseListeners() != nil {
				p.TriggerExitRuleEvent()
			}
			_prevctx = localctx
			p.SetState(98)
			p.GetErrorHandler().Sync(p)
			if p.HasError() {
				goto errorExit
			}

			switch p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 11, p.GetParserRuleContext()) {
			case 1:
				localctx = NewPropertyAccessContext(p, NewExprAccessContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_exprAccess)
				p.SetState(82)

				if !(p.Precpred(p.GetParserRuleContext(), 5)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 5)", ""))
					goto errorExit
				}
				p.SetState(85)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = 1
				for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1:
						{
							p.SetState(83)
							p.Match(ActionsParserDOT)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}
						{
							p.SetState(84)

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

					p.SetState(87)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 9, p.GetParserRuleContext())
					if p.HasError() {
						goto errorExit
					}
				}

			case 2:
				localctx = NewIndexAccessContext(p, NewExprAccessContext(p, _parentctx, _parentState))
				p.PushNewRecursionContext(localctx, _startState, ActionsParserRULE_exprAccess)
				p.SetState(89)

				if !(p.Precpred(p.GetParserRuleContext(), 4)) {
					p.SetError(antlr.NewFailedPredicateException(p, "p.Precpred(p.GetParserRuleContext(), 4)", ""))
					goto errorExit
				}
				p.SetState(94)
				p.GetErrorHandler().Sync(p)
				if p.HasError() {
					goto errorExit
				}
				_alt = 1
				for ok := true; ok; ok = _alt != 2 && _alt != antlr.ATNInvalidAltNumber {
					switch _alt {
					case 1:
						{
							p.SetState(90)
							p.Match(ActionsParserLBRACK)
							if p.HasError() {
								// Recognition error - abort rule
								goto errorExit
							}
						}
						{
							p.SetState(91)

							var _x = p.expr(0)

							localctx.(*IndexAccessContext)._expr = _x
						}
						localctx.(*IndexAccessContext).indexes = append(localctx.(*IndexAccessContext).indexes, localctx.(*IndexAccessContext)._expr)
						{
							p.SetState(92)
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

					p.SetState(96)
					p.GetErrorHandler().Sync(p)
					_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 10, p.GetParserRuleContext())
					if p.HasError() {
						goto errorExit
					}
				}

			case antlr.ATNInvalidAltNumber:
				goto errorExit
			}

		}
		p.SetState(102)
		p.GetErrorHandler().Sync(p)
		if p.HasError() {
			goto errorExit
		}
		_alt = p.GetInterpreter().AdaptivePredict(p.BaseParser, p.GetTokenStream(), 12, p.GetParserRuleContext())
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
	p.EnterRule(localctx, 12, ActionsParserRULE_identifier)
	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(103)
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
	p.EnterRule(localctx, 14, ActionsParserRULE_literal)
	var _la int

	p.EnterOuterAlt(localctx, 1)
	{
		p.SetState(105)
		_la = p.GetTokenStream().LA(1)

		if !((int64(_la) & ^0x3f) == 0 && ((int64(1)<<_la)&65011712) != 0) {
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
	case 4:
		var t *ExprContext = nil
		if localctx != nil {
			t = localctx.(*ExprContext)
		}
		return p.Expr_Sempred(t, predIndex)

	case 5:
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
