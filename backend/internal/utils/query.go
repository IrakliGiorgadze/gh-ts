package utils

import (
	"net/url"
	"strconv"
)

// QueryInt safely parses an integer from query parameters.
// If missing or invalid, returns the provided default.
func QueryInt(q url.Values, key string, def int) int {
	v := q.Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
