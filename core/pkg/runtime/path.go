/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package runtime

import (
	"iter"
	"path/filepath"
	"strings"

	"drassi.run/core/util/string"
)

type PathTranslator interface {
	TranslatePath(string) (string, bool)
}

func MapPath(origin string, m iter.Seq2[string, string]) string {
	strippedOrigin := strings.TrimRight(origin, "/")
	for k, v := range m {
		if strings.TrimRight(k, "/") == strippedOrigin {
			return v
		}

		k = xstring.EnsureSuffix(k, "/")
		if inner, ok := strings.CutPrefix(origin, k); ok {
			return filepath.Join(v, inner)
		}
	}
	return ""
}
