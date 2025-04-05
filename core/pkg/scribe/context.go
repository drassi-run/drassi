/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package scribe

import (
	"context"
	"strings"
)

type diaryKey struct{}

func diaryFromContext(ctx context.Context) Diary {
	if d, ok := ctx.Value(diaryKey{}).(Diary); ok {
		return d
	}
	return discard{}
}

func FromContext(ctx context.Context) *Scribe {
	diary := diaryFromContext(ctx)
	return New(ctx, diary)
}

func ContextWithScribe(ctx context.Context, s Diary) context.Context {
	return context.WithValue(ctx, diaryKey{}, s)
}

func GroupDetails(ctx context.Context, groupName string, details ...func(*Scribe)) {
	s := FromContext(ctx)
	end := s.Groupf("Run %s", groupName)
	defer end()

	for _, detail := range details {
		detail(s)
	}
}

func WithList(name string, a []string) func(*Scribe) {
	return func(s *Scribe) {
		switch l := len(a); {
		case l == 0: // does nothing
		case l <= 3:
			s.Writef("%s: [%s]", name, strings.Join(a, ", "))
		default:
			s.Writef("%s:", name)
			for _, e := range a {
				s.Writef("  - %s", e)
			}
		}
	}
}

func WithMap(name string, m map[string]string) func(*Scribe) {
	return func(s *Scribe) {
		if len(m) == 0 {
			return
		}
		s.Writef("%s:", name)
		for k, v := range m {
			s.Writef("  %s: %s", k, v)
		}
	}
}

func WithPair(key, value string) func(*Scribe) {
	return func(s *Scribe) {
		if value == "" {
			return
		}
		s.Writef("%s: %s", key, value)
	}
}
