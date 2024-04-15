package parser

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/dungdm93/drasi/pkg/parser/constants"
)

type Lexer struct {
	expression     string
	index          int
	lastToken      *LexicalToken
	unclosedTokens []*LexicalToken
}

func (l *Lexer) TryGetNextToken() (result *LexicalToken, haveResult bool) {
	runes := []rune(l.expression)
	for l.index < len(l.expression) && unicode.IsSpace(runes[l.index]) {
		l.index++
	}
	// testcase end of string
	if l.index >= len(l.expression) {
		return nil, false
	}

	// read the first character to determine the type of token.
	c := string(l.expression[l.index])
	switch c {
	case constants.StartGroup:
		// function call
		if l.lastToken != nil && l.lastToken.kind == LTKFunction {
			result = l.createToken(LTKStartParameters, string(c), l.index, nil)
		} else {
			// logical grouping
			result = l.createToken(LTKStartGroup, c, l.index, nil)
		}
		l.index++
	case constants.StartIndex:
		result = l.createToken(LTKStartIndex, c, l.index, nil)
		l.index++
	case constants.EndGroup:
		// function call
		if len(l.unclosedTokens) > 0 && l.unclosedTokens[len(l.unclosedTokens)-1].kind == LTKStartParameters {
			result = l.createToken(LTKEndParameters, string(c), l.index, nil)
		} else {
			// logical grouping
			result = l.createToken(LTKEndGroup, c, l.index, nil)
		}
		l.index++
	case constants.EndIndex:
		result = l.createToken(LTKEndIndex, string(c), l.index, nil)
		l.index++
	case constants.Separator:
		result = l.createToken(LTKSeparator, string(c), l.index, nil)
		l.index++
	case constants.Wildcard:
		result = l.createToken(LTKWildcard, string(c), l.index, nil)
		l.index++
	case "'":
		result = l.readStringToken()
	case "!", ">", "<", "=", "&", "|":
		// also catch != >= <= == && ||
		result = l.readOperator()
	default:
		if c == "." {
			// number
			if l.lastToken == nil || l.lastToken.kind == LTKSeparator || l.lastToken.kind == LTKStartGroup || l.lastToken.kind == LTKStartIndex || l.lastToken.kind == LTKStartParameters || l.lastToken.kind == LTKLogicalOperator {
				result = l.readNumberToken()
			} else {
				// .
				result = l.createToken(LTKDereference, c, l.index, nil)
				l.index++
			}
		} else if c == "-" || c == "+" || (c >= "0" && c <= "9") {
			result = l.readNumberToken()
		} else {
			result = l.readKeywordToken()
		}
	}
	l.lastToken = result
	return result, true
}

func (l *Lexer) UnclosedTokens() []*LexicalToken {
	return l.unclosedTokens
}

func (l *Lexer) readKeywordToken() *LexicalToken {
	runes := []rune(l.expression)
	startIndex := l.index
	l.index++

	for l.index < len(l.expression) && !testTokenBoundary(runes[l.index]) {
		l.index++
	}
	str := l.expression[startIndex:l.index]
	if IsLegalKeyWord(str) {
		// Test if follows property dereference operator.
		if l.lastToken != nil && l.lastToken.kind == LTKDereference {
			return l.createToken(LTKPropertyName, str, startIndex, nil)
		}
		// LTKNull
		if strings.EqualFold(str, constants.Null) {
			return l.createToken(LTKNull, str, startIndex, nil)
		}
		// LTKBoolean
		if strings.EqualFold(str, constants.True) {
			return l.createToken(LTKBoolean, str, startIndex, true)
		}
		if strings.EqualFold(str, constants.False) {
			return l.createToken(LTKBoolean, str, startIndex, false)
		}
		// NaN
		if strings.EqualFold(str, constants.NaN) {
			return l.createToken(LTKNumber, str, startIndex, math.NaN())
		}
		// Infinity
		if strings.EqualFold(str, constants.Infinity) {
			return l.createToken(LTKNumber, str, startIndex, math.Inf(1))
		}
		// Lookahead
		tmpIndex := l.index
		for tmpIndex < len(runes) && unicode.IsSpace(runes[tmpIndex]) {
			tmpIndex++
		}
		// Fn. Eg: success(), always()
		if tmpIndex < len(l.expression) && string(l.expression[tmpIndex]) == constants.StartGroup {
			return l.createToken(LTKFunction, str, startIndex, nil)
		} else {
			// Named values. Eg github
			return l.createToken(LTKNamedValue, str, startIndex, nil)
		}
	} else {
		return l.createToken(LTKUnexpected, str, startIndex, nil)
	}
}

func (l *Lexer) readNumberToken() *LexicalToken {
	start := l.index
	for {
		l.index++
		if l.index >= len(l.expression) || (testTokenBoundary(rune(l.expression[l.index])) && l.expression[l.
			index] != '.') {
			break
		}
	}
	str := l.expression[start:l.index]
	d := ParseNumber(str)
	if math.IsNaN(d) {
		return l.createToken(LTKUnexpected, str, start, nil)
	}
	return l.createToken(LTKNumber, str, start, d)
}

func (l *Lexer) readStringToken() *LexicalToken {
	start := l.index
	var closed bool
	var str strings.Builder
	l.index++
	for l.index < len(l.expression) {
		c := fmt.Sprintf("%c", l.expression[l.index])
		// move to next char
		l.index++
		if c == ("'") {
			// End of string
			if l.index >= len(l.expression) || fmt.Sprintf("%c", l.expression[l.index]) != "'" {
				closed = true
				break
			}
			// Escaped single quote.
			// Example: ${{ 'It''s open source!' }}
			l.index++
		}
		_, err := str.WriteString(c)
		if err != nil {
			panic(err)
		}
	}
	rawValue := l.expression[start:l.index]
	if closed {
		return l.createToken(LTKString, rawValue, start, str.String())
	}
	return l.createToken(LTKUnexpected, rawValue, start, nil)
}

func (l *Lexer) readOperator() *LexicalToken {
	start := l.index
	// skip first char since we already knows what it is
	l.index++
	// check for 2 characters operator
	if l.index < len(l.expression) {
		// increase index, in case this is a valid 2 characters operator. We remember that this was read.
		l.index++
		raw := l.expression[start:l.index]
		switch raw {
		case constants.NotEqual, constants.GreaterThanOrEqual, constants.LessThanOrEqual, constants.Equal,
			constants.And, constants.Or:
			return l.createToken(LTKLogicalOperator, raw, start, nil)
		}
		l.index--
	}

	// check for one-character operator
	raw := l.expression[start:l.index]
	switch raw {
	case constants.Not, constants.GreaterThan, constants.LessThan:
		return l.createToken(LTKLogicalOperator, raw, start, nil)
	}
	// unexpected
	for l.index < len(l.expression) && !testTokenBoundary(rune(l.expression[l.index])) {
		l.index++
	}
	return l.createToken(LTKUnexpected, l.expression[start:l.index], start, nil)
}

// createToken performs valid check based on last token stored in lexer, return a new LexicalToken if condition check is passed
func (l *Lexer) createToken(kind LexicalTokenKind, rawValue string, startIndex int, parsedValue any) *LexicalToken {
	var legal bool
	switch kind {
	case LTKStartGroup:
		legal = l.checkLastToken(LTKNotInitialized, LTKSeparator, LTKStartGroup, LTKStartParameters, LTKStartIndex, LTKLogicalOperator)
	case LTKStartIndex:
		legal = l.checkLastToken(LTKEndGroup, LTKEndParameters, LTKEndIndex, LTKWildcard,
			LTKPropertyName, LTKNamedValue)
	case LTKStartParameters:
		legal = l.checkLastToken(LTKFunction)
	case LTKEndGroup:
		legal = l.checkLastToken(LTKEndGroup, LTKEndParameters, LTKEndIndex, LTKWildcard,
			LTKNull, LTKBoolean, LTKNumber, LTKString, LTKPropertyName, LTKNamedValue)
	case LTKEndIndex:
		legal = l.checkLastToken(LTKEndGroup, LTKEndParameters, LTKEndIndex, LTKWildcard, LTKNull, LTKBoolean, LTKNumber, LTKString, LTKPropertyName, LTKNamedValue)
	case LTKEndParameters:
		legal = l.checkLastToken(LTKStartParameters, LTKEndGroup, LTKEndParameters, LTKEndIndex, LTKWildcard, LTKNull, LTKBoolean, LTKNumber, LTKString, LTKPropertyName, LTKNamedValue)
		break
	case LTKSeparator:
		legal = l.checkLastToken(LTKEndGroup, LTKEndParameters, LTKEndIndex, LTKWildcard, LTKNull, LTKBoolean, LTKNumber, LTKString, LTKPropertyName, LTKNamedValue)
	case LTKDereference:
		legal = l.checkLastToken(LTKEndGroup, LTKEndParameters, LTKEndIndex, LTKWildcard, LTKPropertyName, LTKNamedValue)
	case LTKWildcard:
		legal = l.checkLastToken(LTKStartIndex, LTKDereference)
	case LTKLogicalOperator:
		if rawValue == constants.Not {
			legal = l.checkLastToken(LTKNotInitialized, LTKSeparator, LTKStartGroup, LTKStartParameters, LTKStartIndex, LTKLogicalOperator)
		} else {
			legal = l.checkLastToken(LTKEndGroup, LTKEndParameters, LTKEndIndex, LTKWildcard, LTKNull, LTKBoolean, LTKNumber, LTKString, LTKPropertyName, LTKNamedValue)
		}
	case LTKPropertyName:
		legal = l.checkLastToken(LTKDereference)
	case LTKNull, LTKBoolean, LTKNumber, LTKString, LTKFunction, LTKNamedValue:
		legal = l.checkLastToken(LTKNotInitialized, LTKSeparator, LTKStartIndex, LTKStartGroup, LTKStartParameters, LTKLogicalOperator)
	}
	// Illegal
	if !legal {
		return &LexicalToken{
			kind:     LTKUnexpected,
			rawValue: rawValue,
			index:    startIndex,
		}
	}

	// Legal so far
	token := &LexicalToken{
		kind:        kind,
		rawValue:    rawValue,
		index:       startIndex,
		parsedValue: parsedValue,
	}
	switch kind {
	case LTKStartGroup, LTKStartIndex, LTKStartParameters:
		// Track start token
		l.unclosedTokens = append(l.unclosedTokens, token)
	case LTKEndGroup:
		// Check inside logical grouping
		if l.lastUnclosedKind() != LTKStartGroup {
			return &LexicalToken{
				kind:     LTKUnexpected,
				rawValue: rawValue,
				index:    startIndex,
			}
		}
		// remove last start group
		l.unclosedTokens = l.unclosedTokens[:len(l.unclosedTokens)-1]
	case LTKEndIndex:
		// Check inside indexer
		if l.lastUnclosedKind() != LTKStartIndex {
			return &LexicalToken{
				kind:     LTKUnexpected,
				rawValue: rawValue,
				index:    startIndex,
			}
		}
		// remove last start startIndex
		l.unclosedTokens = l.unclosedTokens[:len(l.unclosedTokens)-1]
	case LTKEndParameters:
		// Check inside function call
		if l.lastUnclosedKind() != LTKStartParameters {
			return &LexicalToken{
				kind:     LTKUnexpected,
				rawValue: rawValue,
				index:    startIndex,
			}
		}
		// remove last start parameter
		l.unclosedTokens = l.unclosedTokens[:len(l.unclosedTokens)-1]
	case LTKSeparator: // ","
		// Check inside function call
		if l.lastUnclosedKind() != LTKStartParameters {
			return &LexicalToken{
				kind:     LTKUnexpected,
				rawValue: rawValue,
				index:    startIndex,
			}
		}
	}
	return token
}

func (l *Lexer) lastUnclosedKind() (kind LexicalTokenKind) {
	if len(l.unclosedTokens) > 0 {
		kind = l.unclosedTokens[len(l.unclosedTokens)-1].kind
	}
	return
}

// checkLastToken check if current token is valid based on it precedence token's kind
func (l *Lexer) checkLastToken(allowedLastKinds ...LexicalTokenKind) (valid bool) {
	var currentLastKind LexicalTokenKind
	if l.lastToken == nil {
		currentLastKind = LTKNotInitialized
	} else {
		currentLastKind = l.lastToken.kind
	}
	for _, validLastKind := range allowedLastKinds {
		if validLastKind == currentLastKind {
			return true
		}
	}
	return
}

func testTokenBoundary(c rune) bool {
	switch c {
	case '(', '[', ')', ']', ',', '.', '!', '>', '<', '=', '&', '|':
		return true
	default:
		return unicode.IsSpace(c)
	}
}

func NewLexer(expression string) *Lexer {
	return &Lexer{expression: expression}
}
