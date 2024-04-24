package ast

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/dungdm93/drasi/pkg/expr/common"
)

type lexer struct {
	expr     string
	pos      int
	last     *token
	unclosed []*token
}

func newLexer(expr string) *lexer {
	return &lexer{expr: expr}
}

func (l *lexer) next() (result *token, haveToken bool) {
	// skip whitespace
	l.skipWhiteSpace()
	// testcase end of string
	if l.pos >= len(l.expr) {
		return nil, false
	}
	// read the first character to determine the type of token.
	currentChar := string(l.expr[l.pos])
	switch currentChar {
	case common.StartGroup:
		// function call
		if l.last != nil && l.last.k == function {
			result = l.newToken(startParameters, currentChar, l.pos, nil)
		} else {
			// logical grouping
			result = l.newToken(startGroup, currentChar, l.pos, nil)
		}
		l.pos++
	case common.StartIndex:
		result = l.newToken(startIndex, currentChar, l.pos, nil)
		l.pos++
	case common.EndGroup:
		// function call
		if len(l.unclosed) > 0 && l.unclosed[len(l.unclosed)-1].k == startParameters {
			result = l.newToken(endParameters, currentChar, l.pos, nil)
		} else {
			// logical grouping
			result = l.newToken(endGroup, currentChar, l.pos, nil)
		}
		l.pos++
	case common.EndIndex:
		result = l.newToken(endIndex, currentChar, l.pos, nil)
		l.pos++
	case common.Separator:
		result = l.newToken(separator, currentChar, l.pos, nil)
		l.pos++
	case common.Wildcard:
		result = l.newToken(wildcard, currentChar, l.pos, nil)
		l.pos++
	case "'":
		result = l.readString()
	case "!", ">", "<", "=", "&", "|":
		// also catch != >= <= == && ||
		result = l.readOperator()
	default:
		if currentChar == "." {
			// number
			if l.last == nil || l.last.k == separator || l.last.k == startGroup || l.last.k == startIndex || l.last.k == startParameters || l.last.k == logicalOperator {
				result = l.readNumber()
			} else {
				// .
				result = l.newToken(dereference, currentChar, l.pos, nil)
				l.pos++
			}
		} else if currentChar == "-" || currentChar == "+" || (currentChar >= "0" && currentChar <= "9") {
			result = l.readNumber()
		} else {
			result = l.readKeyword()
		}
	}
	l.last = result
	return result, true
}

func (l *lexer) skipWhiteSpace() {
	for l.pos < len(l.expr) && unicode.IsSpace([]rune(l.expr)[l.pos]) {
		l.pos++
	}
}
func (l *lexer) getUnclosedTokens() []*token {
	return l.unclosed
}

func (l *lexer) readKeyword() *token {
	runes := []rune(l.expr)
	pos := l.pos
	l.pos++

	for l.pos < len(l.expr) && !testTokenBoundary(runes[l.pos]) {
		l.pos++
	}
	str := l.expr[pos:l.pos]
	if legalKeyWord(str) {
		// Test if follows property dereference operator.
		if l.last != nil && l.last.k == dereference {
			return l.newToken(propertyName, str, pos, nil)
		}
		// null
		if strings.EqualFold(str, common.Null) {
			return l.newToken(null, str, pos, nil)
		}
		// boolean
		if strings.EqualFold(str, common.True) {
			return l.newToken(boolean, str, pos, true)
		}
		if strings.EqualFold(str, common.False) {
			return l.newToken(boolean, str, pos, false)
		}
		// NaN
		if strings.EqualFold(str, common.NaN) {
			return l.newToken(number, str, pos, math.NaN())
		}
		// Infinity
		if strings.EqualFold(str, common.Infinity) {
			return l.newToken(number, str, pos, math.Inf(1))
		}
		// Lookahead
		tmpIndex := l.pos
		for tmpIndex < len(runes) && unicode.IsSpace(runes[tmpIndex]) {
			tmpIndex++
		}
		// Fn. Eg: success(), always()
		if tmpIndex < len(l.expr) && string(l.expr[tmpIndex]) == common.StartGroup {
			return l.newToken(function, str, pos, nil)
		} else {
			// Named values. Eg github
			return l.newToken(namedValue, str, pos, nil)
		}
	} else {
		return l.newToken(unexpected, str, pos, nil)
	}
}

func (l *lexer) readNumber() *token {
	pos := l.pos
	for {
		l.pos++
		if l.pos >= len(l.expr) || (testTokenBoundary(rune(l.expr[l.pos])) && l.expr[l.
			pos] != '.') {
			break
		}
	}
	str := l.expr[pos:l.pos]
	d := common.ParseNumber(str)
	if math.IsNaN(d) {
		return l.newToken(unexpected, str, pos, nil)
	}
	return l.newToken(number, str, pos, d)
}

func (l *lexer) readString() *token {
	pos := l.pos
	var closed bool
	var s strings.Builder
	l.pos++
	for l.pos < len(l.expr) {
		c := fmt.Sprintf("%c", l.expr[l.pos])
		// move to next char
		l.pos++
		if c == ("'") {
			// End of string
			if l.pos >= len(l.expr) || fmt.Sprintf("%c", l.expr[l.pos]) != "'" {
				closed = true
				break
			}
			// Escaped single quote.
			// Example: ${{ 'It''s open source!' }}
			l.pos++
		}
		_, err := s.WriteString(c)
		if err != nil {
			panic(err)
		}
	}
	rawValue := l.expr[pos:l.pos]
	if closed {
		return l.newToken(str, rawValue, pos, s.String())
	}
	return l.newToken(unexpected, rawValue, pos, nil)
}

func (l *lexer) readOperator() *token {
	pos := l.pos
	// skip first char since we already knows what it is
	l.pos++
	// check for 2 characters operator
	if l.pos < len(l.expr) {
		// increase pos, in case this is a valid 2 characters operator. We remember that this was read.
		l.pos++
		raw := l.expr[pos:l.pos]
		switch raw {
		case common.NotEqual, common.GreaterThanOrEqual, common.LessThanOrEqual, common.Equal,
			common.And, common.Or:
			return l.newToken(logicalOperator, raw, pos, nil)
		}
		l.pos--
	}

	// check for one-character operator
	raw := l.expr[pos:l.pos]
	switch raw {
	case common.Not, common.GreaterThan, common.LessThan:
		return l.newToken(logicalOperator, raw, pos, nil)
	}
	// unexpected
	for l.pos < len(l.expr) && !testTokenBoundary(rune(l.expr[l.pos])) {
		l.pos++
	}
	return l.newToken(unexpected, l.expr[pos:l.pos], pos, nil)
}

// newToken performs valid check based on last token stored in lexer, return a new token if condition check is passed
func (l *lexer) newToken(kind tokenKind, rawValue string, pos int, parsedValue any) *token {
	var legal bool
	switch kind {
	case startGroup:
		legal = l.lastTokenValid(notInitialized, separator, startGroup, startParameters, startIndex, logicalOperator)
	case startIndex:
		legal = l.lastTokenValid(endGroup, endParameters, endIndex, wildcard,
			propertyName, namedValue)
	case startParameters:
		legal = l.lastTokenValid(function)
	case endGroup:
		legal = l.lastTokenValid(endGroup, endParameters, endIndex, wildcard,
			null, boolean, number, str, propertyName, namedValue)
	case endIndex:
		legal = l.lastTokenValid(endGroup, endParameters, endIndex, wildcard, null, boolean, number, str, propertyName, namedValue)
	case endParameters:
		legal = l.lastTokenValid(startParameters, endGroup, endParameters, endIndex, wildcard, null, boolean, number, str, propertyName, namedValue)
		break
	case separator:
		legal = l.lastTokenValid(endGroup, endParameters, endIndex, wildcard, null, boolean, number, str, propertyName, namedValue)
	case dereference:
		legal = l.lastTokenValid(endGroup, endParameters, endIndex, wildcard, propertyName, namedValue)
	case wildcard:
		legal = l.lastTokenValid(startIndex, dereference)
	case logicalOperator:
		if rawValue == common.Not {
			legal = l.lastTokenValid(notInitialized, separator, startGroup, startParameters, startIndex, logicalOperator)
		} else {
			legal = l.lastTokenValid(endGroup, endParameters, endIndex, wildcard, null, boolean, number, str, propertyName, namedValue)
		}
	case propertyName:
		legal = l.lastTokenValid(dereference)
	case null, boolean, number, str, function, namedValue:
		legal = l.lastTokenValid(notInitialized, separator, startIndex, startGroup, startParameters, logicalOperator)
	}
	// Illegal
	if !legal {
		return &token{
			k:      unexpected,
			rawVal: rawValue,
			pos:    pos,
		}
	}

	// Legal so far
	tk := &token{
		k:         kind,
		rawVal:    rawValue,
		pos:       pos,
		parsedVal: parsedValue,
	}
	switch kind {
	case startGroup, startIndex, startParameters:
		// Track start tk
		l.unclosed = append(l.unclosed, tk)
	case endGroup:
		// Check inside logical grouping
		if l.lastUnclosedKind() != startGroup {
			return &token{
				k:      unexpected,
				rawVal: rawValue,
				pos:    pos,
			}
		}
		// remove last start group
		l.unclosed = l.unclosed[:len(l.unclosed)-1]
	case endIndex:
		// Check inside indexer
		if l.lastUnclosedKind() != startIndex {
			return &token{
				k:      unexpected,
				rawVal: rawValue,
				pos:    pos,
			}
		}
		// remove last start pos
		l.unclosed = l.unclosed[:len(l.unclosed)-1]
	case endParameters:
		// Check inside function call
		if l.lastUnclosedKind() != startParameters {
			return &token{
				k:      unexpected,
				rawVal: rawValue,
				pos:    pos,
			}
		}
		// remove last start parameter
		l.unclosed = l.unclosed[:len(l.unclosed)-1]
	case separator: // ","
		// Check inside function call
		if l.lastUnclosedKind() != startParameters {
			return &token{
				k:      unexpected,
				rawVal: rawValue,
				pos:    pos,
			}
		}
	}
	return tk
}

func (l *lexer) lastUnclosedKind() (kind tokenKind) {
	if len(l.unclosed) > 0 {
		kind = l.unclosed[len(l.unclosed)-1].k
	}
	return
}

// lastTokenValid check if current token is valid based on it precedence token's k
func (l *lexer) lastTokenValid(allowedLastKinds ...tokenKind) (valid bool) {
	var currentLastKind tokenKind
	if l.last == nil {
		currentLastKind = notInitialized
	} else {
		currentLastKind = l.last.k
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
