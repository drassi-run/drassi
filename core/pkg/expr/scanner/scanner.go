package scanner

import (
	"fmt"
	"math"
	"strings"
	"unicode"

	"github.com/dungdm93/drassi/core/pkg/expr/common"
	"github.com/dungdm93/drassi/core/pkg/expr/token"
)

type Scanner struct {
	expr     string
	pos      int
	last     *token.Token
	unclosed []*token.Token
}

func NewScanner(expr string) *Scanner {
	return &Scanner{expr: expr}
}

func (s *Scanner) Next() (result *token.Token, haveToken bool) {
	// skip whitespace
	s.skipWhiteSpace()
	// testcase end of string
	if s.pos >= len(s.expr) {
		return nil, false
	}
	// read the first character to determine the type of token.
	currentChar := string(s.expr[s.pos])
	switch currentChar {
	case common.StartGroup:
		// function call
		if s.last != nil && s.last.Kind() == token.Function {
			result = s.newToken(token.StartParameters, currentChar, s.pos, nil)
		} else {
			// logical grouping
			result = s.newToken(token.StartGroup, currentChar, s.pos, nil)
		}
		s.pos++
	case common.StartIndex:
		result = s.newToken(token.StartIndex, currentChar, s.pos, nil)
		s.pos++
	case common.EndGroup:
		// function call
		if len(s.unclosed) > 0 && s.unclosed[len(s.unclosed)-1].Kind() == token.StartParameters {
			result = s.newToken(token.EndParameters, currentChar, s.pos, nil)
		} else {
			// logical grouping
			result = s.newToken(token.EndGroup, currentChar, s.pos, nil)
		}
		s.pos++
	case common.EndIndex:
		result = s.newToken(token.EndIndex, currentChar, s.pos, nil)
		s.pos++
	case common.Separator:
		result = s.newToken(token.Separator, currentChar, s.pos, nil)
		s.pos++
	case common.Wildcard:
		result = s.newToken(token.Wildcard, currentChar, s.pos, nil)
		s.pos++
	case "'":
		result = s.readString()
	case "!", ">", "<", "=", "&", "|":
		// also catch != >= <= == && ||
		result = s.readOperator()
	default:
		if currentChar == "." {
			// number
			if s.last == nil || s.last.Kind() == token.Separator || s.last.Kind() == token.StartGroup || s.last.
				Kind() == token.
				StartIndex || s.last.Kind() == token.StartParameters || s.last.Kind() == token.LogicalOperator {
				result = s.readNumber()
			} else {
				// .
				result = s.newToken(token.Dereference, currentChar, s.pos, nil)
				s.pos++
			}
		} else if currentChar == "-" || currentChar == "+" || (currentChar >= "0" && currentChar <= "9") {
			result = s.readNumber()
		} else {
			result = s.readKeyword()
		}
	}
	s.last = result
	return result, true
}

func (s *Scanner) skipWhiteSpace() {
	for s.pos < len(s.expr) && unicode.IsSpace([]rune(s.expr)[s.pos]) {
		s.pos++
	}
}
func (s *Scanner) GetUnclosedTokens() []*token.Token {
	return s.unclosed
}

func (s *Scanner) readKeyword() *token.Token {
	runes := []rune(s.expr)
	pos := s.pos
	s.pos++

	for s.pos < len(s.expr) && !testTokenBoundary(runes[s.pos]) {
		s.pos++
	}
	str := s.expr[pos:s.pos]
	if token.LegalKeyWord(str) {
		// Test if follows property dereference operator.
		if s.last != nil && s.last.Kind() == token.Dereference {
			return s.newToken(token.PropertyName, str, pos, nil)
		}
		// null
		if strings.EqualFold(str, common.Null) {
			return s.newToken(token.Null, str, pos, nil)
		}
		// boolean
		if strings.EqualFold(str, common.True) {
			return s.newToken(token.Boolean, str, pos, true)
		}
		if strings.EqualFold(str, common.False) {
			return s.newToken(token.Boolean, str, pos, false)
		}
		// NaN
		if strings.EqualFold(str, common.NaN) {
			return s.newToken(token.Number, str, pos, math.NaN())
		}
		// Infinity
		if strings.EqualFold(str, common.Infinity) {
			return s.newToken(token.Number, str, pos, math.Inf(1))
		}
		// Lookahead
		tmpIndex := s.pos
		for tmpIndex < len(runes) && unicode.IsSpace(runes[tmpIndex]) {
			tmpIndex++
		}
		// Fn. Eg: success(), always()
		if tmpIndex < len(s.expr) && string(s.expr[tmpIndex]) == common.StartGroup {
			return s.newToken(token.Function, str, pos, nil)
		} else {
			// Named values. Eg github
			return s.newToken(token.NamedValue, str, pos, nil)
		}
	} else {
		return s.newToken(token.Unexpected, str, pos, nil)
	}
}

func (s *Scanner) readNumber() *token.Token {
	pos := s.pos
	for {
		s.pos++
		if s.pos >= len(s.expr) || (testTokenBoundary(rune(s.expr[s.pos])) && s.expr[s.
			pos] != '.') {
			break
		}
	}
	str := s.expr[pos:s.pos]
	// try to parse to int first
	i, err := common.ParseInt(str)
	if err == nil {
		return s.newToken(token.Number, str, pos, int(i))
	}
	// if failed, parse float
	d := common.ParseFloat(str)
	if math.IsNaN(d) {
		return s.newToken(token.Unexpected, str, pos, nil)
	}
	return s.newToken(token.Number, str, pos, d)
}

func (s *Scanner) readString() *token.Token {
	pos := s.pos
	var closed bool
	var builder strings.Builder
	s.pos++
	for s.pos < len(s.expr) {
		c := fmt.Sprintf("%c", s.expr[s.pos])
		// move to Next char
		s.pos++
		if c == ("'") {
			// End of string
			if s.pos >= len(s.expr) || fmt.Sprintf("%c", s.expr[s.pos]) != "'" {
				closed = true
				break
			}
			// Escaped single quote.
			// Example: ${{ 'It''s open source!' }}
			s.pos++
		}
		_, err := builder.WriteString(c)
		if err != nil {
			panic(err)
		}
	}
	rawValue := s.expr[pos:s.pos]
	if closed {
		return s.newToken(token.Str, rawValue, pos, builder.String())
	}
	return s.newToken(token.Unexpected, rawValue, pos, nil)
}

func (s *Scanner) readOperator() *token.Token {
	pos := s.pos
	// skip first char since we already knows what it is
	s.pos++
	// check for 2 characters operator
	if s.pos < len(s.expr) {
		// increase pos, in case this is a valid 2 characters operator. We remember that this was read.
		s.pos++
		raw := s.expr[pos:s.pos]
		switch raw {
		case common.NotEqual, common.GreaterThanOrEqual, common.LessThanOrEqual, common.Equal,
			common.And, common.Or:
			return s.newToken(token.LogicalOperator, raw, pos, nil)
		}
		s.pos--
	}

	// check for one-character operator
	raw := s.expr[pos:s.pos]
	switch raw {
	case common.Not, common.GreaterThan, common.LessThan:
		return s.newToken(token.LogicalOperator, raw, pos, nil)
	}
	// unexpected
	for s.pos < len(s.expr) && !testTokenBoundary(rune(s.expr[s.pos])) {
		s.pos++
	}
	return s.newToken(token.Unexpected, s.expr[pos:s.pos], pos, nil)
}

// newToken performs valid check based on last token stored in Scanner, return a new token if condition check is passed
func (s *Scanner) newToken(kind token.Kind, rawValue string, pos int, parsedValue any) *token.Token {
	var legal bool
	switch kind {
	case token.StartGroup:
		legal = s.lastTokenValid(token.NotInitialized, token.Separator, token.StartGroup, token.StartParameters, token.StartIndex, token.LogicalOperator)
	case token.StartIndex:
		legal = s.lastTokenValid(token.EndGroup, token.EndParameters, token.EndIndex, token.Wildcard,
			token.PropertyName, token.NamedValue)
	case token.StartParameters:
		legal = s.lastTokenValid(token.Function)
	case token.EndGroup:
		legal = s.lastTokenValid(token.EndGroup, token.EndParameters, token.EndIndex, token.Wildcard,
			token.Null, token.Boolean, token.Number, token.Str, token.PropertyName, token.NamedValue)
	case token.EndIndex:
		legal = s.lastTokenValid(token.EndGroup, token.EndParameters, token.EndIndex, token.Wildcard, token.Null, token.Boolean, token.Number, token.Str, token.PropertyName, token.NamedValue)
	case token.EndParameters:
		legal = s.lastTokenValid(token.StartParameters, token.EndGroup, token.EndParameters, token.EndIndex, token.Wildcard, token.Null, token.Boolean, token.Number, token.Str, token.PropertyName, token.NamedValue)
		break
	case token.Separator:
		legal = s.lastTokenValid(token.EndGroup, token.EndParameters, token.EndIndex, token.Wildcard, token.Null, token.Boolean, token.Number, token.Str, token.PropertyName, token.NamedValue)
	case token.Dereference:
		legal = s.lastTokenValid(token.EndGroup, token.EndParameters, token.EndIndex, token.Wildcard, token.PropertyName, token.NamedValue)
	case token.Wildcard:
		legal = s.lastTokenValid(token.StartIndex, token.Dereference)
	case token.LogicalOperator:
		if rawValue == common.Not {
			legal = s.lastTokenValid(token.NotInitialized, token.Separator, token.StartGroup, token.StartParameters, token.StartIndex, token.LogicalOperator)
		} else {
			legal = s.lastTokenValid(token.EndGroup, token.EndParameters, token.EndIndex, token.Wildcard, token.Null, token.Boolean, token.Number, token.Str, token.PropertyName, token.NamedValue)
		}
	case token.PropertyName:
		legal = s.lastTokenValid(token.Dereference)
	case token.Null, token.Boolean, token.Number, token.Str, token.Function, token.NamedValue:
		legal = s.lastTokenValid(token.NotInitialized, token.Separator, token.StartIndex, token.StartGroup, token.StartParameters, token.LogicalOperator)
	}
	// Illegal
	if !legal {
		return &token.Token{
			K:      token.Unexpected,
			RawVal: rawValue,
			Pos:    pos,
		}
	}

	// Legal so far
	tk := &token.Token{
		K:         kind,
		RawVal:    rawValue,
		Pos:       pos,
		ParsedVal: parsedValue,
	}
	switch kind {
	case token.StartGroup, token.StartIndex, token.StartParameters:
		// Track start tk
		s.unclosed = append(s.unclosed, tk)
	case token.EndGroup:
		// Check inside logical grouping
		if s.lastUnclosedKind() != token.StartGroup {
			return &token.Token{
				K:      token.Unexpected,
				RawVal: rawValue,
				Pos:    pos,
			}
		}
		// remove last start group
		s.unclosed = s.unclosed[:len(s.unclosed)-1]
	case token.EndIndex:
		// Check inside indexer
		if s.lastUnclosedKind() != token.StartIndex {
			return &token.Token{
				K:      token.Unexpected,
				RawVal: rawValue,
				Pos:    pos,
			}
		}
		// remove last start pos
		s.unclosed = s.unclosed[:len(s.unclosed)-1]
	case token.EndParameters:
		// Check inside function call
		if s.lastUnclosedKind() != token.StartParameters {
			return &token.Token{
				K:      token.Unexpected,
				RawVal: rawValue,
				Pos:    pos,
			}
		}
		// remove last start parameter
		s.unclosed = s.unclosed[:len(s.unclosed)-1]
	case token.Separator: // ","
		// Check inside function call
		if s.lastUnclosedKind() != token.StartParameters {
			return &token.Token{
				K:      token.Unexpected,
				RawVal: rawValue,
				Pos:    pos,
			}
		}
	}
	return tk
}

func (s *Scanner) lastUnclosedKind() (kind token.Kind) {
	if len(s.unclosed) > 0 {
		kind = s.unclosed[len(s.unclosed)-1].Kind()
	}
	return
}

// lastTokenValid check if current token is valid based on it precedence token's k
func (s *Scanner) lastTokenValid(allowedLastKinds ...token.Kind) (valid bool) {
	var currentLastKind token.Kind
	if s.last == nil {
		currentLastKind = token.NotInitialized
	} else {
		currentLastKind = s.last.Kind()
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
