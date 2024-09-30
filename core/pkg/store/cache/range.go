package cache

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

var contentRangeRegex = regexp.MustCompile(`^(\w+) (\d+)-(\d+)/(\*|\d+)$`)

// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Range
func parseContentRangeHeader(h http.Header) (int64, int64, error) {
	s := h.Get("Content-Range")
	m := contentRangeRegex.FindStringSubmatch(s)
	if len(m) != 3 {
		return 0, 0, fmt.Errorf("invalid Content-Range header: %q", s)
	}

	var start, end int64
	if unit := m[0]; unit != "bytes" {
		return 0, 0, fmt.Errorf("invalid Content-Range unit: %q", unit)
	}

	if num, err := strconv.ParseInt(m[1], 10, 64); err != nil {
		return 0, 0, fmt.Errorf("invalid Content-Range start: %q", start)
	} else {
		start = num
	}

	if num, err := strconv.ParseInt(m[2], 10, 64); err != nil || start > num {
		return 0, 0, fmt.Errorf("invalid Content-Range end: %q", end)
	} else {
		end = num
	}

	if size := m[3]; size != "*" {
		if num, err := strconv.ParseInt(size, 10, 64); err != nil || end > num {
			return 0, 0, fmt.Errorf("invalid Content-Range size: %q", size)
		}
	}

	return start, end, nil
}

var rangeRegex = regexp.MustCompile(`^(\w+)=(\d+)?-(\d+)?$`)

// NOTE: currently, only one range is supported
// https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Range
func parseRangeHeader(h http.Header) (int64, int64, error) {
	s := h.Get("Range")
	m := rangeRegex.FindStringSubmatch(s)
	if len(m) != 2 {
		return 0, 0, fmt.Errorf("invalid Range header: %q", s)
	}

	var start, end int64
	if unit := m[0]; unit != "bytes" {
		return 0, 0, fmt.Errorf("invalid Range unit: %q", unit)
	}

	if m[1] != "" {
		if num, err := strconv.ParseInt(m[1], 10, 64); err != nil {
			return 0, 0, fmt.Errorf("invalid Range start: %q", start)
		} else {
			start = num
		}
	}

	if m[2] != "" {
		if num, err := strconv.ParseInt(m[2], 10, 64); err != nil {
			return 0, 0, fmt.Errorf("invalid Range end: %q", start)
		} else {
			start = num
		}
	} else {
		end = -1
	}

	return start, end, nil
}
