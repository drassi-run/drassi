package parser

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/dungdm93/drasi/pkg/expression"
)

type lexer struct {
	expression     string
	index          int
	lastToken      *lexicalToken
	unclosedTokens []*lexicalToken
}

func (l *lexer) tryGetNextToken() (result *lexicalToken, haveToken bool) {
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
	case expression.StartGroup:
		// function call
		if l.lastToken != nil && l.lastToken.kind == lexicalTokenKindFunction {
			result = l.createToken(lexicalTokenKindStartParameters, c, l.index, nil)
		} else {
			// logical grouping
			result = l.createToken(lexicalTokenKindStartGroup, c, l.index, nil)
		}
		l.index++
	case expression.StartIndex:
		result = l.createToken(lexicalTokenKindStartIndex, c, l.index, nil)
		l.index++
	case expression.EndGroup:
		// function call
		if len(l.unclosedTokens) > 0 && l.unclosedTokens[len(l.unclosedTokens)-1].kind == lexicalTokenKindStartParameters {
			result = l.createToken(lexicalTokenKindEndParameters, c, l.index, nil)
		} else {
			// logical grouping
			result = l.createToken(lexicalTokenKindEndGroup, c, l.index, nil)
		}
		l.index++
	case expression.EndIndex:
		result = l.createToken(lexicalTokenKindEndIndex, c, l.index, nil)
		l.index++
	case expression.Separator:
		result = l.createToken(lexicalTokenKindSeparator, c, l.index, nil)
		l.index++
	case expression.Wildcard:
		result = l.createToken(lexicalTokenKindWildcard, c, l.index, nil)
		l.index++
	case "'":
		result = l.readStringToken()
	case "!", ">", "<", "=", "&", "|":
		// also catch != >= <= == && ||
		result = l.readOperator()
	default:
		if c == "." {
			// number
			if l.lastToken == nil || l.lastToken.kind == lexicalTokenKindSeparator || l.lastToken.kind == lexicalTokenKindStartGroup || l.lastToken.kind == lexicalTokenKindStartIndex || l.lastToken.kind == lexicalTokenKindStartParameters || l.lastToken.kind == lexicalTokenKindLogicalOperator {
				result = l.readNumberToken()
			} else {
				// .
				result = l.createToken(lexicalTokenKindDereference, c, l.index, nil)
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

func (l *lexer) getUnclosedTokens() []*lexicalToken {
	return l.unclosedTokens
}

func (l *lexer) readKeywordToken() *lexicalToken {
	runes := []rune(l.expression)
	startIndex := l.index
	l.index++

	for l.index < len(l.expression) && !testTokenBoundary(runes[l.index]) {
		l.index++
	}
	str := l.expression[startIndex:l.index]
	if isLegalKeyWord(str) {
		// Test if follows property dereference operator.
		if l.lastToken != nil && l.lastToken.kind == lexicalTokenKindDereference {
			return l.createToken(lexicalTokenKindPropertyName, str, startIndex, nil)
		}
		// lexicalTokenKindNull
		if strings.EqualFold(str, expression.Null) {
			return l.createToken(lexicalTokenKindNull, str, startIndex, nil)
		}
		// lexicalTokenKindBoolean
		if strings.EqualFold(str, expression.True) {
			return l.createToken(lexicalTokenKindBoolean, str, startIndex, true)
		}
		if strings.EqualFold(str, expression.False) {
			return l.createToken(lexicalTokenKindBoolean, str, startIndex, false)
		}
		// NaN
		if strings.EqualFold(str, expression.NaN) {
			return l.createToken(lexicalTokenKindNumber, str, startIndex, math.NaN())
		}
		// Infinity
		if strings.EqualFold(str, expression.Infinity) {
			return l.createToken(lexicalTokenKindNumber, str, startIndex, math.Inf(1))
		}
		// Lookahead
		tmpIndex := l.index
		for tmpIndex < len(runes) && unicode.IsSpace(runes[tmpIndex]) {
			tmpIndex++
		}
		// Fn. Eg: success(), always()
		if tmpIndex < len(l.expression) && string(l.expression[tmpIndex]) == expression.StartGroup {
			return l.createToken(lexicalTokenKindFunction, str, startIndex, nil)
		} else {
			// Named values. Eg github
			return l.createToken(lexicalTokenKindNamedValue, str, startIndex, nil)
		}
	} else {
		return l.createToken(lexicalTokenKindUnexpected, str, startIndex, nil)
	}
}

func (l *lexer) readNumberToken() *lexicalToken {
	start := l.index
	for {
		l.index++
		if l.index >= len(l.expression) || (testTokenBoundary(rune(l.expression[l.index])) && l.expression[l.
			index] != '.') {
			break
		}
	}
	str := l.expression[start:l.index]
	d := parseNumber(str)
	if math.IsNaN(d) {
		return l.createToken(lexicalTokenKindUnexpected, str, start, nil)
	}
	return l.createToken(lexicalTokenKindNumber, str, start, d)
}

func (l *lexer) readStringToken() *lexicalToken {
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
		return l.createToken(lexicalTokenKindString, rawValue, start, str.String())
	}
	return l.createToken(lexicalTokenKindUnexpected, rawValue, start, nil)
}

func (l *lexer) readOperator() *lexicalToken {
	start := l.index
	// skip first char since we already knows what it is
	l.index++
	// check for 2 characters operator
	if l.index < len(l.expression) {
		// increase index, in case this is a valid 2 characters operator. We remember that this was read.
		l.index++
		raw := l.expression[start:l.index]
		switch raw {
		case expression.NotEqual, expression.GreaterThanOrEqual, expression.LessThanOrEqual, expression.Equal,
			expression.And, expression.Or:
			return l.createToken(lexicalTokenKindLogicalOperator, raw, start, nil)
		}
		l.index--
	}

	// check for one-character operator
	raw := l.expression[start:l.index]
	switch raw {
	case expression.Not, expression.GreaterThan, expression.LessThan:
		return l.createToken(lexicalTokenKindLogicalOperator, raw, start, nil)
	}
	// unexpected
	for l.index < len(l.expression) && !testTokenBoundary(rune(l.expression[l.index])) {
		l.index++
	}
	return l.createToken(lexicalTokenKindUnexpected, l.expression[start:l.index], start, nil)
}

// createToken performs valid check based on last token stored in lexer, return a new lexicalToken if condition check is passed
func (l *lexer) createToken(kind lexicalTokenKind, rawValue string, startIndex int, parsedValue any) *lexicalToken {
	var legal bool
	switch kind {
	case lexicalTokenKindStartGroup:
		legal = l.checkLastToken(lexicalTokenKindNotInitialized, lexicalTokenKindSeparator, lexicalTokenKindStartGroup, lexicalTokenKindStartParameters, lexicalTokenKindStartIndex, lexicalTokenKindLogicalOperator)
	case lexicalTokenKindStartIndex:
		legal = l.checkLastToken(lexicalTokenKindEndGroup, lexicalTokenKindEndParameters, lexicalTokenKindEndIndex, lexicalTokenKindWildcard,
			lexicalTokenKindPropertyName, lexicalTokenKindNamedValue)
	case lexicalTokenKindStartParameters:
		legal = l.checkLastToken(lexicalTokenKindFunction)
	case lexicalTokenKindEndGroup:
		legal = l.checkLastToken(lexicalTokenKindEndGroup, lexicalTokenKindEndParameters, lexicalTokenKindEndIndex, lexicalTokenKindWildcard,
			lexicalTokenKindNull, lexicalTokenKindBoolean, lexicalTokenKindNumber, lexicalTokenKindString, lexicalTokenKindPropertyName, lexicalTokenKindNamedValue)
	case lexicalTokenKindEndIndex:
		legal = l.checkLastToken(lexicalTokenKindEndGroup, lexicalTokenKindEndParameters, lexicalTokenKindEndIndex, lexicalTokenKindWildcard, lexicalTokenKindNull, lexicalTokenKindBoolean, lexicalTokenKindNumber, lexicalTokenKindString, lexicalTokenKindPropertyName, lexicalTokenKindNamedValue)
	case lexicalTokenKindEndParameters:
		legal = l.checkLastToken(lexicalTokenKindStartParameters, lexicalTokenKindEndGroup, lexicalTokenKindEndParameters, lexicalTokenKindEndIndex, lexicalTokenKindWildcard, lexicalTokenKindNull, lexicalTokenKindBoolean, lexicalTokenKindNumber, lexicalTokenKindString, lexicalTokenKindPropertyName, lexicalTokenKindNamedValue)
		break
	case lexicalTokenKindSeparator:
		legal = l.checkLastToken(lexicalTokenKindEndGroup, lexicalTokenKindEndParameters, lexicalTokenKindEndIndex, lexicalTokenKindWildcard, lexicalTokenKindNull, lexicalTokenKindBoolean, lexicalTokenKindNumber, lexicalTokenKindString, lexicalTokenKindPropertyName, lexicalTokenKindNamedValue)
	case lexicalTokenKindDereference:
		legal = l.checkLastToken(lexicalTokenKindEndGroup, lexicalTokenKindEndParameters, lexicalTokenKindEndIndex, lexicalTokenKindWildcard, lexicalTokenKindPropertyName, lexicalTokenKindNamedValue)
	case lexicalTokenKindWildcard:
		legal = l.checkLastToken(lexicalTokenKindStartIndex, lexicalTokenKindDereference)
	case lexicalTokenKindLogicalOperator:
		if rawValue == expression.Not {
			legal = l.checkLastToken(lexicalTokenKindNotInitialized, lexicalTokenKindSeparator, lexicalTokenKindStartGroup, lexicalTokenKindStartParameters, lexicalTokenKindStartIndex, lexicalTokenKindLogicalOperator)
		} else {
			legal = l.checkLastToken(lexicalTokenKindEndGroup, lexicalTokenKindEndParameters, lexicalTokenKindEndIndex, lexicalTokenKindWildcard, lexicalTokenKindNull, lexicalTokenKindBoolean, lexicalTokenKindNumber, lexicalTokenKindString, lexicalTokenKindPropertyName, lexicalTokenKindNamedValue)
		}
	case lexicalTokenKindPropertyName:
		legal = l.checkLastToken(lexicalTokenKindDereference)
	case lexicalTokenKindNull, lexicalTokenKindBoolean, lexicalTokenKindNumber, lexicalTokenKindString, lexicalTokenKindFunction, lexicalTokenKindNamedValue:
		legal = l.checkLastToken(lexicalTokenKindNotInitialized, lexicalTokenKindSeparator, lexicalTokenKindStartIndex, lexicalTokenKindStartGroup, lexicalTokenKindStartParameters, lexicalTokenKindLogicalOperator)
	}
	// Illegal
	if !legal {
		return &lexicalToken{
			kind:     lexicalTokenKindUnexpected,
			rawValue: rawValue,
			index:    startIndex,
		}
	}

	// Legal so far
	token := &lexicalToken{
		kind:        kind,
		rawValue:    rawValue,
		index:       startIndex,
		parsedValue: parsedValue,
	}
	switch kind {
	case lexicalTokenKindStartGroup, lexicalTokenKindStartIndex, lexicalTokenKindStartParameters:
		// Track start token
		l.unclosedTokens = append(l.unclosedTokens, token)
	case lexicalTokenKindEndGroup:
		// Check inside logical grouping
		if l.lastUnclosedKind() != lexicalTokenKindStartGroup {
			return &lexicalToken{
				kind:     lexicalTokenKindUnexpected,
				rawValue: rawValue,
				index:    startIndex,
			}
		}
		// remove last start group
		l.unclosedTokens = l.unclosedTokens[:len(l.unclosedTokens)-1]
	case lexicalTokenKindEndIndex:
		// Check inside indexer
		if l.lastUnclosedKind() != lexicalTokenKindStartIndex {
			return &lexicalToken{
				kind:     lexicalTokenKindUnexpected,
				rawValue: rawValue,
				index:    startIndex,
			}
		}
		// remove last start startIndex
		l.unclosedTokens = l.unclosedTokens[:len(l.unclosedTokens)-1]
	case lexicalTokenKindEndParameters:
		// Check inside function call
		if l.lastUnclosedKind() != lexicalTokenKindStartParameters {
			return &lexicalToken{
				kind:     lexicalTokenKindUnexpected,
				rawValue: rawValue,
				index:    startIndex,
			}
		}
		// remove last start parameter
		l.unclosedTokens = l.unclosedTokens[:len(l.unclosedTokens)-1]
	case lexicalTokenKindSeparator: // ","
		// Check inside function call
		if l.lastUnclosedKind() != lexicalTokenKindStartParameters {
			return &lexicalToken{
				kind:     lexicalTokenKindUnexpected,
				rawValue: rawValue,
				index:    startIndex,
			}
		}
	}
	return token
}

func (l *lexer) lastUnclosedKind() (kind lexicalTokenKind) {
	if len(l.unclosedTokens) > 0 {
		kind = l.unclosedTokens[len(l.unclosedTokens)-1].kind
	}
	return
}

// checkLastToken check if current token is valid based on it precedence token's kind
func (l *lexer) checkLastToken(allowedLastKinds ...lexicalTokenKind) (valid bool) {
	var currentLastKind lexicalTokenKind
	if l.lastToken == nil {
		currentLastKind = lexicalTokenKindNotInitialized
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

func newLexer(expression string) *lexer {
	return &lexer{expression: expression}
}

// TODO: merge with evaluator.ParseNumber
func parseNumber(str string) (out float64) {
	if len(str) == 0 || len(strings.TrimSpace(str)) == 0 {
		return 0
	}
	out, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return
	}
	// hex
	if str[0] == '0' && str[1] == 'x' && len(str) > 2 {
		for i := 1; i < len(str); i++ {
			x := str[i]
			if (x >= '0' && x <= '9') || (x >= 'a' && x <= 'f') || (x >= 'A' && x <= 'F') {
				// example:
				// Convert hexadecimal string to uint64
				if intVal, err := strconv.ParseUint(str[2:], 16, 64); err == nil {
					// Convert uint64 to float64
					return float64(intVal)
				}
			}
		}
	}
	if strings.EqualFold(str, expression.Infinity) {
		return math.Inf(1)
	}
	if strings.EqualFold(str, expression.NegativeInfinity) {
		return math.Inf(0)
	}
	return math.NaN()
}
