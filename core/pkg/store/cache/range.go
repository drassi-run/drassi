/*
 * SPDX-FileCopyrightText: (c) 2024 The Drassi Authors
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

var (
	reContentRange  = regexp.MustCompile(`^(\w+) (\d+)-(\d+)/(\*|\d+)$`)
	hdrContentRange = "Content-Range"

	reRange  = regexp.MustCompile(`^(\w+)=(\d+)?-(\d+)?$`)
	hdrRange = "Range"

	ErrInvalidHeader = errors.New("invalid HTTP header")
)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Range
func parseContentRangeHeader(h http.Header) (int64, int64, error) {
	s := h.Get(hdrContentRange)
	m := reContentRange.FindStringSubmatch(s)
	if len(m) != reContentRange.NumSubexp()+1 {
		return 0, 0, fmt.Errorf("%w: %q=%q", ErrInvalidHeader, hdrContentRange, s)
	}

	var start, end int64
	if unit := m[1]; unit != "bytes" {
		return 0, 0, fmt.Errorf("%w: %q=%q: unknown units=%q", ErrInvalidHeader, hdrContentRange, s, unit)
	}

	if num, err := strconv.ParseInt(m[2], 10, 64); err != nil {
		return 0, 0, fmt.Errorf("%w: %q=%q: invalid start", ErrInvalidHeader, hdrContentRange, s)
	} else {
		start = num
	}

	if num, err := strconv.ParseInt(m[3], 10, 64); err != nil || start > num {
		return 0, 0, fmt.Errorf("%w: %q=%q: invalid end", ErrInvalidHeader, hdrContentRange, s)
	} else {
		end = num
	}

	if size := m[4]; size != "*" {
		if num, err := strconv.ParseInt(size, 10, 64); err != nil || end > num {
			return 0, 0, fmt.Errorf("%w: %q=%q: invalid size", ErrInvalidHeader, hdrContentRange, s)
		}
	}

	return start, end, nil
}

// NOTE: currently, only one range is supported
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Range
func parseRangeHeader(h http.Header) (int64, int64, error) {
	var start, end int64 = 0, -1
	s := h.Get(hdrRange)
	if s == "" {
		return start, end, nil
	}

	m := reRange.FindStringSubmatch(s)
	if len(m) != reRange.NumSubexp()+1 {
		return 0, 0, fmt.Errorf("%w: %q=%q", ErrInvalidHeader, hdrRange, s)
	}

	if unit := m[1]; unit != "bytes" {
		return 0, 0, fmt.Errorf("%w: %q=%q: unknown units=%q", ErrInvalidHeader, hdrRange, s, unit)
	}

	if m[2] != "" {
		if num, err := strconv.ParseInt(m[2], 10, 64); err != nil {
			return 0, 0, fmt.Errorf("%w: %q=%q: invalid start", ErrInvalidHeader, hdrRange, s)
		} else {
			start = num
		}
	}

	if m[3] != "" {
		if num, err := strconv.ParseInt(m[3], 10, 64); err != nil || start > num {
			return 0, 0, fmt.Errorf("%w: %q=%q: invalid end", ErrInvalidHeader, hdrRange, s)
		} else {
			end = num
		}
	}

	return start, end, nil
}
