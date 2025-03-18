/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package ast

import (
	"fmt"

	"github.com/antlr4-go/antlr/v4"
)

type syntaxError struct {
	line, column int
	message      string
}

func (e *syntaxError) Error() string {
	return fmt.Sprintf("syntax error at %d:%d: %s", e.line, e.column, e.message)
}

func NewSyntaxError(line, column int, message string) error {
	return &syntaxError{line, column, message}
}

type parseError struct {
	token antlr.Token
}

func (p *parseError) Error() string {
	return fmt.Sprintf("parse error at %d:%d: %s", p.token.GetLine(), p.token.GetColumn(), p.token.GetText())
}

func NewParseError(node antlr.ErrorNode) error {
	token := node.GetSymbol()
	return &parseError{token}
}

type tooManyErrors struct {
	errors []error
}

func (t *tooManyErrors) Error() string {
	errString := fmt.Sprintf("more than %d errors occurred", len(t.errors))
	for _, err := range t.errors {
		errString += "\n" + err.Error()
	}
	return errString
}

func NewTooManyErrors(errs []error) error {
	return &tooManyErrors{errs}
}
