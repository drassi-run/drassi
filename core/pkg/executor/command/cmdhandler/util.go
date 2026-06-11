/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cmdhandler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"iter"
	"strings"

	"drassi.run/core/pkg/executor/command/cmdtypes"
	"drassi.run/core/pkg/executor/runtime"
)

var ErrInvalidFile = errors.New("invalid file")

func splitLine(line string) iter.Seq[string] {
	splitter := func(c rune) bool { return c == '\n' || c == '\r' }

	return func(yield func(string) bool) {
		for _, l := range strings.FieldsFunc(line, splitter) {
			l = strings.TrimSpace(l)
			if l == "" {
				continue
			}
			if !yield(l) {
				return
			}
		}
	}
}

func readLine(reader io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		l := scanner.Text()
		if l != "" && !strings.HasPrefix(l, "#") {
			lines = append(lines, l)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// https://github.com/actions/runner/blob/v2.315.0/src/Runner.Worker/FileCommandManager.cs#L342-L403
func parseEnvVars(reader io.Reader) (map[string]string, error) {
	env := make(map[string]string)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		equalsIndex := strings.Index(line, "=")
		heredocIndex := strings.Index(line, "<<")

		// Normal style NAME=VALUE
		if 0 <= equalsIndex && (heredocIndex < 0 || equalsIndex < heredocIndex) {
			key, value := line[:equalsIndex], line[equalsIndex+1:]
			if key == "" {
				return nil, fmt.Errorf("%w: line=%q nil key", ErrInvalidFile, line)
			}
			env[key] = value
			continue
		}

		// Heredoc style NAME<<EOF
		if 0 <= heredocIndex && (equalsIndex < 0 || heredocIndex < equalsIndex) {
			key, delimiter := line[:heredocIndex], line[heredocIndex+2:]
			if key == "" || delimiter == "" {
				return nil, fmt.Errorf("%w: line=%q key and delimiter MUST NOT be empty", ErrInvalidFile, line)
			}
			value, finish := make([]string, 0), false
			for scanner.Scan() {
				doc := scanner.Text()
				if doc == delimiter {
					finish = true
					break
				}
				value = append(value, doc)
			}
			if err := scanner.Err(); err != nil {
				return nil, err
			}
			if !finish {
				return nil, fmt.Errorf("%w: EOF marker missing new line", ErrInvalidFile)
			}

			env[key] = strings.Join(value, "\n")
			continue
		}

		return nil, fmt.Errorf("%w: line=%q invalid format", ErrInvalidFile, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return env, nil
}

func getPathTranslator(res any) runtime.PathTranslator {
	if res == nil {
		return nil
	}
	if prov, ok := res.(cmdtypes.SupportPathTranslator); ok {
		return prov.PathTranslator()
	}
	return nil
}
